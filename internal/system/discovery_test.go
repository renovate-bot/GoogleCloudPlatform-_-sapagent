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

package system

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	dpb "google.golang.org/protobuf/types/known/durationpb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
	sappb "github.com/GoogleCloudPlatform/sapagent/protos/sapapp"

	"cloud.google.com/go/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/GoogleCloudPlatform/sapagent/internal/configuration"
	"github.com/GoogleCloudPlatform/sapagent/internal/system/appsdiscovery"
	appsdiscoveryfake "github.com/GoogleCloudPlatform/sapagent/internal/system/appsdiscovery/fake"
	clouddiscoveryfake "github.com/GoogleCloudPlatform/sapagent/internal/system/clouddiscovery/fake"
	hostdiscoveryfake "github.com/GoogleCloudPlatform/sapagent/internal/system/hostdiscovery/fake"
	"github.com/GoogleCloudPlatform/sapagent/internal/system/hostdiscovery"
	"github.com/GoogleCloudPlatform/sapagent/internal/workloadmanager"

	cpb "github.com/GoogleCloudPlatform/sapagent/protos/configuration"
	instancepb "github.com/GoogleCloudPlatform/sapagent/protos/instanceinfo"
	wlmfake "github.com/GoogleCloudPlatform/workloadagentplatform/sharedlibraries/gce/fake"
	logfake "github.com/GoogleCloudPlatform/workloadagentplatform/sharedlibraries/log/fake"
	"github.com/GoogleCloudPlatform/workloadagentplatform/sharedlibraries/log"
	dwpb "github.com/GoogleCloudPlatform/workloadagentplatform/sharedprotos/datawarehouse"
	statpb "github.com/GoogleCloudPlatform/workloadagentplatform/sharedprotos/status"
	syspb "github.com/GoogleCloudPlatform/workloadagentplatform/sharedprotos/system"
)

const (
	defaultInstanceName  = "test-instance-id"
	defaultProjectID     = "test-project-id"
	defaultZone          = "test-zone-a"
	secondaryZone        = "test-zone-b"
	defaultClusterOutput = `
	line1
	line2
	rsc_vip_int-primary IPaddr2
	anotherline
	params ip 127.0.0.1 other text
	line3
	line4
	`
	defaultUserstoreOutput = `
KEY default
	ENV: 
	a:b:c
  ENV : test-instance:30013
  USER: SAPABAP1
  DATABASE: DEH
Operation succeed.
`
	defaultSID                       = "ABC"
	defaultInstanceNumber            = "00"
	defaultLandscapeOutputSingleNode = `
	| Host        | Host   | Host   | Failover | Remove | Storage   | Storage   | Failover | Failover | NameServer | NameServer | IndexServer | IndexServer | Host    | Host    | Worker  | Worker  |
|             | Active | Status | Status   | Status | Config    | Actual    | Config   | Actual   | Config     | Actual     | Config      | Actual      | Config  | Actual  | Config  | Actual  |
|             |        |        |          |        | Partition | Partition | Group    | Group    | Role       | Role       | Role        | Role        | Roles   | Roles   | Groups  | Groups  |
| ----------- | ------ | ------ | -------- | ------ | --------- | --------- | -------- | -------- | ---------- | ---------- | ----------- | ----------- | ------- | ------- | ------- | ------- |
| dru-s4dan   | yes    | info   |          |        |         1 |         0 | default  | default  | master 1   | slave      | worker      | standby     | worker  | standby | default | -       |

overall host status: info
`
	defaultLandscapeOutputMultipleNodes = `
| Host        | Host   | Host   | Failover | Remove | Storage   | Storage   | Failover | Failover | NameServer | NameServer | IndexServer | IndexServer | Host    | Host    | Worker  | Worker  |
|             | Active | Status | Status   | Status | Config    | Actual    | Config   | Actual   | Config     | Actual     | Config      | Actual      | Config  | Actual  | Config  | Actual  |
|             |        |        |          |        | Partition | Partition | Group    | Group    | Role       | Role       | Role        | Role        | Roles   | Roles   | Groups  | Groups  |
| ----------- | ------ | ------ | -------- | ------ | --------- | --------- | -------- | -------- | ---------- | ---------- | ----------- | ----------- | ------- | ------- | ------- | ------- |
| dru-s4dan   | yes    | info   |          |        |         1 |         0 | default  | default  | master 1   | slave      | worker      | standby     | worker  | standby | default | -       |
| dru-s4danw1 | yes    | ok     |          |        |         2 |         2 | default  | default  | master 2   | slave      | worker      | slave       | worker  | worker  | default | default |
| dru-s4danw2 | yes    | ok     |          |        |         3 |         3 | default  | default  | slave      | slave      | worker      | slave       | worker  | worker  | default | default |
| dru-s4danw3 | yes    | info   |          |        |         0 |         1 | default  | default  | master 3   | master     | standby     | master      | standby | worker  | default | default |

overall host status: info
`
	defaultSubnetwork = "test-subnetwork"
)

var (
	defaultInstanceURI     = makeZonalURI(defaultProjectID, defaultZone, "instances", defaultInstanceName)
	secondaryInstanceURI   = makeZonalURI(defaultProjectID, secondaryZone, "instances", "secondary-instance-id")
	defaultCloudProperties = &instancepb.CloudProperties{
		InstanceName:     defaultInstanceName,
		ProjectId:        defaultProjectID,
		Zone:             defaultZone,
		NumericProjectId: "12345",
	}
	resourceListDiffOpts = []cmp.Option{
		protocmp.Transform(),
		protocmp.IgnoreFields(&syspb.SapDiscovery_Resource{}, "update_time"),
		protocmp.SortRepeatedFields(&syspb.SapDiscovery_Resource{}, "related_resources"),
		protocmp.SortRepeatedFields(&syspb.SapDiscovery_Component{}, "resources"),
		cmpopts.SortSlices(resourceLess),
		protocmp.SortRepeatedFields(&syspb.SapDiscovery_Resource_InstanceProperties{}, "app_instances", "disk_device_names", "disk_mounts"),
		cmpopts.SortSlices(appInstanceLess),
	}
	defaultInstanceResource = &syspb.SapDiscovery_Resource{
		ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
		ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
		ResourceUri:  defaultInstanceURI,
	}
)

func TestMain(t *testing.M) {
	log.SetupLoggingForTest()
	os.Exit(t.Run())
}

func resourceLess(a, b *syspb.SapDiscovery_Resource) bool {
	return a.String() < b.String()
}

func appInstanceLess(a, b *syspb.SapDiscovery_Resource_InstanceProperties_AppInstance) bool {
	return a.Name < b.Name
}

func makeZonalURI(project, zone, resType, name string) string {
	return fmt.Sprintf("projects/%s/zones/%s/%s/%s", project, zone, resType, name)
}

type MockFileInfo struct {
}

func (mfi MockFileInfo) Name() string       { return "name" }
func (mfi MockFileInfo) Size() int64        { return int64(8) }
func (mfi MockFileInfo) Mode() os.FileMode  { return os.ModePerm }
func (mfi MockFileInfo) ModTime() time.Time { return time.Now() }
func (mfi MockFileInfo) IsDir() bool        { return false }
func (mfi MockFileInfo) Sys() any           { return nil }

func TestStartSAPSystemDiscovery(t *testing.T) {
	config := &cpb.Configuration{
		CloudProperties: defaultCloudProperties,
		DiscoveryConfiguration: &cpb.DiscoveryConfiguration{
			EnableDiscovery:                &wpb.BoolValue{Value: true},
			SapInstancesUpdateFrequency:    &dpb.Duration{Seconds: 10},
			SystemDiscoveryUpdateFrequency: &dpb.Duration{Seconds: 10},
		},
	}

	d := &Discovery{
		SapDiscoveryInterface: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{}},
		},
		CloudDiscoveryInterface: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{}, {}},
		},
		HostDiscoveryInterface: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		AppsDiscovery:     func(context.Context, SapSystemDiscoveryInterface) *sappb.SAPInstances { return &sappb.SAPInstances{} },
		CloudLogInterface: &logfake.TestCloudLogging{FlushErr: []error{nil}},
		OSStatReader:      func(string) (os.FileInfo, error) { return nil, nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	got := StartSAPSystemDiscovery(ctx, config, d)
	if got != true {
		t.Errorf("StartSAPSystemDiscovery(%#v) = %t, want: %t", config, got, true)
	}
	cancel()
}

func TestDiscoverSAPSystems(t *testing.T) {
	tests := []struct {
		name               string
		config             *cpb.Configuration
		testSapDiscovery   *appsdiscoveryfake.SapDiscovery
		testCloudDiscovery *clouddiscoveryfake.CloudDiscovery
		testHostDiscovery  *hostdiscoveryfake.HostDiscovery
		testOSStatReader   workloadmanager.OSStatReader
		testFileReader     workloadmanager.ConfigFileReader
		want               []*syspb.SapDiscovery
	}{{
		name:   "noDiscovery",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{}, {}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{},
	}, {
		name:   "justHANA",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
						DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
							SharedNfsUri:   "some-shared-nfs-uri",
							InstanceNumber: "00",
						},
					},
				},
				DBHosts: []string{"some-db-host"},
				InstanceProperties: []*syspb.SapDiscovery_Resource_InstanceProperties{{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					VirtualHostname: "some-db-host",
				}},
				DBOnHost: true,
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {},
				{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
					DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
						SharedNfsUri:   "some-shared-nfs-uri",
						InstanceNumber: "00",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
						VirtualHostname: "some-db-host",
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "hanaWithDiskMounts",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
						DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
							SharedNfsUri:   "some-shared-nfs-uri",
							InstanceNumber: "00",
						},
					},
				},
				DBHosts: []string{"some-db-host"},
				InstanceProperties: []*syspb.SapDiscovery_Resource_InstanceProperties{{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					VirtualHostname: "some-db-host",
				}},
				DBOnHost:  true,
				DBDiskMap: map[string][]string{"/hana/data": {"hana-data"}, "/hana/log": {"hana-log"}},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					DiskDeviceNames: []*syspb.SapDiscovery_Resource_InstanceProperties_DiskDeviceName{
						&syspb.SapDiscovery_Resource_InstanceProperties_DiskDeviceName{
							DeviceName: "hana-data",
							Source:     "hana-data-source",
						},
						&syspb.SapDiscovery_Resource_InstanceProperties_DiskDeviceName{
							DeviceName: "hana-log",
							Source:     "hana-log-source",
						},
					},
				},
			}}, {},
				{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-cluster-host", "some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				ClusterAddrs: []string{"some-cluster-host"},
				NFSAddrs:     []string{"some-shared-nfs-uri"},
				KernelVersion: &statpb.KernelVersion{
					RawString: "kernel-version",
					OsKernel: &statpb.KernelVersion_Version{
						Major:     1,
						Minor:     2,
						Build:     3,
						Patch:     4,
						Remainder: "5",
					},
					DistroKernel: &statpb.KernelVersion_Version{
						Major:     6,
						Minor:     7,
						Build:     8,
						Patch:     9,
						Remainder: "10",
					},
				},
			}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
					DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
						SharedNfsUri:   "some-shared-nfs-uri",
						InstanceNumber: "00",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
						VirtualHostname: "some-db-host",
						DiskDeviceNames: []*syspb.SapDiscovery_Resource_InstanceProperties_DiskDeviceName{
							&syspb.SapDiscovery_Resource_InstanceProperties_DiskDeviceName{
								DeviceName: "hana-data",
								Source:     "hana-data-source",
							},
							&syspb.SapDiscovery_Resource_InstanceProperties_DiskDeviceName{
								DeviceName: "hana-log",
								Source:     "hana-log-source",
							},
						},
						DiskMounts: []*syspb.SapDiscovery_Resource_InstanceProperties_DiskMount{
							&syspb.SapDiscovery_Resource_InstanceProperties_DiskMount{
								DiskNames:  []string{"hana-log-source"},
								MountPoint: "/hana/log",
							},
							&syspb.SapDiscovery_Resource_InstanceProperties_DiskMount{
								DiskNames:  []string{"hana-data-source"},
								MountPoint: "/hana/data",
							},
						},
						OsKernelVersion: &statpb.KernelVersion{
							RawString: "kernel-version",
							OsKernel: &statpb.KernelVersion_Version{
								Major:     1,
								Minor:     2,
								Build:     3,
								Patch:     4,
								Remainder: "5",
							},
							DistroKernel: &statpb.KernelVersion_Version{
								Major:     6,
								Minor:     7,
								Build:     8,
								Patch:     9,
								Remainder: "10",
							},
						},
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "justApp",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
						ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
							NfsUri:  "some-nfs-host",
							AscsUri: "some-ascs-host",
						},
					},
				},
				AppHosts: []string{"some-app-host"},
				WorkloadProperties: &syspb.SapDiscovery_WorkloadProperties{
					ProductVersions: []*syspb.SapDiscovery_WorkloadProperties_ProductVersion{
						&syspb.SapDiscovery_WorkloadProperties_ProductVersion{
							Name:    "some-product-name",
							Version: "some-product-version",
						},
					},
					SoftwareComponentVersions: []*syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
						&syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
							Name:       "some-software-component-name",
							Version:    "some-software-component-version",
							ExtVersion: "some-software-component-ext-version",
							Type:       "some-software-component-type",
						},
					},
				},
				InstanceProperties: []*syspb.SapDiscovery_Resource_InstanceProperties{{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
					VirtualHostname: "some-app-host",
					AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
						Name:   "some-app-instance-name",
						Number: "99",
					}},
				}},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-nfs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  "some-ascs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-nfs-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-ascs-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-host"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
					ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
						AscsUri: "some-ascs-uri",
						NfsUri:  "some-nfs-uri",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
						VirtualHostname: "some-app-host",
						AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
							Name:   "some-app-instance-name",
							Number: "99",
						}},
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-nfs-uri",
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some-ascs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
			WorkloadProperties: &syspb.SapDiscovery_WorkloadProperties{
				ProductVersions: []*syspb.SapDiscovery_WorkloadProperties_ProductVersion{
					&syspb.SapDiscovery_WorkloadProperties_ProductVersion{
						Name:    "some-product-name",
						Version: "some-product-version",
					},
				},
				SoftwareComponentVersions: []*syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
					&syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
						Name:       "some-software-component-name",
						Version:    "some-software-component-version",
						ExtVersion: "some-software-component-ext-version",
						Type:       "some-software-component-type",
					},
				},
			},
		}},
	}, {
		name:   "noASCSResource",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
						ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
							NfsUri:  "some-nfs-uri",
							AscsUri: "some-ascs-uri",
						},
					},
				},
				AppHosts: []string{"some-app-host"},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-nfs-uri",
			}}, {}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-ascs-uri"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
					ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
						AscsUri: "some-ascs-uri",
						NfsUri:  "some-nfs-uri",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-nfs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "noNFSResource",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
						ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
							NfsUri:  "some-nfs-uri",
							AscsUri: "some-ascs-uri",
						},
					},
				},
				AppHosts: []string{"some-app-host"},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{
				{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}, {}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}, {}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some-ascs-uri",
				}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-ascs-uri"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
					ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
						AscsUri: "some-ascs-uri",
						NfsUri:  "some-nfs-uri",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some-ascs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "appAndDB",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
				},
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "DEF",
				},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{defaultInstanceResource}, {}, {}, {}, {}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid:         "ABC",
				HostProject: "12345",
			},
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid:         "DEF",
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "appOnHost",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
				},
				AppOnHost: true,
				AppHosts:  []string{"some-app-resource"},
			}}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				ClusterAddrs: []string{"some-host-resource"},
			}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceUri:      defaultInstanceURI,
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{"some-host-resource"},
			}}, {{
				ResourceUri:      "some-host-resource",
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{defaultInstanceURI},
			}}, {{
				ResourceUri:  "some-app-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-host-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-resource"},
				CP:       defaultCloudProperties,
			}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:      defaultInstanceURI,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{"some-host-resource"},
				}, {
					ResourceUri:  "some-app-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}, {
					ResourceUri:      "some-host-resource",
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{defaultInstanceURI},
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "DBOnHostNoReplication",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "DEF",
				},
				DBOnHost: true,
				DBHosts:  []string{"some-db-resource"},
			}}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				ClusterAddrs: []string{"some-host-resource"},
			}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceUri:      defaultInstanceURI,
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{"some-host-resource"},
			}}, {{
				ResourceUri:      "some-host-resource",
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{defaultInstanceURI},
			}}, {{
				ResourceUri:  "some-db-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-host-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-resource"},
				CP:       defaultCloudProperties,
			}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "DEF",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:      defaultInstanceURI,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{"some-host-resource"},
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceUri:  "some-db-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}, {
					ResourceUri:      "some-host-resource",
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{defaultInstanceURI},
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "appAndDBOnHost",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
				},
				AppOnHost: true,
				AppHosts:  []string{"some-app-resource"},
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "DEF",
				},
				DBOnHost: true,
				DBHosts:  []string{"some-db-resource"},
			}}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				ClusterAddrs: []string{"some-host-resource"},
			}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceUri:      defaultInstanceURI,
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{"some-host-resource"},
			}}, {{
				ResourceUri:      "some-host-resource",
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{defaultInstanceURI},
			}}, {{
				ResourceUri:  "some-app-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}, {{
				ResourceUri:  "some-db-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-host-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-resource"},
				CP:       defaultCloudProperties,
			}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:      defaultInstanceURI,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{"some-host-resource"},
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceUri:  "some-app-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}, {
					ResourceUri:      "some-host-resource",
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{defaultInstanceURI},
				}},
				HostProject: "12345",
			},
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "DEF",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:      defaultInstanceURI,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{"some-host-resource"},
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceUri:  "some-db-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}, {
					ResourceUri:      "some-host-resource",
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{defaultInstanceURI},
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "appOnHostDBOffHost",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
				},
				AppOnHost: true,
				AppHosts:  []string{"some-app-resource"},
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "DEF",
				},
				DBOnHost: false,
				DBHosts:  []string{"some-db-resource"},
			}}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				ClusterAddrs: []string{"some-host-resource"},
			}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceUri:      defaultInstanceURI,
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{"some-host-resource"},
			}}, {{
				ResourceUri:      "some-host-resource",
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{defaultInstanceURI},
			}}, {{
				ResourceUri:  "some-app-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}, {{
				ResourceUri:  "some-db-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-host-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-resource"},
				CP:       defaultCloudProperties,
			}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:      defaultInstanceURI,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{"some-host-resource"},
				}, {
					ResourceUri:  "some-app-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}, {
					ResourceUri:      "some-host-resource",
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{defaultInstanceURI},
				}},
				HostProject: "12345",
			},
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "DEF",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:  "some-db-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "DBOnHostAppOffHost",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
				},
				AppOnHost: false,
				AppHosts:  []string{"some-app-resource"},
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "DEF",
				},
				DBOnHost: true,
				DBHosts:  []string{"some-db-resource"},
			}}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				ClusterAddrs: []string{"some-host-resource"},
			}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceUri:      defaultInstanceURI,
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{"some-host-resource"},
			}}, {{
				ResourceUri:      "some-host-resource",
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				RelatedResources: []string{defaultInstanceURI},
			}}, {{
				ResourceUri:  "some-app-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}, {{
				ResourceUri:  "some-db-resource",
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-host-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-resource"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-resource"},
				CP:       defaultCloudProperties,
			}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:  "some-app-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}},
				HostProject: "12345",
			},
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "DEF",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceUri:      defaultInstanceURI,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{"some-host-resource"},
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceUri:  "some-db-resource",
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_ADDRESS,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				}, {
					ResourceUri:      "some-host-resource",
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					RelatedResources: []string{defaultInstanceURI},
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "databaseIPropNotAlreadyDiscovered",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
						DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
							SharedNfsUri: "some-shared-nfs-uri",
						},
					},
				},
				DBHosts: []string{"some-db-host"},
				InstanceProperties: []*syspb.SapDiscovery_Resource_InstanceProperties{{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					VirtualHostname: "some-other-db-host",
				}},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-shared-nfs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  "some/other/instance",
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-other-db-host"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
					DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
						SharedNfsUri: "some-shared-nfs-uri",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some/other/instance",
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
						VirtualHostname: "some-other-db-host",
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "databaseIPropMergesWtihDiscoveredIProp",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
						DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
							SharedNfsUri: "some-shared-nfs-uri",
						},
					},
				},
				DBHosts:  []string{"some-db-host"},
				DBOnHost: true,
				InstanceProperties: []*syspb.SapDiscovery_Resource_InstanceProperties{{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					VirtualHostname: "some-db-host",
				}},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
					VirtualHostname: "old-db-host",
				},
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-shared-nfs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}, {
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
				ResourceUri:  "some/disk/uri",
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
					DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
						SharedNfsUri: "some-shared-nfs-uri",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER | syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
						VirtualHostname: "some-db-host",
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceUri:  "some/disk/uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "appIPropNotAlreadyDiscovered",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
						ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
							NfsUri:  "some-nfs-host",
							AscsUri: "some-ascs-host",
						},
					},
				},
				AppHosts: []string{"some-app-host"},
				WorkloadProperties: &syspb.SapDiscovery_WorkloadProperties{
					ProductVersions: []*syspb.SapDiscovery_WorkloadProperties_ProductVersion{
						&syspb.SapDiscovery_WorkloadProperties_ProductVersion{
							Name:    "some-product-name",
							Version: "some-product-version",
						},
					},
					SoftwareComponentVersions: []*syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
						&syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
							Name:       "some-software-component-name",
							Version:    "some-software-component-version",
							ExtVersion: "some-software-component-ext-version",
							Type:       "some-software-component-type",
						},
					},
				},
				InstanceProperties: []*syspb.SapDiscovery_Resource_InstanceProperties{{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
					VirtualHostname: "some-other-app-host",
					AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
						Name:   "some-app-instance-name",
						Number: "99",
					}},
				}},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-nfs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  "some-ascs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  "some/other/instance",
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
					VirtualHostname: "some-other-app-instance-name",
					AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
						Name:   "some-app-instance-name",
						Number: "99",
					}},
				},
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-nfs-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-ascs-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-other-app-host"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
					ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
						AscsUri: "some-ascs-uri",
						NfsUri:  "some-nfs-uri",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-nfs-uri",
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some-ascs-uri",
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some/other/instance",
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
						VirtualHostname: "some-other-app-host",
						AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
							Name:   "some-app-instance-name",
							Number: "99",
						}},
					},
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
			WorkloadProperties: &syspb.SapDiscovery_WorkloadProperties{
				ProductVersions: []*syspb.SapDiscovery_WorkloadProperties_ProductVersion{
					&syspb.SapDiscovery_WorkloadProperties_ProductVersion{
						Name:    "some-product-name",
						Version: "some-product-version",
					},
				},
				SoftwareComponentVersions: []*syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
					&syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
						Name:       "some-software-component-name",
						Version:    "some-software-component-version",
						ExtVersion: "some-software-component-ext-version",
						Type:       "some-software-component-type",
					},
				},
			},
		}},
	}, {
		name:   "appIPropMergesWithAlreadyDiscoveredIProp",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				AppComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
						ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
							NfsUri:  "some-nfs-host",
							AscsUri: "some-ascs-host",
						},
					},
				},
				AppHosts: []string{"some-app-host"},
				WorkloadProperties: &syspb.SapDiscovery_WorkloadProperties{
					ProductVersions: []*syspb.SapDiscovery_WorkloadProperties_ProductVersion{
						&syspb.SapDiscovery_WorkloadProperties_ProductVersion{
							Name:    "some-product-name",
							Version: "some-product-version",
						},
					},
					SoftwareComponentVersions: []*syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
						&syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
							Name:       "some-software-component-name",
							Version:    "some-software-component-version",
							ExtVersion: "some-software-component-ext-version",
							Type:       "some-software-component-type",
						},
					},
				},
				InstanceProperties: []*syspb.SapDiscovery_Resource_InstanceProperties{{
					InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_ERS,
					AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
						Name:   "some-ers-instance-name",
						Number: "88",
					}, {
						Name:   "some-app-instance-name",
						Number: "11",
					}, {
						Name:   "some-other-instance",
						Number: "12",
					}, {
						Name:   "some-other-instance",
						Number: "12",
					}},
				}},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
					VirtualHostname: "some-app-host",
					AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
						Name:   "some-app-instance-name",
						Number: "11",
					}},
				},
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-nfs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  "some-ascs-uri",
			}}, {{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER,
					VirtualHostname: "",
					AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
						Name:   "some-ers-instance-name",
						Number: "99",
					}},
				},
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-app-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-nfs-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-ascs-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{""},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_ApplicationProperties_{
					ApplicationProperties: &syspb.SapDiscovery_Component_ApplicationProperties{
						AscsUri: "some-ascs-uri",
						NfsUri:  "some-nfs-uri",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_APP_SERVER | syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_ERS,
						VirtualHostname: "some-app-host",
						AppInstances: []*syspb.SapDiscovery_Resource_InstanceProperties_AppInstance{{
							Name:   "some-ers-instance-name",
							Number: "99",
						}, {
							Name:   "some-app-instance-name",
							Number: "11",
						}, {
							Name:   "some-other-instance",
							Number: "12",
						}},
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-nfs-uri",
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some-ascs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
			WorkloadProperties: &syspb.SapDiscovery_WorkloadProperties{
				ProductVersions: []*syspb.SapDiscovery_WorkloadProperties_ProductVersion{
					&syspb.SapDiscovery_WorkloadProperties_ProductVersion{
						Name:    "some-product-name",
						Version: "some-product-version",
					},
				},
				SoftwareComponentVersions: []*syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
					&syspb.SapDiscovery_WorkloadProperties_SoftwareComponentProperties{
						Name:       "some-software-component-name",
						Version:    "some-software-component-version",
						ExtVersion: "some-software-component-ext-version",
						Type:       "some-software-component-type",
					},
				},
			},
		}},
	}, {
		name: "usesOverrideFile",
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		testOSStatReader: func(string) (os.FileInfo, error) {
			return MockFileInfo{}, nil
		},
		testFileReader: func(string) (io.ReadCloser, error) {
			return fakeReadCloser{
				fileContents: `{
					"databaseLayer": {
						"hostProject": "12345",
						"sid": "DEF"
					},
					"applicationLayer": {
						"hostProject": "12345",
						"sid": "ABC"
					},
					"projectNumber": "12345"
				}`,
			}, nil
		},
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid:         "ABC",
				HostProject: "12345",
			},
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid:         "DEF",
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "hostIsPrimary",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
				},
				DBOnHost: true,
				DBInstance: &sappb.SAPInstance{
					Sapsid:         "ABC",
					InstanceNumber: "00",
					HanaReplicationTree: &sappb.HANAReplicaSite{
						Name: "primary-site",
						Targets: []*sappb.HANAReplicaSite{
							{
								Name: "secondary-site",
							}},
					},
				},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				// Host instance
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				// Host instance resources
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}, {
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-shared-nfs-uri",
			}, {
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_BACKEND_SERVICE,
				ResourceUri:      "some-backend-service-uri",
				RelatedResources: []string{"primary-instance-group"},
			}, {
				ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
				ResourceUri:      "primary-instance-group",
				RelatedResources: []string{defaultInstanceURI},
			}}, {
				// Database resources
				{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				},
			}, {{
				// Primary site
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				// Secondary site
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  secondaryInstanceURI,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				// Host instance
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				// Host resources
				Parent:   defaultInstanceResource,
				HostList: []string{"1.2.3.4"},
				CP:       defaultCloudProperties,
			}, {
				// Database resources
				Parent: defaultInstanceResource,
				CP:     defaultCloudProperties,
			}, {
				// Primary site
				Parent:   defaultInstanceResource,
				HostList: []string{"primary-site"},
				CP:       defaultCloudProperties,
			}, {
				// Secondary site
				Parent:   defaultInstanceResource,
				HostList: []string{"secondary-site"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				NFSAddrs: []string{"1.2.3.4"},
			}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_BACKEND_SERVICE,
					ResourceUri:      "some-backend-service-uri",
					RelatedResources: []string{"primary-instance-group"},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
					ResourceUri:      "primary-instance-group",
					RelatedResources: []string{defaultInstanceURI},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}},
				HostProject: "12345",
				ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Sid: "ABC",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  secondaryInstanceURI,
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
							},
						}},
					}},
				},
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "multiTargetReplication",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
						DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
							SharedNfsUri:   "some-shared-nfs-uri",
							InstanceNumber: "00",
						},
					},
				},
				DBHosts:  []string{"some-db-host"},
				DBOnHost: true,
				DBInstance: &sappb.SAPInstance{
					Sapsid:         "ABC",
					InstanceNumber: "00",
					HanaReplicationTree: &sappb.HANAReplicaSite{
						Name: "primary-site",
						Targets: []*sappb.HANAReplicaSite{
							{
								Name: "secondary-site",
							},
							{
								Name: "tertiary-site",
							},
						},
					},
				},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				// Host instance
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				// Host resources
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {
				// Database resources
			}, {{
				// NFS resources
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-shared-nfs-uri",
			}}, {{
				// Primary site
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				// Secondary site
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  "secondary/site/resource",
			}}, {{
				// Tertiary site
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  "tertiary/site/resource",
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"primary-site"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"secondary-site"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"tertiary-site"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
					DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
						SharedNfsUri:   "some-shared-nfs-uri",
						InstanceNumber: "00",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}},
				HostProject: "12345",
				ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Sid: "ABC",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  "secondary/site/resource",
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
							},
						}},
					},
				}, {
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Sid: "ABC",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  "tertiary/site/resource",
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
							},
						}},
					},
				}},
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "multiTierReplication",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
						DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
							SharedNfsUri:   "some-shared-nfs-uri",
							InstanceNumber: "00",
						},
					},
				},
				DBHosts: []string{"some-db-host"},
				DBInstance: &sappb.SAPInstance{
					Sapsid:         "ABC",
					InstanceNumber: "00",
					HanaReplicationTree: &sappb.HANAReplicaSite{
						Name: "primary-site",
						Targets: []*sappb.HANAReplicaSite{
							{
								Name: "secondary-site",
								Targets: []*sappb.HANAReplicaSite{
									{
										Name: "tertiary-site",
									}},
							}},
					},
				},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}},
				{}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}}, {{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}}, {{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  "secondary/site/resource",
				}}, {{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  "tertiary/site/resource",
				}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: nil,
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"primary-site"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"secondary-site"},
				CP:       defaultCloudProperties,
			}, {
				Parent:   defaultInstanceResource,
				HostList: []string{"tertiary-site"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
					DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
						SharedNfsUri:   "some-shared-nfs-uri",
						InstanceNumber: "00",
					},
				},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}},
				HostProject: "12345",
				ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Sid: "ABC",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  "secondary/site/resource",
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
							},
						}},
						ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
							SourceSite: "secondary-site",
							Component: &syspb.SapDiscovery_Component{
								Sid: "ABC",
								Resources: []*syspb.SapDiscovery_Resource{{
									ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
									ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
									ResourceUri:  "tertiary/site/resource",
									InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
										InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
									},
								}},
							}},
						},
					},
				}},
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "primaryIsInCluster",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
					Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
						DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
							SharedNfsUri:   "some-shared-nfs-uri",
							InstanceNumber: "00",
						},
					},
				},
				DBOnHost: true,
				DBHosts:  []string{"some-db-host"},
				DBInstance: &sappb.SAPInstance{
					Sapsid:         "ABC",
					InstanceNumber: "00",
					HanaReplicationTree: &sappb.HANAReplicaSite{
						Name: "primary-site",
						Targets: []*sappb.HANAReplicaSite{
							{
								Name: "secondary-site",
							}},
					},
				},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				// Host instance
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				// Host resources
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {
				// Database resources
				{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_BACKEND_SERVICE,
					ResourceUri:      "some-backend-service-uri",
					RelatedResources: []string{"primary-instance-group", "secondary-instance-group"},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
					ResourceUri:      "primary-instance-group",
					RelatedResources: []string{defaultInstanceURI},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
					ResourceUri:      "secondary-instance-group",
					RelatedResources: []string{secondaryInstanceURI},
				}, {
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  secondaryInstanceURI,
				},
			}, {{
				// Database NFS
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-shared-nfs-uri",
			}}, {{
				// Primary site
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				// Secondary site
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  secondaryInstanceURI,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				// Host instance
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				// Host resources
				Parent:   defaultInstanceResource,
				HostList: []string{"1.2.3.4"},
				CP:       defaultCloudProperties,
			}, {
				// Database resources
				Parent:   defaultInstanceResource,
				HostList: []string{"some-db-host"},
				CP:       defaultCloudProperties,
			}, {
				// Database NFS
				Parent:   defaultInstanceResource,
				HostList: []string{"some-shared-nfs-uri"},
				CP:       defaultCloudProperties,
			}, {
				// Primary site
				Parent:   defaultInstanceResource,
				HostList: []string{"primary-site"},
				CP:       defaultCloudProperties,
			}, {
				// Secondary site
				Parent:   defaultInstanceResource,
				HostList: []string{"secondary-site"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				NFSAddrs: []string{"1.2.3.4"},
			}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Properties: &syspb.SapDiscovery_Component_DatabaseProperties_{
					DatabaseProperties: &syspb.SapDiscovery_Component_DatabaseProperties{
						SharedNfsUri:   "some-shared-nfs-uri",
						InstanceNumber: "00",
					},
				},
				HaHosts: []string{defaultInstanceURI, secondaryInstanceURI},
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  secondaryInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_BACKEND_SERVICE,
					ResourceUri:      "some-backend-service-uri",
					RelatedResources: []string{"primary-instance-group", "secondary-instance-group"},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
					ResourceUri:      "primary-instance-group",
					RelatedResources: []string{defaultInstanceURI},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
					ResourceUri:      "secondary-instance-group",
					RelatedResources: []string{secondaryInstanceURI},
				}, {
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
					ResourceUri:  "some-shared-nfs-uri",
				}},
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:   "hostIsNotPrimary",
		config: &cpb.Configuration{CloudProperties: defaultCloudProperties},
		testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
			DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
				DBComponent: &syspb.SapDiscovery_Component{
					Sid: "ABC",
				},
				DBOnHost: true,
				DBInstance: &sappb.SAPInstance{
					Sapsid:         "ABC",
					InstanceNumber: "00",
					HanaReplicationTree: &sappb.HANAReplicaSite{
						Name: "primary-site",
						Targets: []*sappb.HANAReplicaSite{
							{
								Name: "secondary-site",
							}},
					},
				},
			}}},
		},
		testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				// Host instance
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}}, {{
				// Host instance resources
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  defaultInstanceURI,
			}, {
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
				ResourceUri:  "some-shared-nfs-uri",
			}}, {
				// Database resources
				{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  defaultInstanceURI,
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_BACKEND_SERVICE,
					ResourceUri:      "some-backend-service-uri",
					RelatedResources: []string{"primary-instance-group"},
				}, {
					ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
					ResourceUri:      "primary-instance-group",
					RelatedResources: []string{defaultInstanceURI},
				},
			}, {{
				// Primary site
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceUri:  secondaryInstanceURI,
			}}, {{
				// Secondary site
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  defaultInstanceURI,
			}}},
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				// Host instance
				Parent:   nil,
				HostList: []string{defaultInstanceURI},
				CP:       defaultCloudProperties,
			}, {
				// Host resources
				Parent:   defaultInstanceResource,
				HostList: []string{"1.2.3.4"},
				CP:       defaultCloudProperties,
			}, {
				// Database resources
				Parent: defaultInstanceResource,
				CP:     defaultCloudProperties,
			}, {
				// Primary site
				Parent:   defaultInstanceResource,
				HostList: []string{"primary-site"},
				CP:       defaultCloudProperties,
			}, {
				// Secondary site
				Parent:   defaultInstanceResource,
				HostList: []string{"secondary-site"},
				CP:       defaultCloudProperties,
			}},
		},
		testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
			DiscoverCurrentHostResp: []hostdiscovery.HostData{{
				NFSAddrs: []string{"1.2.3.4"},
			}},
		},
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  secondaryInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}},
				HostProject: "12345",
				ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Sid: "ABC",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  defaultInstanceURI,
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								InstanceRole: syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
							},
						}, {
							ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_BACKEND_SERVICE,
							ResourceUri:      "some-backend-service-uri",
							RelatedResources: []string{"primary-instance-group"},
						}, {
							ResourceType:     syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceKind:     syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE_GROUP,
							ResourceUri:      "primary-instance-group",
							RelatedResources: []string{defaultInstanceURI},
						}, {
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_STORAGE,
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_FILESTORE,
							ResourceUri:  "some-shared-nfs-uri",
						}},
					}},
				},
			},
			ProjectNumber: "12345",
		}},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := &Discovery{
				SapDiscoveryInterface:   test.testSapDiscovery,
				CloudDiscoveryInterface: test.testCloudDiscovery,
				HostDiscoveryInterface:  test.testHostDiscovery,
				OSStatReader: func(string) (os.FileInfo, error) {
					return nil, errors.New("No file")
				},
			}
			if test.testOSStatReader != nil {
				d.OSStatReader = test.testOSStatReader
			}
			if test.testFileReader != nil {
				d.FileReader = test.testFileReader
			}
			got := d.discoverSAPSystems(context.Background(), defaultCloudProperties, test.config)
			t.Logf("Got systems: %+v ", got)
			t.Logf("Want systems: %+v ", test.want)
			if diff := cmp.Diff(test.want, got, append(resourceListDiffOpts, protocmp.IgnoreFields(&syspb.SapDiscovery{}, "update_time"))...); diff != "" {
				t.Errorf("discoverSAPSystems() mismatch (-want, +got):\n%s", diff)
			}
			if len(test.testCloudDiscovery.DiscoverComputeResourcesArgsDiffs) != 0 {
				for _, diff := range test.testCloudDiscovery.DiscoverComputeResourcesArgsDiffs {
					t.Errorf("discoverSAPSystems() discoverCloudResourcesArgs mismatch (-want, +got):\n%s", diff)
				}
			}
		})
	}
}

func TestWriteToCloudLogging(t *testing.T) {
	tests := []struct {
		name         string
		system       *syspb.SapDiscovery
		logInterface *logfake.TestCloudLogging
	}{{
		name:   "writeEmptySystem",
		system: &syspb.SapDiscovery{},
		logInterface: &logfake.TestCloudLogging{
			ExpectedLogEntries: []logging.Entry{{
				Severity: logging.Info,
				Payload:  map[string]string{"type": "SapDiscovery", "discovery": ""},
			}},
		},
	}, {
		name: "writeFullSystem",
		system: &syspb.SapDiscovery{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid:         "APP",
				HostProject: "test/project",
				Resources: []*syspb.SapDiscovery_Resource{
					{ResourceUri: "some/compute/instance", ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE, ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE},
				},
			},
			DatabaseLayer: &syspb.SapDiscovery_Component{Sid: "DAT", HostProject: "test/project", Resources: []*syspb.SapDiscovery_Resource{
				{ResourceUri: "some/compute/instance", ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE, ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE},
			},
			},
		},
		logInterface: &logfake.TestCloudLogging{
			ExpectedLogEntries: []logging.Entry{{
				Severity: logging.Info,
				Payload:  map[string]string{"type": "SapDiscovery", "discovery": ""},
			}},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.logInterface.T = t
			d := Discovery{
				CloudLogInterface: test.logInterface,
			}
			d.writeToCloudLogging(test.system)
		})
	}
}

func TestUpdateSAPInstances(t *testing.T) {
	tests := []struct {
		name              string
		config            *cpb.Configuration
		discoverResponses []*sappb.SAPInstances
		wantInstances     []*sappb.SAPInstances // An array to test asynchronous update functionality
	}{{
		name: "singleUpdate",
		config: &cpb.Configuration{DiscoveryConfiguration: &cpb.DiscoveryConfiguration{
			SapInstancesUpdateFrequency: &dpb.Duration{Seconds: 5},
		}},
		discoverResponses: []*sappb.SAPInstances{{Instances: []*sappb.SAPInstance{{
			Sapsid: "abc",
		}}}},
		wantInstances: []*sappb.SAPInstances{{Instances: []*sappb.SAPInstance{{
			Sapsid: "abc",
		}}}},
	}, {
		name: "multipleUpdates",
		config: &cpb.Configuration{DiscoveryConfiguration: &cpb.DiscoveryConfiguration{
			SapInstancesUpdateFrequency: &dpb.Duration{Seconds: 5},
		}},
		discoverResponses: []*sappb.SAPInstances{{Instances: []*sappb.SAPInstance{{
			Sapsid: "abc",
		}}}, {Instances: []*sappb.SAPInstance{{
			Sapsid: "def",
		}}}},
		wantInstances: []*sappb.SAPInstances{{Instances: []*sappb.SAPInstance{{
			Sapsid: "abc",
		}}}, {Instances: []*sappb.SAPInstance{{
			Sapsid: "def",
		}}}},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			discoverCalls := 0
			d := &Discovery{
				AppsDiscovery: func(context.Context, SapSystemDiscoveryInterface) *sappb.SAPInstances {
					defer func() {
						discoverCalls++
					}()
					return test.discoverResponses[discoverCalls]
				},
				OSStatReader: func(string) (os.FileInfo, error) {
					return nil, errors.New("No file")
				},
			}
			ctx, cancel := context.WithCancel(context.Background())
			go updateSAPInstances(ctx, updateSapInstancesArgs{d: d, config: test.config})
			var oldInstances *sappb.SAPInstances
			for _, want := range test.wantInstances {
				// Wait the update time
				log.CtxLogger(ctx).Info("Checking updated instances")
				var got *sappb.SAPInstances
				for {
					got = d.GetSAPInstances()
					if got != nil && (oldInstances == nil || got != oldInstances) {
						oldInstances = got
						break
					}
					time.Sleep(test.config.GetDiscoveryConfiguration().GetSapInstancesUpdateFrequency().AsDuration() / 2)
				}
				if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
					t.Errorf("updateSAPInstances() mismatch (-want, +got):\n%s", diff)
				}
			}
			cancel()
		})
	}
}

func TestRunDiscovery(t *testing.T) {
	tests := []struct {
		name               string
		config             *cpb.Configuration
		testLog            *logfake.TestCloudLogging
		testSapDiscovery   *appsdiscoveryfake.SapDiscovery
		testCloudDiscovery *clouddiscoveryfake.CloudDiscovery
		testHostDiscovery  *hostdiscoveryfake.HostDiscovery
		testWLM            *wlmfake.TestWLM
		wantSystems        [][]*syspb.SapDiscovery
	}{
		{
			name: "disableWrite",
			config: &cpb.Configuration{
				CloudProperties: defaultCloudProperties,
				DiscoveryConfiguration: &cpb.DiscoveryConfiguration{
					EnableDiscovery:                &wpb.BoolValue{Value: false},
					SystemDiscoveryUpdateFrequency: &dpb.Duration{Seconds: 5},
				},
			},
			testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
				DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
					AppComponent: &syspb.SapDiscovery_Component{Sid: "ABC"},
					DBComponent:  &syspb.SapDiscovery_Component{Sid: "DEF"},
				}}},
			},
			testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
				DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{defaultInstanceResource}, {}, {}, {}},
			},
			testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
				DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
			},
			testLog: &logfake.TestCloudLogging{
				ExpectedLogEntries: []logging.Entry{},
			},
			testWLM: &wlmfake.TestWLM{
				WriteInsightArgs: []wlmfake.WriteInsightArgs{},
				WriteInsightErrs: []error{nil},
			},
			wantSystems: [][]*syspb.SapDiscovery{{{
				ApplicationLayer: &syspb.SapDiscovery_Component{
					Sid:         "ABC",
					HostProject: "12345",
				},
				DatabaseLayer: &syspb.SapDiscovery_Component{
					Sid:         "DEF",
					HostProject: "12345",
				},
				ProjectNumber: "12345",
			}}},
		},
		{
			name: "singleUpdate",
			config: &cpb.Configuration{
				CloudProperties: defaultCloudProperties,
				DiscoveryConfiguration: &cpb.DiscoveryConfiguration{
					EnableDiscovery:                &wpb.BoolValue{Value: true},
					SystemDiscoveryUpdateFrequency: &dpb.Duration{Seconds: 5},
				},
			},
			testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
				DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
					AppComponent: &syspb.SapDiscovery_Component{Sid: "ABC"},
					DBComponent:  &syspb.SapDiscovery_Component{Sid: "DEF"},
				}}},
			},
			testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
				DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{defaultInstanceResource}, {}, {}, {}},
			},
			testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
				DiscoverCurrentHostResp: []hostdiscovery.HostData{{}},
			},
			testLog: &logfake.TestCloudLogging{
				ExpectedLogEntries: []logging.Entry{{
					Severity: logging.Info,
					Payload:  map[string]string{"type": "SapDiscovery", "discovery": ""},
				}},
			},
			testWLM: &wlmfake.TestWLM{
				WriteInsightArgs: []wlmfake.WriteInsightArgs{{
					Project:  "test-project-id",
					Location: "test-zone",
					Req: &dwpb.WriteInsightRequest{
						Insight: &dwpb.Insight{
							SapDiscovery: &syspb.SapDiscovery{
								ApplicationLayer: &syspb.SapDiscovery_Component{
									Sid:         "ABC",
									HostProject: "12345",
								},
								DatabaseLayer: &syspb.SapDiscovery_Component{
									Sid:         "DEF",
									HostProject: "12345",
								},
								ProjectNumber: "12345",
							},
						},
						AgentVersion: configuration.AgentVersion,
					},
				}},
				WriteInsightErrs: []error{nil},
			},
			wantSystems: [][]*syspb.SapDiscovery{{{
				ApplicationLayer: &syspb.SapDiscovery_Component{
					Sid:         "ABC",
					HostProject: "12345",
				},
				DatabaseLayer: &syspb.SapDiscovery_Component{
					Sid:         "DEF",
					HostProject: "12345",
				},
				ProjectNumber: "12345",
			}}},
		},
		{
			name: "multipleUpdates",
			config: &cpb.Configuration{
				CloudProperties: defaultCloudProperties,
				DiscoveryConfiguration: &cpb.DiscoveryConfiguration{
					EnableDiscovery:                &wpb.BoolValue{Value: true},
					SystemDiscoveryUpdateFrequency: &dpb.Duration{Seconds: 5},
				},
			},
			testSapDiscovery: &appsdiscoveryfake.SapDiscovery{
				DiscoverSapAppsResp: [][]appsdiscovery.SapSystemDetails{{{
					AppComponent: &syspb.SapDiscovery_Component{Sid: "ABC"},
					DBComponent:  &syspb.SapDiscovery_Component{Sid: "DEF"},
				}}, {{
					AppComponent: &syspb.SapDiscovery_Component{Sid: "GHI"},
					DBComponent:  &syspb.SapDiscovery_Component{Sid: "JKL"},
				}}},
			},
			testCloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
				DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{defaultInstanceResource}, {}, {}, {}, {defaultInstanceResource}, {}, {}, {}},
			},
			testHostDiscovery: &hostdiscoveryfake.HostDiscovery{
				DiscoverCurrentHostResp: []hostdiscovery.HostData{{}, {}},
			},
			testLog: &logfake.TestCloudLogging{
				ExpectedLogEntries: []logging.Entry{{
					Severity: logging.Info,
					Payload:  map[string]string{"type": "SapDiscovery", "discovery": ""},
				}, {
					Severity: logging.Info,
					Payload:  map[string]string{"type": "SapDiscovery", "discovery": ""},
				}},
			},
			testWLM: &wlmfake.TestWLM{
				WriteInsightArgs: []wlmfake.WriteInsightArgs{{
					Project:  "test-project-id",
					Location: "test-zone",
					Req: &dwpb.WriteInsightRequest{
						Insight: &dwpb.Insight{
							SapDiscovery: &syspb.SapDiscovery{
								ApplicationLayer: &syspb.SapDiscovery_Component{
									Sid:         "ABC",
									HostProject: "12345",
								},
								DatabaseLayer: &syspb.SapDiscovery_Component{
									Sid:         "DEF",
									HostProject: "12345",
								},
								ProjectNumber: "12345",
							},
						},
						AgentVersion: configuration.AgentVersion,
					},
				}, {
					Project:  "test-project-id",
					Location: "test-zone",
					Req: &dwpb.WriteInsightRequest{
						Insight: &dwpb.Insight{
							SapDiscovery: &syspb.SapDiscovery{
								ApplicationLayer: &syspb.SapDiscovery_Component{
									Sid:         "GHI",
									HostProject: "12345",
								},
								DatabaseLayer: &syspb.SapDiscovery_Component{
									Sid:         "JKL",
									HostProject: "12345",
								},
								ProjectNumber: "12345",
							},
						},
						AgentVersion: configuration.AgentVersion,
					},
				}},
				WriteInsightErrs: []error{nil, nil},
			},
			wantSystems: [][]*syspb.SapDiscovery{{{
				ApplicationLayer: &syspb.SapDiscovery_Component{
					Sid:         "ABC",
					HostProject: "12345",
				},
				DatabaseLayer: &syspb.SapDiscovery_Component{
					Sid:         "DEF",
					HostProject: "12345",
				},
				ProjectNumber: "12345",
			}}, {{
				ApplicationLayer: &syspb.SapDiscovery_Component{
					Sid:         "GHI",
					HostProject: "12345",
				},
				DatabaseLayer: &syspb.SapDiscovery_Component{
					Sid:         "JKL",
					HostProject: "12345",
				},
				ProjectNumber: "12345",
			}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.testLog.T = t
			test.testWLM.T = t
			d := &Discovery{
				WlmService:              test.testWLM,
				CloudLogInterface:       test.testLog,
				SapDiscoveryInterface:   test.testSapDiscovery,
				CloudDiscoveryInterface: test.testCloudDiscovery,
				HostDiscoveryInterface:  test.testHostDiscovery,
				OSStatReader: func(string) (os.FileInfo, error) {
					return nil, errors.New("No file")
				},
			}
			ctx, cancel := context.WithCancel(context.Background())
			go runDiscovery(ctx, runDiscoveryArgs{config: test.config, d: d})

			var oldSystems []*syspb.SapDiscovery
			for _, want := range test.wantSystems {
				log.CtxLogger(ctx).Info("Checking updated instances")
				var got []*syspb.SapDiscovery
				for {
					got = d.GetSAPSystems()
					if got != nil && (oldSystems == nil || got[0].GetUpdateTime() != oldSystems[0].GetUpdateTime()) {
						// Got something different, compare with wanted
						oldSystems = got
						break
					}
					// Wait half the refresh interval and check for an update again
					time.Sleep(test.config.GetDiscoveryConfiguration().GetSystemDiscoveryUpdateFrequency().AsDuration() / 2)
				}
				if diff := cmp.Diff(want, got, append(resourceListDiffOpts, protocmp.IgnoreFields(&syspb.SapDiscovery{}, "update_time"))...); diff != "" {
					t.Errorf("runDiscovery() mismatch (-want, +got):\n%s", diff)
				}
			}
			cancel()
		})
	}
}

type fakeReadCloser struct {
	fileContents string
	readError    error
	bytesRead    int
}

func (f fakeReadCloser) Read(p []byte) (n int, err error) {
	if f.readError != nil {
		return 0, f.readError
	}
	log.Logger.Infof("Reading from string %s", f.fileContents)
	bytesLeft := len(f.fileContents) - f.bytesRead
	log.Logger.Infof("bytesLeft: %d", bytesLeft)
	bytesToRead := min(len(p), bytesLeft)
	log.Logger.Infof("bytesToRead: %d", bytesToRead)
	copy(p, []byte(f.fileContents[f.bytesRead:f.bytesRead+bytesToRead]))
	log.Logger.Infof("p: %s", string(p))
	f.bytesRead += bytesToRead
	log.Logger.Infof("f.bytesRead: %d", f.bytesRead)
	if f.bytesRead == len(f.fileContents) {
		return bytesToRead, io.EOF
	}
	return bytesToRead, nil
}
func (f fakeReadCloser) Close() error {
	return nil
}

func TestDiscoverOverrideSystem(t *testing.T) {
	tests := []struct {
		name          string
		fileContents  string
		fileOpenErr   error
		fileReadError error
		instance      *syspb.SapDiscovery_Resource
		want          []*syspb.SapDiscovery
	}{{
		name: "success",
		fileContents: `{
			"databaseLayer": {
				"hostProject": "12345",
				"sid": "DEF"
			},
			"applicationLayer": {
				"hostProject": "12345",
				"sid": "ABC"
			},
			"projectNumber": "12345"
		}`,
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid:         "ABC",
				HostProject: "12345",
			},
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid:         "DEF",
				HostProject: "12345",
			},
			ProjectNumber: "12345",
		}},
	}, {
		name:        "readerError",
		fileOpenErr: errors.New("some error"),
		want:        nil,
	}, {
		name: "readAllError",
		fileContents: `{
			"databaseLayer": {
				"hostProject": "12345",
				"sid": "DEF"
			},
			"applicationLayer": {
				"hostProject": "12345",
				"sid": "ABC"
			},
			"projectNumber": "12345"
		}`,
		fileReadError: errors.New("some error"),
		want:          nil,
	}, {
		name:         "jsonError",
		fileContents: "not json",
		want:         nil,
	}, {
		name: "overwritesDatabaseInstance",
		fileContents: `{
			"databaseLayer": {
				"resources": [{
					"resourceType": "RESOURCE_TYPE_COMPUTE",
					"resourceKind": "RESOURCE_KIND_INSTANCE"
				}],
				"hostProject": "12345",
				"sid": "DEF"
			},
			"projectNumber": "12345"
		}`,
		instance: defaultInstanceResource,
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid:         "DEF",
				HostProject: "12345",
				Resources:   []*syspb.SapDiscovery_Resource{defaultInstanceResource},
			},
			ProjectNumber: "12345",
		}},
	}, {
		name: "overwritesApplicationInstance",
		fileContents: `{
			"applicationLayer": {
				"resources": [{
					"resourceType": "RESOURCE_TYPE_COMPUTE",
					"resourceKind": "RESOURCE_KIND_INSTANCE"
				}],
				"hostProject": "12345",
				"sid": "DEF"
			},
			"projectNumber": "12345"
		}`,
		instance: defaultInstanceResource,
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid:         "DEF",
				HostProject: "12345",
				Resources:   []*syspb.SapDiscovery_Resource{defaultInstanceResource},
			},
			ProjectNumber: "12345",
		}},
	}, {
		name: "doesntOverwiteDatabaseInstanceWithURI",
		fileContents: `{
			"databaseLayer": {
				"resources": [{
					"resourceType": "RESOURCE_TYPE_COMPUTE",
					"resourceKind": "RESOURCE_KIND_INSTANCE",
					"resourceUri": "some-uri"
				}],
				"hostProject": "12345",
				"sid": "DEF"
			},
			"projectNumber": "12345"
		}`,
		instance: defaultInstanceResource,
		want: []*syspb.SapDiscovery{{
			DatabaseLayer: &syspb.SapDiscovery_Component{
				Sid:         "DEF",
				HostProject: "12345",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some-uri",
				}},
			},
			ProjectNumber: "12345",
		}},
	}, {
		name: "doesntOverwiteApplicationInstanceWithURI",
		fileContents: `{
			"applicationLayer": {
				"resources": [{
					"resourceType": "RESOURCE_TYPE_COMPUTE",
					"resourceKind": "RESOURCE_KIND_INSTANCE",
					"resourceUri": "some-uri"
				}],
				"hostProject": "12345",
				"sid": "DEF"
			},
			"projectNumber": "12345"
		}`,
		instance: defaultInstanceResource,
		want: []*syspb.SapDiscovery{{
			ApplicationLayer: &syspb.SapDiscovery_Component{
				Sid:         "DEF",
				HostProject: "12345",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceUri:  "some-uri",
				}},
			},
			ProjectNumber: "12345",
		}},
	}}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := &Discovery{
				FileReader: func(string) (io.ReadCloser, error) {
					return fakeReadCloser{fileContents: tc.fileContents, readError: tc.fileReadError}, tc.fileOpenErr
				},
			}
			got := d.discoverOverrideSystem(ctx, "overrideFile", tc.instance)
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("discoverOverrideSystem() returned an unexpected diff (-want +got): %v", diff)
			}
		})
	}
}

func TestDiscoverReplicationSite(t *testing.T) {
	tests := []struct {
		name           string
		site           *sappb.HANAReplicaSite
		lbGroups       []loadBalancerGroup
		cloudDiscovery *clouddiscoveryfake.CloudDiscovery
		want           *syspb.SapDiscovery_Component_ReplicationSite
	}{{
		name: "success",
		site: &sappb.HANAReplicaSite{
			Name: "primary-site",
			Targets: []*sappb.HANAReplicaSite{{
				Name: "secondary-site",
			}},
		},
		cloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"primary-site"},
				CP:         defaultCloudProperties,
			}, {
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"secondary-site"},
				CP:         defaultCloudProperties,
			}},
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "primary-site",
				},
			}}, {{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  makeZonalURI(defaultProjectID, "replication-region-a", "instances", "secondary-instance"),
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "secondary-site",
				},
			}}},
		},
		want: &syspb.SapDiscovery_Component_ReplicationSite{
			Component: &syspb.SapDiscovery_Component{
				Region: "test-zone",
				Sid:    "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "primary-site",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}},
				ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Region: "replication-region",
						Sid:    "ABC",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  makeZonalURI(defaultProjectID, "replication-region-a", "instances", "secondary-instance"),
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
								VirtualHostname: "secondary-site",
							},
						}},
					},
				}},
			},
		},
	}, {
		name: "noInstancesInSite",
		site: &sappb.HANAReplicaSite{
			Name: "primary-site",
		},
		cloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"primary-site"},
				CP:         defaultCloudProperties,
			}},
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  makeZonalURI(defaultProjectID, "replication-region-a", "disks", "primary-disk"),
			}}},
		},
		want: &syspb.SapDiscovery_Component_ReplicationSite{
			Component: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_DISK,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  makeZonalURI(defaultProjectID, "replication-region-a", "disks", "primary-disk"),
				}},
			},
		},
	}, {
		name: "instanceHasDifferentVirtualHostname",
		site: &sappb.HANAReplicaSite{
			Name: "primary-site",
		},
		cloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"primary-site"},
				CP:         defaultCloudProperties,
			}},
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "other-hostname",
				},
			}}},
		},
		want: &syspb.SapDiscovery_Component_ReplicationSite{
			Component: &syspb.SapDiscovery_Component{
				Sid: "ABC",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "other-hostname",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}},
			},
		},
	}, {
		name: "onlyUsesRegionForSameHostname",
		site: &sappb.HANAReplicaSite{
			Name: "primary-site",
		},
		cloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"primary-site"},
				CP:         defaultCloudProperties,
			}},
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  makeZonalURI(defaultProjectID, "other-region-a", "instances", "other-instance"),
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "other-hostname",
				},
			}, {
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "primary-site",
				},
			}}},
		},
		want: &syspb.SapDiscovery_Component_ReplicationSite{
			Component: &syspb.SapDiscovery_Component{
				Sid:    "ABC",
				Region: "test-zone",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  makeZonalURI(defaultProjectID, "other-region-a", "instances", "other-instance"),
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "other-hostname",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "primary-site",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}},
			},
		},
	}, {
		name: "replicationSiteIsPartOfPrimaryLBGroup",
		site: &sappb.HANAReplicaSite{
			Name: "primary-site",
			Targets: []*sappb.HANAReplicaSite{{
				Name: "secondary-site",
			}},
		},
		cloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"primary-site"},
				CP:         defaultCloudProperties,
			}, {
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"secondary-site"},
				CP:         defaultCloudProperties,
			}},
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "primary-site",
				},
			}}, {{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  secondaryInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "secondary-site",
				},
			}}},
		},
		lbGroups: []loadBalancerGroup{{
			instanceURIs: []string{defaultInstanceURI, secondaryInstanceURI},
		}},
		want: &syspb.SapDiscovery_Component_ReplicationSite{
			Component: &syspb.SapDiscovery_Component{
				Sid:    "ABC",
				Region: "test-zone",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "primary-site",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}, {
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  secondaryInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "secondary-site",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}},
				HaHosts: []string{defaultInstanceURI, secondaryInstanceURI},
			},
		},
	}, {
		name: "replicationSiteIsPartOfDifferentLBGroup",
		site: &sappb.HANAReplicaSite{
			Name: "primary-site",
			Targets: []*sappb.HANAReplicaSite{{
				Name: "secondary-site",
			}},
		},
		cloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"primary-site"},
				CP:         defaultCloudProperties,
			}, {
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"secondary-site"},
				CP:         defaultCloudProperties,
			}},
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "primary-site",
				},
			}}, {{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  secondaryInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "secondary-site",
				},
			}}},
		},
		lbGroups: []loadBalancerGroup{{
			instanceURIs: []string{defaultInstanceURI},
		}, {
			instanceURIs: []string{secondaryInstanceURI},
		}},
		want: &syspb.SapDiscovery_Component_ReplicationSite{
			Component: &syspb.SapDiscovery_Component{
				Sid:    "ABC",
				Region: "test-zone",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "primary-site",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}},
				ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Sid:    "ABC",
						Region: "test-zone",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  secondaryInstanceURI,
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								VirtualHostname: "secondary-site",
								InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
							},
						}},
					},
				}},
			},
		},
	}, {
		name: "replicationSiteIsPartOfDifferentLBGroup",
		site: &sappb.HANAReplicaSite{
			Name: "primary-site",
			Targets: []*sappb.HANAReplicaSite{{
				Name: "secondary-site",
			}},
		},
		cloudDiscovery: &clouddiscoveryfake.CloudDiscovery{
			DiscoverComputeResourcesArgs: []clouddiscoveryfake.DiscoverComputeResourcesArgs{{
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"primary-site"},
				CP:         defaultCloudProperties,
			}, {
				Parent:     defaultInstanceResource,
				Subnetwork: defaultSubnetwork,
				HostList:   []string{"secondary-site"},
				CP:         defaultCloudProperties,
			}},
			DiscoverComputeResourcesResp: [][]*syspb.SapDiscovery_Resource{{{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  defaultInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "primary-site",
				},
			}}, {{
				ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
				ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
				ResourceUri:  secondaryInstanceURI,
				InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
					VirtualHostname: "secondary-site",
				},
			}}},
		},
		lbGroups: []loadBalancerGroup{{
			instanceURIs: []string{defaultInstanceURI},
		}, {
			instanceURIs: []string{secondaryInstanceURI},
		}},
		want: &syspb.SapDiscovery_Component_ReplicationSite{
			Component: &syspb.SapDiscovery_Component{
				Sid:    "ABC",
				Region: "test-zone",
				Resources: []*syspb.SapDiscovery_Resource{{
					ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
					ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
					ResourceUri:  defaultInstanceURI,
					InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
						VirtualHostname: "primary-site",
						InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
					},
				}},
				ReplicationSites: []*syspb.SapDiscovery_Component_ReplicationSite{{
					SourceSite: "primary-site",
					Component: &syspb.SapDiscovery_Component{
						Sid:    "ABC",
						Region: "test-zone",
						Resources: []*syspb.SapDiscovery_Resource{{
							ResourceKind: syspb.SapDiscovery_Resource_RESOURCE_KIND_INSTANCE,
							ResourceType: syspb.SapDiscovery_Resource_RESOURCE_TYPE_COMPUTE,
							ResourceUri:  secondaryInstanceURI,
							InstanceProperties: &syspb.SapDiscovery_Resource_InstanceProperties{
								VirtualHostname: "secondary-site",
								InstanceRole:    syspb.SapDiscovery_Resource_InstanceProperties_INSTANCE_ROLE_DATABASE,
							},
						}},
					},
				}},
			},
		},
	}}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := &Discovery{
				CloudDiscoveryInterface: tc.cloudDiscovery,
			}
			got := d.discoverReplicationSite(ctx, tc.site, defaultSID, defaultInstanceResource, defaultSubnetwork, tc.lbGroups, defaultCloudProperties)
			if diff := cmp.Diff(tc.want, got, resourceListDiffOpts...); diff != "" {
				t.Errorf("discoverReplicationSite() returned an unexpected diff (-want +got): %v", diff)
			}
		})
	}
}
