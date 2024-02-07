/*
Copyright 2022 Google LLC

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

// Package system contains types and functions needed to perform SAP System discovery operations.
package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	backoff "github.com/cenkalti/backoff/v4"
	logging "cloud.google.com/go/logging"
	"golang.org/x/exp/slices"
	workloadmanager "google.golang.org/api/workloadmanager/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"github.com/GoogleCloudPlatform/sapagent/internal/recovery"
	"github.com/GoogleCloudPlatform/sapagent/internal/system/appsdiscovery"
	"github.com/GoogleCloudPlatform/sapagent/internal/usagemetrics"
	cpb "github.com/GoogleCloudPlatform/sapagent/protos/configuration"
	ipb "github.com/GoogleCloudPlatform/sapagent/protos/instanceinfo"
	sappb "github.com/GoogleCloudPlatform/sapagent/protos/sapapp"
	spb "github.com/GoogleCloudPlatform/sapagent/protos/system"
	"github.com/GoogleCloudPlatform/sapagent/shared/log"
)

// Discovery is a type used to perform SAP System discovery operations.
type Discovery struct {
	WlmService              wlmInterface
	CloudLogInterface       cloudLogInterface
	CloudDiscoveryInterface cloudDiscoveryInterface
	HostDiscoveryInterface  hostDiscoveryInterface
	SapDiscoveryInterface   sapDiscoveryInterface
	AppsDiscovery           func(context.Context) *sappb.SAPInstances
	systems                 []*spb.SapDiscovery
	systemMu                sync.Mutex
	sapInstances            *sappb.SAPInstances
	sapMu                   sync.Mutex
	sapInstancesRoutine     *recovery.RecoverableRoutine
	systemDiscoveryRoutine  *recovery.RecoverableRoutine
}

// GetSAPSystems returns the current list of SAP Systems discovered on the current host.
func (d *Discovery) GetSAPSystems() []*spb.SapDiscovery {
	d.systemMu.Lock()
	defer d.systemMu.Unlock()
	return d.systems
}

// GetSAPInstances returns the current list of SAP Instances discovered on the current host.
func (d *Discovery) GetSAPInstances() *sappb.SAPInstances {
	d.sapMu.Lock()
	defer d.sapMu.Unlock()
	return d.sapInstances
}

// StartSAPSystemDiscovery Initializes the discovery object and starts the discovery subroutine.
// Returns true if the discovery goroutine is started, and false otherwise.
func StartSAPSystemDiscovery(ctx context.Context, config *cpb.Configuration, d *Discovery) bool {

	d.sapInstancesRoutine = &recovery.RecoverableRoutine{
		Routine:             updateSAPInstances,
		RoutineArg:          updateSapInstancesArgs{config, d},
		ErrorCode:           usagemetrics.DiscoverSapInstanceFailure,
		ExpectedMinDuration: 5 * time.Second,
	}
	d.sapInstancesRoutine.StartRoutine(ctx)

	// Ensure SAP instances is populated before starting system discovery
	backoff.Retry(func() error {
		if d.GetSAPInstances() != nil {
			return nil
		}
		return fmt.Errorf("SAP Instances not ready yet")
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 120))

	d.systemDiscoveryRoutine = &recovery.RecoverableRoutine{
		Routine:             runDiscovery,
		RoutineArg:          runDiscoveryArgs{config, d},
		ErrorCode:           usagemetrics.DiscoverSapSystemFailure,
		ExpectedMinDuration: 10 * time.Second,
	}

	d.systemDiscoveryRoutine.StartRoutine(ctx)

	// Ensure systems are populated before returning
	backoff.Retry(func() error {
		if d.GetSAPSystems() != nil {
			return nil
		}
		return fmt.Errorf("SAP Systems not ready yet")
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 120))
	return true
}

type cloudLogInterface interface {
	Log(e logging.Entry)
	Flush() error
}

type wlmInterface interface {
	WriteInsight(project, location string, writeInsightRequest *workloadmanager.WriteInsightRequest) error
}

type cloudDiscoveryInterface interface {
	DiscoverComputeResources(context.Context, *spb.SapDiscovery_Resource, []string, *ipb.CloudProperties) []*spb.SapDiscovery_Resource
}

type hostDiscoveryInterface interface {
	DiscoverCurrentHost(context.Context) []string
}

type sapDiscoveryInterface interface {
	DiscoverSAPApps(ctx context.Context, sapApps *sappb.SAPInstances, conf *cpb.DiscoveryConfiguration) []appsdiscovery.SapSystemDetails
}

func removeDuplicates(res []*spb.SapDiscovery_Resource) []*spb.SapDiscovery_Resource {
	var out []*spb.SapDiscovery_Resource
	uris := make(map[string]*spb.SapDiscovery_Resource)
	for _, r := range res {
		outRes, ok := uris[r.ResourceUri]
		if !ok {
			uris[r.ResourceUri] = r
			out = append(out, r)
		} else {
			for _, rel := range r.RelatedResources {
				if !slices.Contains(outRes.RelatedResources, rel) {
					outRes.RelatedResources = append(outRes.RelatedResources, rel)
				}
			}
		}
	}
	return out
}

func insightResourceFromSystemResource(r *spb.SapDiscovery_Resource) *workloadmanager.SapDiscoveryResource {
	o := &workloadmanager.SapDiscoveryResource{
		RelatedResources: r.RelatedResources,
		ResourceKind:     r.ResourceKind.String(),
		ResourceType:     r.ResourceType.String(),
		ResourceUri:      r.ResourceUri,
		UpdateTime:       r.UpdateTime.AsTime().Format(time.RFC3339),
	}
	if r.GetInstanceProperties() != nil {
		o.InstanceProperties = &workloadmanager.SapDiscoveryResourceInstanceProperties{
			VirtualHostname:  r.InstanceProperties.GetVirtualHostname(),
			ClusterInstances: r.InstanceProperties.GetClusterInstances(),
		}
	}
	return o
}

func insightComponentFromSystemComponent(comp *spb.SapDiscovery_Component) *workloadmanager.SapDiscoveryComponent {
	iComp := &workloadmanager.SapDiscoveryComponent{
		HostProject: comp.HostProject,
		Sid:         comp.Sid,
		HaHosts:     comp.HaHosts,
	}

	for _, r := range comp.Resources {
		iComp.Resources = append(iComp.Resources, insightResourceFromSystemResource(r))
	}

	switch x := comp.Properties.(type) {
	case *spb.SapDiscovery_Component_ApplicationProperties_:
		iComp.ApplicationProperties = &workloadmanager.SapDiscoveryComponentApplicationProperties{
			ApplicationType: x.ApplicationProperties.GetApplicationType().String(),
			AscsUri:         x.ApplicationProperties.GetAscsUri(),
			NfsUri:          x.ApplicationProperties.GetNfsUri(),
			Abap:            x.ApplicationProperties.GetAbap(),
			KernelVersion:   x.ApplicationProperties.GetKernelVersion(),
		}
	case *spb.SapDiscovery_Component_DatabaseProperties_:
		iComp.DatabaseProperties = &workloadmanager.SapDiscoveryComponentDatabaseProperties{
			DatabaseType:       x.DatabaseProperties.GetDatabaseType().String(),
			PrimaryInstanceUri: x.DatabaseProperties.GetPrimaryInstanceUri(),
			SharedNfsUri:       x.DatabaseProperties.GetSharedNfsUri(),
			DatabaseVersion:    x.DatabaseProperties.GetDatabaseVersion(),
		}
	}

	return iComp
}

func insightFromSAPSystem(sys *spb.SapDiscovery) *workloadmanager.Insight {
	iDiscovery := &workloadmanager.SapDiscovery{
		SystemId:      sys.SystemId,
		ProjectNumber: sys.ProjectNumber,
		UpdateTime:    sys.UpdateTime.AsTime().Format(time.RFC3339),
	}
	if sys.ApplicationLayer != nil {
		iDiscovery.ApplicationLayer = insightComponentFromSystemComponent(sys.ApplicationLayer)

	}
	if sys.DatabaseLayer != nil {
		iDiscovery.DatabaseLayer = insightComponentFromSystemComponent(sys.DatabaseLayer)
	}

	return &workloadmanager.Insight{SapDiscovery: iDiscovery}
}

type updateSapInstancesArgs struct {
	config *cpb.Configuration
	d      *Discovery
}

type runDiscoveryArgs struct {
	config *cpb.Configuration
	d      *Discovery
}

func updateSAPInstances(ctx context.Context, a any) {
	var args updateSapInstancesArgs
	var ok bool
	if args, ok = a.(updateSapInstancesArgs); !ok {
		log.CtxLogger(ctx).Warn("args is not of type updateSapInstancesArgs")
		return
	}

	log.CtxLogger(ctx).Info("Starting SAP Instances update")
	updateTicker := time.NewTicker(args.config.GetDiscoveryConfiguration().GetSapInstancesUpdateFrequency().AsDuration())
	for {
		log.CtxLogger(ctx).Info("Updating SAP Instances")
		sapInst := args.d.AppsDiscovery(ctx)
		args.d.sapMu.Lock()
		args.d.sapInstances = sapInst
		args.d.sapMu.Unlock()

		select {
		case <-ctx.Done():
			log.CtxLogger(ctx).Info("SAP Discovery cancellation requested")
			return
		case <-updateTicker.C:
			continue
		}
	}
}

func runDiscovery(ctx context.Context, a any) {
	var args runDiscoveryArgs
	var ok bool
	if args, ok = a.(runDiscoveryArgs); !ok {
		log.CtxLogger(ctx).Warn("args is not of type runDiscoveryArgs")
		return
	}
	cp := args.config.GetCloudProperties()
	if cp == nil {
		log.CtxLogger(ctx).Warn("No Metadata Cloud Properties found, cannot collect resource information from the Compute API")
		return
	}

	updateTicker := time.NewTicker(args.config.GetDiscoveryConfiguration().GetSystemDiscoveryUpdateFrequency().AsDuration())
	for {
		sapSystems := args.d.discoverSAPSystems(ctx, cp)

		locationParts := strings.Split(cp.GetZone(), "-")
		region := strings.Join([]string{locationParts[0], locationParts[1]}, "-")

		// Write SAP system discovery data only if sap_system_discovery is enabled.
		if args.config.GetDiscoveryConfiguration().GetEnableDiscovery().GetValue() {
			log.CtxLogger(ctx).Info("Sending systems to WLM API")
			for _, sys := range sapSystems {
				// Send System to DW API
				req := &workloadmanager.WriteInsightRequest{
					Insight: insightFromSAPSystem(sys),
				}
				req.Insight.InstanceId = cp.GetInstanceId()

				err := args.d.WlmService.WriteInsight(cp.ProjectId, region, req)
				if err != nil {
					log.CtxLogger(ctx).Warnw("Encountered error writing to WLM", "error", err)
				}

				if args.d.CloudLogInterface == nil {
					continue
				}
				err = args.d.writeToCloudLogging(sys)
				if err != nil {
					log.CtxLogger(ctx).Warnw("Encountered error writing to cloud logging", "error", err)
				}
			}
		}

		log.CtxLogger(ctx).Info("Done SAP System Discovery")

		args.d.systemMu.Lock()
		args.d.systems = sapSystems
		args.d.systemMu.Unlock()

		select {
		case <-ctx.Done():
			log.CtxLogger(ctx).Info("SAP Discovery cancellation requested")
			return
		case <-updateTicker.C:
			continue
		}
	}
}

func (d *Discovery) discoverSAPSystems(ctx context.Context, cp *ipb.CloudProperties) []*spb.SapDiscovery {
	sapSystems := []*spb.SapDiscovery{}

	instanceURI := fmt.Sprintf("projects/%s/zones/%s/instances/%s", cp.GetProjectId(), cp.GetZone(), cp.GetInstanceName())
	log.CtxLogger(ctx).Info("Starting SAP Discovery")
	sapDetails := d.SapDiscoveryInterface.DiscoverSAPApps(ctx, d.GetSAPInstances(), nil)
	log.CtxLogger(ctx).Debugf("SAP Details: %v", sapDetails)
	log.CtxLogger(ctx).Info("Starting host discovery")
	hostResourceNames := d.HostDiscoveryInterface.DiscoverCurrentHost(ctx)
	log.CtxLogger(ctx).Debugf("Host Resource Names: %v", hostResourceNames)
	log.CtxLogger(ctx).Debug("Discovering current host")
	hostResources := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, nil, append([]string{instanceURI}, hostResourceNames...), cp)
	log.CtxLogger(ctx).Debugf("Host Resources: %v", hostResources)
	var instanceResource *spb.SapDiscovery_Resource
	// Find the instance resource
	for _, r := range hostResources {
		if strings.Contains(r.ResourceUri, cp.GetInstanceName()) {
			log.CtxLogger(ctx).Debugf("Instance Resource: %v", r)
			instanceResource = r
			break
		}
	}
	if instanceResource == nil {
		log.CtxLogger(ctx).Debug("No instance resource found")
	}
	for _, s := range sapDetails {
		system := &spb.SapDiscovery{}
		if s.AppComponent != nil {
			log.CtxLogger(ctx).Info("Discovering cloud resources for app")
			appRes := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, instanceResource, s.AppHosts, cp)
			log.CtxLogger(ctx).Debugf("App Resources: %v", appRes)
			if s.AppOnHost {
				appRes = append(appRes, hostResources...)
				log.CtxLogger(ctx).Debugf("App On Host Resources: %v", appRes)
			}
			if s.AppComponent.GetApplicationProperties().GetNfsUri() != "" {
				log.CtxLogger(ctx).Info("Discovering cloud resources for app NFS")
				nfsRes := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, instanceResource, []string{s.AppComponent.GetApplicationProperties().GetNfsUri()}, cp)
				if len(nfsRes) > 0 {
					appRes = append(appRes, nfsRes...)
					s.AppComponent.GetApplicationProperties().NfsUri = nfsRes[0].GetResourceUri()
				}
			}
			if s.AppComponent.GetApplicationProperties().GetAscsUri() != "" {
				log.CtxLogger(ctx).Info("Discovering cloud resources for app ASCS")
				ascsRes := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, instanceResource, []string{s.AppComponent.GetApplicationProperties().GetAscsUri()}, cp)
				if len(ascsRes) > 0 {
					log.CtxLogger(ctx).Debugw("ASCS Resources", "res", ascsRes)
					appRes = append(appRes, ascsRes...)
					s.AppComponent.GetApplicationProperties().AscsUri = ascsRes[0].GetResourceUri()
				}
			}
			if len(s.AppComponent.GetHaHosts()) > 0 {
				haRes := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, instanceResource, s.AppComponent.GetHaHosts(), cp)
				// Find the instances
				var haURIs []string
				for _, res := range haRes {
					if res.GetResourceKind() == spb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE {
						haURIs = append(haURIs, res.GetResourceUri())
					}
				}
				appRes = append(appRes, haRes...)
				s.AppComponent.HaHosts = haURIs
			}
			s.AppComponent.HostProject = cp.GetNumericProjectId()
			s.AppComponent.Resources = removeDuplicates(appRes)
			system.ApplicationLayer = s.AppComponent
		}
		if s.DBComponent != nil {
			log.CtxLogger(ctx).Info("Discovering cloud resources for database")
			dbRes := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, instanceResource, s.DBHosts, cp)
			if s.DBOnHost {
				dbRes = append(dbRes, hostResources...)
			}
			if s.DBComponent.GetDatabaseProperties().GetSharedNfsUri() != "" {
				log.CtxLogger(ctx).Info("Discovering cloud resources for database NFS")
				nfsRes := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, instanceResource, []string{s.DBComponent.GetDatabaseProperties().GetSharedNfsUri()}, cp)
				if len(nfsRes) > 0 {
					dbRes = append(dbRes, nfsRes...)
					s.DBComponent.GetDatabaseProperties().SharedNfsUri = nfsRes[0].GetResourceUri()
				}
			}
			if len(s.DBComponent.GetHaHosts()) > 0 {

				haRes := d.CloudDiscoveryInterface.DiscoverComputeResources(ctx, instanceResource, s.DBComponent.GetHaHosts(), cp)
				// Find the instances
				var haURIs []string
				for _, res := range haRes {
					if res.GetResourceKind() == spb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE {
						haURIs = append(haURIs, res.GetResourceUri())
					}
				}
				dbRes = append(dbRes, haRes...)
				s.DBComponent.HaHosts = haURIs
			}
			s.DBComponent.HostProject = cp.GetNumericProjectId()
			s.DBComponent.Resources = removeDuplicates(dbRes)
			system.DatabaseLayer = s.DBComponent
		}
		system.ProjectNumber = cp.GetNumericProjectId()
		system.UpdateTime = timestamppb.Now()
		sapSystems = append(sapSystems, system)
	}
	return sapSystems
}

func (d *Discovery) writeToCloudLogging(sys *spb.SapDiscovery) error {
	s, err := protojson.Marshal(sys)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	json.Indent(&buf, s, "", "  ")

	payload := make(map[string]string)
	payload["type"] = "SapDiscovery"
	payload["discovery"] = buf.String()

	d.CloudLogInterface.Log(logging.Entry{
		Timestamp: time.Now(),
		Severity:  logging.Info,
		Payload:   payload,
	})

	return nil
}
