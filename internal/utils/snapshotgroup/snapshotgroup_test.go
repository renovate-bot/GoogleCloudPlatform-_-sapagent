/*
Copyright 2025 Google LLC

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

package snapshotgroup

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/compute/v1"
	"golang.org/x/oauth2"
)

type (
	mockHTTPClient struct {
		// responses is a map of method -> URL -> slice of responses.
		// Each call to Do will consume one response from the slice.
		responses map[string]map[string][]httpResponse
		// callCount tracks the number of calls for each URL and method.
		callCount map[string]int
	}
	httpResponse struct {
		statusCode int
		body       string
	}
)

func (c *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if c.callCount == nil {
		c.callCount = make(map[string]int)
	}
	url := req.URL.String()
	method := req.Method
	key := fmt.Sprintf("%s %s", method, url)

	methodResponses, ok := c.responses[method]
	if !ok {
		return nil, fmt.Errorf("unexpected verb: %s for URL: %s", method, url)
	}

	responseList, ok := methodResponses[url]
	if !ok || len(responseList) == 0 {
		return nil, fmt.Errorf("unexpected URL: %s for verb: %s", url, method)
	}

	callIndex := c.callCount[key]
	c.callCount[key]++

	// If we've exhausted the prepared responses, just use the last one.
	// This handles polling after a final state is reached.
	if callIndex >= len(responseList) {
		callIndex = len(responseList) - 1
	}

	resp := responseList[callIndex]
	return newResponse(resp.statusCode, resp.body), nil
}

func newResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// mockTokenSource is a mock implementation of oauth2.TokenSource.
type mockTokenSource struct {
	AccessToken string
	Err         error
}

func (mts *mockTokenSource) Token() (*oauth2.Token, error) {
	if mts.Err != nil {
		return nil, mts.Err
	}
	return &oauth2.Token{AccessToken: mts.AccessToken}, nil
}

func defaultTokenGetterMock(err error) func(context.Context, ...string) (oauth2.TokenSource, error) {
	return func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
		if err != nil {
			return nil, err
		}
		return &mockTokenSource{AccessToken: "test-token"}, nil
	}
}

func TestNewService(t *testing.T) {
	s := &SGService{}
	err := s.NewService()
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}
	if s.rest == nil {
		t.Error("NewService() did not initialize rest client")
	}
	if s.backoff == nil {
		t.Error("NewService() did not initialize backoff")
	}
	if s.maxRetries == 0 {
		t.Error("NewService() did not initialize maxRetries")
	}
}

func TestSetupBackoff(t *testing.T) {
	var b backoff.BackOff
	b = &backoff.ExponentialBackOff{
		InitialInterval:     2 * time.Second,
		RandomizationFactor: 0,
		Multiplier:          2,
		MaxInterval:         1 * time.Hour,
		MaxElapsedTime:      30 * time.Minute,
		Clock:               backoff.SystemClock,
	}
	gotB := setupBackoff()
	if diff := cmp.Diff(b, gotB, cmpopts.IgnoreUnexported(backoff.ExponentialBackOff{})); diff != "" {
		t.Errorf("setupBackoff() returned diff (-want +got):\n%s", diff)
	}
}

func TestGetResponse(t *testing.T) {
	tests := []struct {
		name           string
		httpResponses  map[string]map[string][]httpResponse
		method         string
		url            string
		data           []byte
		expectedBody   string
		expectedError  bool
		tokenGetterErr error
	}{
		{
			name: "success",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://test.com/success": {{statusCode: 200, body: `{"key":"value"}`}}},
			},
			method:       "GET",
			url:          "https://test.com/success",
			expectedBody: `{"key":"value"}`,
		},
		{
			name: "http_error",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://test.com/error": {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			method:        "GET",
			url:           "https://test.com/error",
			expectedError: true, // GetResponse now returns the googleapi.Error as error
		},
		{
			name: "unmarshal_error_generic_response",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://test.com/unmarshal_error": {{statusCode: 200, body: `invalid_json`}}},
			},
			method:        "GET",
			url:           "https://test.com/unmarshal_error",
			expectedError: true,
		},
		{
			name: "unmarshal_error_googleapi_error",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://test.com/unmarshal_google_error": {{statusCode: 400, body: `{"error": "not_an_object"}`}}},
			},
			method:        "GET",
			url:           "https://test.com/unmarshal_google_error",
			expectedError: true,
		},
		{
			name:           "token_getter_error",
			httpResponses:  map[string]map[string][]httpResponse{},
			method:         "GET",
			url:            "https://test.com/any",
			expectedError:  true,
			tokenGetterErr: fmt.Errorf("token error"),
		},
		{
			name: "op_with_error_success",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://test.com/op_with_error": {{statusCode: 200, body: `{"name":"operation-123", "error": {"errors": [{"message": "failed to bulk insert"}]}}`}}},
			},
			method:        "POST",
			url:           "https://test.com/op_with_error",
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService() // Initialize with default backoff and retries
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(test.tokenGetterErr)

			body, err := sgService.GetResponse(ctx, test.method, test.url, test.data)

			if (err != nil) != test.expectedError {
				t.Errorf("GetResponse() error = %v, wantErr %v", err, test.expectedError)
				return
			}
			if err == nil && string(body) != test.expectedBody {
				t.Errorf("GetResponse() body = %s, want %s", string(body), test.expectedBody)
			}
		})
	}
}

func TestBulkInsertFromSG(t *testing.T) {
	tests := []struct {
		name              string
		project           string
		zone              string
		data              []byte
		httpResponses     map[string]map[string][]httpResponse
		expectedOperation *compute.Operation
		expectedError     bool
		tokenGetterErr    error
	}{
		{
			name:    "success",
			project: "test-project",
			zone:    "test-zone",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks/bulkInsert": {{statusCode: 200, body: `{"name":"operation-123", "status":"RUNNING"}`}}},
			},
			expectedOperation: &compute.Operation{
				Name:   "operation-123",
				Status: "RUNNING",
			},
		},
		{
			name:    "http_error",
			project: "test-project",
			zone:    "test-zone",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks/bulkInsert": {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError: true,
		},
		{
			name:    "api_error",
			project: "test-project",
			zone:    "test-zone",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks/bulkInsert": {{statusCode: 200, body: `{"name":"operation-123", "error": {"errors": [{"message": "failed to bulk insert"}]}}`}}},
			},
			expectedError: true,
		},
		{
			name:    "api_error_with_http_code",
			project: "test-project",
			zone:    "test-zone",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks/bulkInsert": {{statusCode: 200, body: `{"name":"operation-123", "error": {"errors": []}, "httpErrorStatusCode": 400, "httpErrorMessage": "bad request"}`}}},
			},
			expectedError: true,
		},
		{
			name:    "api_error_no_details",
			project: "test-project",
			zone:    "test-zone",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks/bulkInsert": {{statusCode: 200, body: `{"name":"operation-123", "error": {"errors": []}}`}}},
			},
			expectedError: true,
		},
		{
			name:    "unmarshal_error",
			project: "test-project",
			zone:    "test-zone",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks/bulkInsert": {{statusCode: 200, body: `invalid_json`}}},
			},
			expectedError: true,
		},
		{
			name:           "token_getter_error",
			project:        "test-project",
			zone:           "test-zone",
			httpResponses:  map[string]map[string][]httpResponse{},
			expectedError:  true,
			tokenGetterErr: fmt.Errorf("token error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.maxRetries = 1
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(test.tokenGetterErr)

			op, err := sgService.BulkInsertFromSG(ctx, test.project, test.zone, test.data)

			if (err != nil) != test.expectedError {
				t.Errorf("BulkInsertFromSG() error = %v, wantErr %v", err, test.expectedError)
			}
			if diff := cmp.Diff(test.expectedOperation, op); diff != "" && !test.expectedError {
				t.Errorf("BulkInsertFromSG() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWaitForSGUploadCompletion(t *testing.T) {
	tests := []struct {
		name           string
		project        string
		sgName         string
		httpResponses  map[string]map[string][]httpResponse
		expectedError  bool
		tokenGetterErr error
	}{
		{
			name:    "success",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 200, body: `{"name":"test-sg", "status":"READY"}`}}},
			},
		},
		{
			name:    "uploading_error",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 200, body: `{"name":"test-sg", "status":"UPLOADING"}`}}},
			},
			expectedError: true,
		},
		{
			name:    "get_sg_error",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError: true,
		},
		{
			name:           "token_getter_error",
			project:        "test-project",
			sgName:         "test-sg",
			httpResponses:  map[string]map[string][]httpResponse{},
			expectedError:  true,
			tokenGetterErr: fmt.Errorf("token error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(test.tokenGetterErr)

			err := sgService.WaitForSGUploadCompletion(ctx, test.project, test.sgName)

			if (err != nil) != test.expectedError {
				t.Errorf("WaitForSGUploadCompletion() error = %v, wantErr %v", err, test.expectedError)
			}
		})
	}
}

func TestGetSG(t *testing.T) {
	tests := []struct {
		name           string
		project        string
		sgName         string
		httpResponses  map[string]map[string][]httpResponse
		expectedSGItem *SGItem
		expectedError  bool
		tokenGetterErr error
	}{
		{
			name:    "success",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 200, body: `{"name":"test-sg"}`}}},
			},
			expectedSGItem: &SGItem{
				Name: "test-sg",
			},
		},
		{
			name:    "http_error",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError: true,
		},
		{
			name:    "unmarshal_error",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 200, body: `invalid_json`}}},
			},
			expectedError: true,
		},
		{
			name:           "token_getter_error",
			project:        "test-project",
			sgName:         "test-sg",
			httpResponses:  map[string]map[string][]httpResponse{},
			expectedError:  true,
			tokenGetterErr: fmt.Errorf("token error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(test.tokenGetterErr)

			sgItem, err := sgService.GetSG(ctx, test.project, test.sgName)

			if (err != nil) != test.expectedError {
				t.Errorf("GetSG() error = %v, wantErr %v", err, test.expectedError)
				return
			}
			if diff := cmp.Diff(test.expectedSGItem, sgItem, cmpopts.IgnoreUnexported(SGItem{})); diff != "" && !test.expectedError {
				t.Errorf("GetSG() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListSGs(t *testing.T) {
	tests := []struct {
		name            string
		project         string
		httpResponses   map[string]map[string][]httpResponse
		expectedSGItems []SGItem
		expectedError   bool
		tokenGetterErr  error
	}{
		{
			name:    "success",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups": {{statusCode: 200, body: `{"items":[{"name":"test-sg1"},{"name":"test-sg2"}]}`}}},
			},
			expectedSGItems: []SGItem{
				{Name: "test-sg1"},
				{Name: "test-sg2"},
			},
		},
		{
			name:    "success_with_pagination",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {
					"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups":                           {{statusCode: 200, body: `{"items":[{"name":"test-sg1"}], "nextPageToken":"next-page-token"}`}},
					"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups?pageToken=next-page-token": {{statusCode: 200, body: `{"items":[{"name":"test-sg2"}]}`}},
				},
			},
			expectedSGItems: []SGItem{
				{Name: "test-sg1"},
				{Name: "test-sg2"},
			},
		},
		{
			name:    "http_error",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups": {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError: true,
		},
		{
			name:    "unmarshal_error",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups": {{statusCode: 200, body: `invalid_json`}}},
			},
			expectedError: true,
		},
		{
			name:           "token_getter_error",
			project:        "test-project",
			httpResponses:  map[string]map[string][]httpResponse{},
			expectedError:  true,
			tokenGetterErr: fmt.Errorf("token error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(test.tokenGetterErr)

			sgItems, err := sgService.ListSGs(ctx, test.project)

			if (err != nil) != test.expectedError {
				t.Errorf("ListSGs() error = %v, wantErr %v", err, test.expectedError)
				return
			}
			if diff := cmp.Diff(test.expectedSGItems, sgItems, cmpopts.IgnoreUnexported(SGItem{})); diff != "" && !test.expectedError {
				t.Errorf("ListSGs() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCreateSG(t *testing.T) {
	tests := []struct {
		name            string
		s               *SGService
		project         string
		data            []byte
		httpResponses   map[string]map[string][]httpResponse
		expectedError   bool
		expectedBaseURL string
	}{
		// TODO: Add a success test.
		{
			name:    "error_http",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups": {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError:   true,
			expectedBaseURL: "https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups",
		},
		{
			name:    "error_unmarshal_operation",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups": {{statusCode: 200, body: `invalid_json`}}},
			},
			expectedError:   true,
			expectedBaseURL: "https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups",
		},
		{
			name:    "error_operation_error",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups": {{statusCode: 200, body: `{"name":"operation-123", "error": {"errors": [{"message": "failed to create"}]}}`}}},
			},
			expectedError:   true,
			expectedBaseURL: "https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups",
		},
		{
			name:    "error_operation_error_http_code",
			project: "test-project",
			httpResponses: map[string]map[string][]httpResponse{
				"POST": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups": {{statusCode: 200, body: `{"name":"operation-123", "error": {"errors": []}, "httpErrorStatusCode": 400, "httpErrorMessage": "bad request"}`}}},
			},
			expectedError:   true,
			expectedBaseURL: "https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups",
		},
	}

	for _, test := range tests {
		test.data = []byte(`{"name": "test-sg", "sourceInstantSnapshotGroup":"projects/test-project/zones/us-central1-a/instantSnapshotGroups/test-isg"}`)
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.maxRetries = 1
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(nil)

			err := sgService.CreateSG(ctx, test.project, test.data)

			if (err != nil) != test.expectedError {
				t.Errorf("CreateSG() error = %v, wantErr %v", err, test.expectedError)
			}
			if sgService.baseURL != test.expectedBaseURL {
				t.Errorf("CreateSG() baseURL = %s, want %s", sgService.baseURL, test.expectedBaseURL)
			}
		})
	}
}

func TestListSnapshotsFromSG(t *testing.T) {
	tests := []struct {
		name                  string
		project               string
		sgName                string
		httpResponses         map[string]map[string][]httpResponse
		expectedSnapshotItems []SnapshotItem
		expectedError         bool
		tokenGetterErr        error
	}{
		{
			name:    "success",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshots?filter=%s", url.QueryEscape(`snapshotGroupName="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg"`)): {{statusCode: 200, body: `{"items":[{"name":"test-snapshot1"},{"name":"test-snapshot2"}]}`}}},
			},
			expectedSnapshotItems: []SnapshotItem{
				{Name: "test-snapshot1"},
				{Name: "test-snapshot2"},
			},
		},
		{
			name:    "success_with_pagination",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {
					fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshots?filter=%s", url.QueryEscape(`snapshotGroupName="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg"`)):                           {{statusCode: 200, body: `{"items":[{"name":"test-snapshot1"}], "nextPageToken":"next-page-token"}`}},
					fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshots?filter=%s&pageToken=next-page-token", url.QueryEscape(`snapshotGroupName="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg"`)): {{statusCode: 200, body: `{"items":[{"name":"test-snapshot2"}]}`}},
				},
			},
			expectedSnapshotItems: []SnapshotItem{
				{Name: "test-snapshot1"},
				{Name: "test-snapshot2"},
			},
		},
		{
			name:    "http_error",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshots?filter=%s", url.QueryEscape(`snapshotGroupName="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg"`)): {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError: true,
		},
		{
			name:    "unmarshal_error",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshots?filter=%s", url.QueryEscape(`snapshotGroupName="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg"`)): {{statusCode: 200, body: `invalid_json`}}},
			},
			expectedError: true,
		},
		{
			name:           "token_getter_error",
			project:        "test-project",
			sgName:         "test-sg",
			httpResponses:  map[string]map[string][]httpResponse{},
			expectedError:  true,
			tokenGetterErr: fmt.Errorf("token error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(test.tokenGetterErr)

			snapshotItems, err := sgService.ListSnapshotsFromSG(ctx, test.project, test.sgName)

			if (err != nil) != test.expectedError {
				t.Errorf("ListSnapshotsFromSG() error = %v, wantErr %v", err, test.expectedError)
				return
			}
			if diff := cmp.Diff(test.expectedSnapshotItems, snapshotItems, cmpopts.IgnoreUnexported(SnapshotItem{})); diff != "" && !test.expectedError {
				t.Errorf("ListSnapshotsFromSG() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListDisksFromSnapshot(t *testing.T) {
	tests := []struct {
		name              string
		project           string
		zone              string
		snapshotName      string
		httpResponses     map[string]map[string][]httpResponse
		expectedDiskItems []DiskItem
		expectedError     bool
		tokenGetterErr    error
	}{
		{
			name:         "success",
			project:      "test-project",
			zone:         "test-zone",
			snapshotName: "test-snapshot",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks?filter=%s", url.QueryEscape(`sourceSnapshot="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshots/test-snapshot"`)): {{statusCode: 200, body: `{"items":[{"name":"test-disk1"},{"name":"test-disk2"}]}`}}},
			},
			expectedDiskItems: []DiskItem{
				{Name: "test-disk1"},
				{Name: "test-disk2"},
			},
		},
		{
			name:         "success_with_pagination",
			project:      "test-project",
			zone:         "test-zone",
			snapshotName: "test-snapshot",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {
					fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks?filter=%s", url.QueryEscape(`sourceSnapshot="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshots/test-snapshot"`)):                           {{statusCode: 200, body: `{"items":[{"name":"test-disk1"}], "nextPageToken":"next-page-token"}`}},
					fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks?filter=%s&pageToken=next-page-token", url.QueryEscape(`sourceSnapshot="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshots/test-snapshot"`)): {{statusCode: 200, body: `{"items":[{"name":"test-disk2"}]}`}},
				},
			},
			expectedDiskItems: []DiskItem{
				{Name: "test-disk1"},
				{Name: "test-disk2"},
			},
		},
		{
			name:         "http_error",
			project:      "test-project",
			zone:         "test-zone",
			snapshotName: "test-snapshot",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks?filter=%s", url.QueryEscape(`sourceSnapshot="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshots/test-snapshot"`)): {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError: true,
		},
		{
			name:         "unmarshal_error",
			project:      "test-project",
			zone:         "test-zone",
			snapshotName: "test-snapshot",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {fmt.Sprintf("https://compute.googleapis.com/compute/alpha/projects/test-project/zones/test-zone/disks?filter=%s", url.QueryEscape(`sourceSnapshot="https://www.googleapis.com/compute/alpha/projects/test-project/global/snapshots/test-snapshot"`)): {{statusCode: 200, body: `invalid_json`}}},
			},
			expectedError: true,
		},
		{
			name:           "token_getter_error",
			project:        "test-project",
			zone:           "test-zone",
			snapshotName:   "test-snapshot",
			httpResponses:  map[string]map[string][]httpResponse{},
			expectedError:  true,
			tokenGetterErr: fmt.Errorf("token error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(test.tokenGetterErr)

			diskItems, err := sgService.ListDisksFromSnapshot(ctx, test.project, test.zone, test.snapshotName)

			if (err != nil) != test.expectedError {
				t.Errorf("ListDisksFromSnapshot() error = %v, wantErr %v", err, test.expectedError)
				return
			}
			if diff := cmp.Diff(test.expectedDiskItems, diskItems, cmpopts.IgnoreUnexported(DiskItem{})); diff != "" && !test.expectedError {
				t.Errorf("ListDisksFromSnapshot() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWaitForSGCreation(t *testing.T) {
	tests := []struct {
		name          string
		project       string
		sgName        string
		httpResponses map[string]map[string][]httpResponse
		expectedError bool
	}{
		{
			name:    "Success",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 200, body: `{"name":"test-sg", "status":"READY"}`}}},
			},
			expectedError: false,
		},
		{
			name:    "Creating",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 200, body: `{"name":"test-sg", "status":"CREATING"}`}}},
			},
			expectedError: true,
		},
		{
			name:    "GetSGError",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {{statusCode: 500, body: `{"error":{"code":500,"message":"server error"}}`}}},
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(nil)

			err := sgService.WaitForSGCreation(ctx, test.project, test.sgName)

			if (err != nil) != test.expectedError {
				t.Errorf("WaitForSGCreation() error = %v, wantErr %v", err, test.expectedError)
			}
		})
	}
}

func TestWaitForSGCreationWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		project       string
		sgName        string
		httpResponses map[string]map[string][]httpResponse
		maxRetries    int
		expectedError bool
	}{
		{
			name:    "SuccessAfterRetry",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {
					{statusCode: 200, body: `{"name":"test-sg", "status":"CREATING"}`},
					{statusCode: 200, body: `{"name":"test-sg", "status":"READY"}`},
				}},
			},
			maxRetries: 3,
		},
		{
			name:    "RetryExhausted",
			project: "test-project",
			sgName:  "test-sg",
			httpResponses: map[string]map[string][]httpResponse{
				"GET": {"https://compute.googleapis.com/compute/alpha/projects/test-project/global/snapshotGroups/test-sg": {
					{statusCode: 200, body: `{"name":"test-sg", "status":"CREATING"}`},
					{statusCode: 200, body: `{"name":"test-sg", "status":"CREATING"}`},
					{statusCode: 200, body: `{"name":"test-sg", "status":"CREATING"}`},
				}},
			},
			maxRetries:    3,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sgService := &SGService{}
			sgService.NewService()
			sgService.rest.HTTPClient = &mockHTTPClient{responses: test.httpResponses}
			sgService.rest.TokenGetter = defaultTokenGetterMock(nil)
			sgService.maxRetries = test.maxRetries
			// Use a fast backoff for testing
			sgService.backoff.InitialInterval = 1 * time.Millisecond
			sgService.backoff.MaxElapsedTime = 100 * time.Millisecond

			err := sgService.WaitForSGCreationWithRetry(ctx, test.project, test.sgName)

			if (err != nil) != test.expectedError {
				t.Errorf("WaitForSGCreationWithRetry() error = %v, wantErr %v", err, test.expectedError)
			}
		})
	}
}
