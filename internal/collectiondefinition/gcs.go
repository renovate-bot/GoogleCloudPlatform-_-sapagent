/*
Copyright 2023 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collectiondefinition

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/sapagent/internal/commandlineexecutor"
	"github.com/GoogleCloudPlatform/sapagent/internal/storage"
	"github.com/GoogleCloudPlatform/sapagent/shared/log"

	cdpb "github.com/GoogleCloudPlatform/sapagent/protos/collectiondefinition"
	cpb "github.com/GoogleCloudPlatform/sapagent/protos/configuration"
)

var (
	bucketEnvMap = map[cpb.TargetEnvironment]string{
		cpb.TargetEnvironment_PRODUCTION:  "cloudsapdeploy",
		cpb.TargetEnvironment_STAGING:     "cloudsapdeploytesting",
		cpb.TargetEnvironment_DEVELOPMENT: "sapagent-collection-definition-dev",
	}
	//go:embed public.pem
	pubkey []byte
)

// createTempFunc is a wrapper around os.CreateTemp.
type createTempFunc func(dir, pattern string) (*os.File, error)

// fetchOptions describes the required arguments for calling fetchFromGCS.
type fetchOptions struct {
	// operating system value as reported by runtime.GOOS
	osType string

	// TargetEnvironment from configuration
	env cpb.TargetEnvironment

	// function for storage client creation
	client storage.Client

	// function for temporary file creation
	createTemp createTempFunc

	// command line execution function
	execute commandlineexecutor.Execute
}

// fetchFromGCS retrieves and authenticates a collection definition from GCS.
func fetchFromGCS(ctx context.Context, opts fetchOptions) *cdpb.CollectionDefinition {
	bucketName := bucketEnvMap[opts.env]
	bh, ok := storage.ConnectToBucket(ctx, opts.client, "", bucketName, "", true)
	if !ok {
		return nil
	}

	// Download the collection definition JSON file.
	// From the Google Cloud Storage docs:
	// Typically, a Reader computes the CRC of the downloaded content
	// and compares it to the stored CRC, returning an error from Read
	// if there is a mismatch.
	cdJSON, err := opts.createTemp("", "collection-definition.*.json")
	if err != nil {
		log.CtxLogger(ctx).Warnw("Could not create temporary JSON file", "error", err)
		return nil
	}
	defer os.Remove(cdJSON.Name())
	defer cdJSON.Close()
	rw1 := storage.ReadWriter{
		Writer:       cdJSON,
		BucketHandle: bh,
		Copier:       io.Copy,
		BucketName:   bucketName,
		ObjectName:   "sapagent/collection-definition/collection_definition.json",
		MaxRetries:   1,
	}
	if _, err = rw1.Download(ctx); err != nil {
		log.CtxLogger(ctx).Warnw("Could not download from cloud storage", "objectName", rw1.ObjectName, "error", err)
		return nil
	}

	// Download the signature file.
	cdSignature, err := opts.createTemp("", "collection-definition.*.signature")
	if err != nil {
		log.CtxLogger(ctx).Warnw("Could not create temporary signature file", "error", err)
		return nil
	}
	defer os.Remove(cdSignature.Name())
	defer cdSignature.Close()
	rw2 := storage.ReadWriter{
		Writer:       cdSignature,
		BucketHandle: bh,
		Copier:       io.Copy,
		BucketName:   bucketName,
		ObjectName:   "sapagent/collection-definition/collection_definition.signature",
		MaxRetries:   1,
	}
	if _, err = rw2.Download(ctx); err != nil {
		log.CtxLogger(ctx).Warnw("Could not download from cloud storage", "objectName", rw2.ObjectName, "error", err)
		return nil
	}

	// Verify using the public key embedded in the agent.
	cdPub, err := opts.createTemp("", "public.*.pem")
	if err != nil {
		log.CtxLogger(ctx).Warnw("Could not create temporary public key file", "error", err)
		return nil
	}
	if _, err = cdPub.Write(pubkey); err != nil {
		log.CtxLogger(ctx).Warnw("Could not write public key to temp file", "error", err)
		return nil
	}
	defer os.Remove(cdPub.Name())
	defer cdPub.Close()
	if opts.osType == "windows" {
		// TODO: Windows verification of downloaded content
		log.CtxLogger(ctx).Warn("Windows signature verification is not supported")
		return nil
	}
	result := opts.execute(ctx, commandlineexecutor.Params{
		Executable:  "openssl",
		ArgsToSplit: fmt.Sprintf("dgst -sha256 -verify %s -signature %s %s", cdPub.Name(), cdSignature.Name(), cdJSON.Name()),
	})
	if strings.TrimSpace(result.StdOut) != "Verified OK" {
		log.CtxLogger(ctx).Warnw("Could not verify downloaded content for collection definition", "stdout", result.StdOut, "stderr", result.StdErr)
		return nil
	}

	// Read and unmarshal the file as a collection definition.
	stat, err := cdJSON.Stat()
	if err != nil {
		log.CtxLogger(ctx).Warn(err)
		return nil
	}
	b := make([]byte, stat.Size())
	if _, err = cdJSON.ReadAt(b, 0); err != nil && err != io.EOF {
		log.CtxLogger(ctx).Warn(err)
		return nil
	}
	cd, err := unmarshal(b)
	if err != nil {
		log.CtxLogger(ctx).Warn(err)
		return nil
	}
	return cd
}