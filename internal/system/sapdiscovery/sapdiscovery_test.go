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

package sapdiscovery

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/GoogleCloudPlatform/sapagent/shared/commandlineexecutor"
	"github.com/GoogleCloudPlatform/sapagent/shared/gce/fake"

	cpb "github.com/GoogleCloudPlatform/sapagent/protos/configuration"
	sapb "github.com/GoogleCloudPlatform/sapagent/protos/sapapp"
)

var (
	sapInitRunningOutput = `
	saphostexec running (pid = 3640)
	sapstartsrv running (pid = 3958)
	saposcol running (pid = 4031)
	pid's (3958 3960 4229 4537)
	running`

	sapInitStoppedOutput = `
	saphostexec stopped
	No process running.`
)

func TestInstances(t *testing.T) {
	tests := []struct {
		name                  string
		fakeReplicationConfig replicationConfig
		fakeList              listInstances
		fakeExec              commandlineexecutor.Execute
		want                  *sapb.SAPInstances
	}{
		{
			name: "NOSAPInstances",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return nil, nil
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			want: &sapb.SAPInstances{},
		},
		{
			name: "SingleHANAStandaloneInstance",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "HDB",
						Snr:           "00",
						InstanceName:  "HDB",
						ProfilePath:   "/usr/sap/HDB/SYS/profile/HDB_HDB00_vm1",
						LDLibraryPath: "/usr/sap/HDB/SYS/exe",
					},
				}, nil
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeReplicationConfig: func(ctx context.Context, user string, sid string, instanceID string) (int, []string, int64, error) {
				return 0, []string{"gce-1", "gce-2"}, 10, nil
			},
			want: &sapb.SAPInstances{
				Instances: []*sapb.SAPInstance{&sapb.SAPInstance{
					Sapsid:         "HDB",
					InstanceNumber: "00",
					Type:           sapb.InstanceType_HANA,
					Site:           sapb.InstanceSite_HANA_STANDALONE,
					HanaHaMembers:  []string{"gce-1", "gce-2"},
					SapcontrolPath: "/usr/sap/HDB/SYS/exe/sapcontrol",
					User:           "hdbadm",
					InstanceId:     "HDB00",
					ProfilePath:    "/usr/sap/HDB/SYS/profile/HDB_HDB00_vm1",
					LdLibraryPath:  "/usr/sap/HDB/SYS/exe",
				}},
			},
		},
		{
			name: "HANAPrimaryInstance",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "HDB",
						Snr:           "00",
						InstanceName:  "HDB",
						ProfilePath:   "/usr/sap/HDB/SYS/profile/HDB_HDB00_vm1",
						LDLibraryPath: "/usr/sap/HDB/SYS/exe",
					},
				}, nil
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeReplicationConfig: func(ctx context.Context, user string, sid string, instanceID string) (int, []string, int64, error) {
				return 1, []string{"gce-1", "gce-2"}, 15, nil
			},
			want: &sapb.SAPInstances{
				Instances: []*sapb.SAPInstance{&sapb.SAPInstance{
					Sapsid:         "HDB",
					InstanceNumber: "00",
					Type:           sapb.InstanceType_HANA,
					Site:           sapb.InstanceSite_HANA_PRIMARY,
					HanaHaMembers:  []string{"gce-1", "gce-2"},
					SapcontrolPath: "/usr/sap/HDB/SYS/exe/sapcontrol",
					User:           "hdbadm",
					InstanceId:     "HDB00",
					ProfilePath:    "/usr/sap/HDB/SYS/profile/HDB_HDB00_vm1",
					LdLibraryPath:  "/usr/sap/HDB/SYS/exe",
				}},
			},
		},
		{
			name: "HANAPathHeterogeneous",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "HSE",
						Snr:           "00",
						InstanceName:  "HDB",
						ProfilePath:   "/usr/sap/HSE/SYS/profile/HDB_HDB00_vm1",
						LDLibraryPath: "/usr/sap/HSE/SYS/exe",
					},
				}, nil
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeReplicationConfig: func(ctx context.Context, user string, sid string, instanceID string) (int, []string, int64, error) {
				return 1, []string{"gce-1", "gce-2"}, 15, nil
			},
			want: &sapb.SAPInstances{
				Instances: []*sapb.SAPInstance{&sapb.SAPInstance{
					Sapsid:         "HSE",
					InstanceNumber: "00",
					Type:           sapb.InstanceType_HANA,
					Site:           sapb.InstanceSite_HANA_PRIMARY,
					HanaHaMembers:  []string{"gce-1", "gce-2"},
					SapcontrolPath: "/usr/sap/HSE/SYS/exe/sapcontrol",
					User:           "hseadm",
					InstanceId:     "HDB00",
					ProfilePath:    "/usr/sap/HSE/SYS/profile/HDB_HDB00_vm1",
					LdLibraryPath:  "/usr/sap/HSE/SYS/exe",
				}},
			},
		},
		{
			name: "NetweaverInstance",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "DEV",
						Snr:           "00",
						InstanceName:  "ASCS",
						ProfilePath:   "/usr/sap/DEV/SYS/profile/ASCS_ASCS00_vm1",
						LDLibraryPath: "/usr/sap/DEV/SYS/exe",
					},
				}, nil
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			want: &sapb.SAPInstances{
				Instances: []*sapb.SAPInstance{
					&sapb.SAPInstance{
						Sapsid:                  "DEV",
						InstanceNumber:          "00",
						Type:                    sapb.InstanceType_NETWEAVER,
						NetweaverHttpPort:       "8100",
						SapcontrolPath:          "/usr/sap/DEV/SYS/exe/sapcontrol",
						User:                    "devadm",
						InstanceId:              "ASCS00",
						Kind:                    sapb.InstanceKind_CS,
						ProfilePath:             "/usr/sap/DEV/SYS/profile/ASCS_ASCS00_vm1",
						LdLibraryPath:           "/usr/sap/DEV/SYS/exe",
						NetweaverHealthCheckUrl: "http://localhost:8100/msgserver/text/logon",
						ServiceName:             "SAP-CS",
					},
				},
			},
		},
		{
			name: "FileReadFailure",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return nil, cmpopts.AnyError
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			want: &sapb.SAPInstances{},
		},
		{
			name: "ReadReplicationConfigFailure",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "HSE",
						Snr:           "00",
						InstanceName:  "HDB",
						ProfilePath:   "/usr/sap/HSE/SYS/profile/HDB_HDB00_vm1",
						LDLibraryPath: "/usr/sap/HSE/SYS/exe",
					},
				}, nil
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeReplicationConfig: func(ctx context.Context, user string, sid string, instanceID string) (int, []string, int64, error) {
				return 0, nil, 0, cmpopts.AnyError
			},
			want: &sapb.SAPInstances{
				Instances: []*sapb.SAPInstance{&sapb.SAPInstance{
					Sapsid:         "HSE",
					InstanceId:     "HDB00",
					User:           "hseadm",
					InstanceNumber: "00",
					Type:           sapb.InstanceType_HANA,
					Site:           sapb.InstanceSite_INSTANCE_SITE_UNDEFINED,
					LdLibraryPath:  "/usr/sap/HSE/SYS/exe",
					SapcontrolPath: "/usr/sap/HSE/SYS/exe/sapcontrol",
					ProfilePath:    "/usr/sap/HSE/SYS/profile/HDB_HDB00_vm1",
				}},
			},
		},
		{
			name: "HANASiteUndefined",
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "HDB",
						Snr:           "00",
						InstanceName:  "HDB",
						ProfilePath:   "/usr/sap/HDB/SYS/profile/HDB_HDB00_vm1",
						LDLibraryPath: "/usr/sap/HDB/SYS/exe",
					},
				}, nil
			},
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeReplicationConfig: func(ctx context.Context, user string, sid string, instanceID string) (int, []string, int64, error) {
				return -1, nil, 0, nil
			},
			want: &sapb.SAPInstances{
				Instances: []*sapb.SAPInstance{&sapb.SAPInstance{
					Sapsid:         "HDB",
					InstanceNumber: "00",
					Type:           sapb.InstanceType_HANA,
					Site:           sapb.InstanceSite_INSTANCE_SITE_UNDEFINED,
					SapcontrolPath: "/usr/sap/HDB/SYS/exe/sapcontrol",
					User:           "hdbadm",
					InstanceId:     "HDB00",
					ProfilePath:    "/usr/sap/HDB/SYS/profile/HDB_HDB00_vm1",
					LdLibraryPath:  "/usr/sap/HDB/SYS/exe",
				}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := instances(context.Background(), test.fakeReplicationConfig, test.fakeList, test.fakeExec, nil)
			if diff := cmp.Diff(test.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("instances() unexpected diff: (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReadReplicationConfig(t *testing.T) {
	tests := []struct {
		name         string
		user         string
		sid          string
		instanceID   string
		fakeExec     commandlineexecutor.Execute
		wantMode     int
		wantHAMebers []string
		wantErr      error
	}{
		{
			name:       "HANAPrimary",
			user:       "hdbadm",
			sid:        "HDB",
			instanceID: "00",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "site/1/REPLICATION_MODE=PRIMARY\nsite/1/SITE_NAME=gce-1\nsite/2/SITE_NAME=gce-2\nlocal_site_id=1\n",
					ExitCode: 15,
				}
			},
			wantMode:     1,
			wantHAMebers: []string{"gce-1", "gce-2"},
			wantErr:      nil,
		},
		{
			name:       "HANASecondary",
			user:       "hdbadm",
			sid:        "HDB",
			instanceID: "00",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "site/1/REPLICATION_MODE=SYNCMEM\nlocal_site_id=1\nsite/1/SITE_NAME=gce-1\nsite/1/PRIMARY_MASTERS=gce-2\n",
					StdErr:   "",
					ExitCode: 15,
				}
			},
			wantMode:     2,
			wantHAMebers: []string{"gce-2", "gce-1"},
			wantErr:      nil,
		},
		{
			name:       "HANAStandalone",
			user:       "hdbadm",
			sid:        "HDB",
			instanceID: "00",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "local_site_id=0\n",
					ExitCode: 10,
				}
			},
			wantMode:     0,
			wantHAMebers: nil,
			wantErr:      nil,
		},
		{
			name:       "HANAPrimarySwappedSites",
			user:       "hdbadm",
			sid:        "HDB",
			instanceID: "00",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "site/1/REPLICATION_MODE=PRIMARY\nsite/2/SITE_NAME=gce-2\nsite/1/SITE_NAME=gce-1\nlocal_site_id=1\n",
					ExitCode: 15,
				}
			},
			wantMode:     1,
			wantHAMebers: []string{"gce-1", "gce-2"},
			wantErr:      nil,
		},
		{
			name: "EmptySiteID",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "local_site_id=\n",
					ExitCode: 15,
				}
			},
			wantMode:     0,
			wantHAMebers: nil,
			wantErr:      cmpopts.AnyError,
		},
		{
			name: "EmptyReplicationMode",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "site/2/REPLICATION_MODE=PRIMARY\nsite/1/SITE_NAME=gce-1\nsite/2/SITE_NAME=gce-2\nlocal_site_id=1\n",
					ExitCode: 15,
				}
			},
			wantMode:     0,
			wantHAMebers: nil,
			wantErr:      cmpopts.AnyError,
		},
		{
			name: "EmptySiteName",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "local_site_id=2\nsite/2/SITE_NAME=",
					ExitCode: 15,
				}
			},
			wantMode:     0,
			wantHAMebers: nil,
			wantErr:      cmpopts.AnyError,
		},
		{
			name: "SiteNamePatternError",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "site/2/REPLICATION_MODE=SYNCMEM\nlocal_site_id=2\nsite/2/SITE",
					ExitCode: 15,
				}
			},
			wantMode:     2,
			wantHAMebers: nil,
			wantErr:      cmpopts.AnyError,
		},
		{
			name: "CmdFailure",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut:   "abc",
					StdErr:   "123",
					ExitCode: 0,
					Error:    cmpopts.AnyError,
				}
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "InvalidStatusCode",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					ExitCode: 99,
				}
			},
			wantErr: cmpopts.AnyError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotMode, gotHAMembers, _, err := readReplicationConfig(context.Background(), test.user, test.sid, test.instanceID, test.fakeExec)

			if !cmp.Equal(err, test.wantErr, cmpopts.EquateErrors()) {
				t.Errorf("readReplicationConfig(%s,%s,%s) error, got: %v want: %v.", test.user, test.sid, test.instanceID, err, test.wantErr)
			}
			if test.wantMode != gotMode {
				t.Errorf("readReplicationConfig(%s,%s,%s) returned incorrect mode, got: %d want: %d.", test.user, test.sid, test.instanceID, gotMode, test.wantMode)
			}
			if cmp.Diff(test.wantHAMebers, gotHAMembers) != "" {
				t.Errorf("readReplicationConfig(%s,%s,%s) returned incorrect haMembers, got: %s want: %s.", test.user, test.sid, test.instanceID, gotHAMembers, test.wantHAMebers)
			}
		})
	}
}

func TestListSAPInstances(t *testing.T) {
	tests := []struct {
		name     string
		fakeExec commandlineexecutor.Execute
		want     []*instanceInfo
		wantErr  error
	}{
		{
			name: "Success",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `LD_LIBRARY_PATH=/usr/sap/DEH/HDB00/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;/usr/sap/DEH/HDB00/exe/sapstartsrv pf=/usr/sap/DEH/SYS/profile/DEH_HDB00_dnwh75ldbci -D -u dehadm
					LD_LIBRARY_PATH=/usr/sap/DEV/ASCS01/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;sapstartsrv pf=/usr/sap/DEV/SYS/profile/DEV_ASCS01_dnwh75ldbci -D -u devadm
					/usr/sap/DEV/D02/exe/sapstartsrv pf=/usr/sap/DEV/SYS/profile/DEV_D02_dnwh75ldbci -D -u devadm`,
				}
			},
			want: []*instanceInfo{
				&instanceInfo{
					Sid:           "DEH",
					Snr:           "00",
					InstanceName:  "HDB",
					ProfilePath:   "/usr/sap/DEH/SYS/profile/DEH_HDB00_dnwh75ldbci",
					LDLibraryPath: "/usr/sap/DEH/HDB00/exe",
				},
				&instanceInfo{
					Sid:           "DEV",
					Snr:           "01",
					InstanceName:  "ASCS",
					ProfilePath:   "/usr/sap/DEV/SYS/profile/DEV_ASCS01_dnwh75ldbci",
					LDLibraryPath: "/usr/sap/DEV/ASCS01/exe",
				},
				&instanceInfo{
					Sid:           "DEV",
					Snr:           "02",
					InstanceName:  "D",
					ProfilePath:   "/usr/sap/DEV/SYS/profile/DEV_D02_dnwh75ldbci",
					LDLibraryPath: "/usr/sap/DEV/D02/exe",
				},
			},
		},
		{
			name: "ERSInstances",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `LD_LIBRARY_PATH=/usr/sap/ED7/ERS12/exe:$LD_LIBRARY_PATH; export LD_LIBRARY_PATH; /usr/sap/ED7/ERS12/exe/sapstartsrv pf=/usr/sap/ED7/SYS/profile/ED7_ERS12_aliders71 -D -u ed7adm
					LD_LIBRARY_PATH=/usr/sap/FD7/ERS22/exe:$LD_LIBRARY_PATH; export LD_LIBRARY_PATH; /usr/sap/FD7/ERS22/exe/sapstartsrv pf=/usr/sap/FD7/ERS22/profile/FD7_ERS22_aliders71 -D -u fd7adm`,
				}
			},
			want: []*instanceInfo{
				&instanceInfo{
					Sid:           "ED7",
					Snr:           "12",
					InstanceName:  "ERS",
					ProfilePath:   "/usr/sap/ED7/SYS/profile/ED7_ERS12_aliders71",
					LDLibraryPath: "/usr/sap/ED7/ERS12/exe",
				},
				&instanceInfo{
					Sid:           "FD7",
					Snr:           "22",
					InstanceName:  "ERS",
					ProfilePath:   "/usr/sap/FD7/ERS22/profile/FD7_ERS22_aliders71",
					LDLibraryPath: "/usr/sap/FD7/ERS22/exe",
				},
			},
		},
		{
			name: "AlphaNumericSID",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `LD_LIBRARY_PATH=/usr/sap/H00/HDB01/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;/usr/sap/H00/HDB01/exe/sapstartsrv pf=/usr/sap/H00/SYS/profile/H00_HDB01_hana-ha-rh81sap-0-u1670561406-primary -D -u h00adm`,
				}
			},
			want: []*instanceInfo{
				&instanceInfo{
					Sid:           "H00",
					Snr:           "01",
					InstanceName:  "HDB",
					ProfilePath:   "/usr/sap/H00/SYS/profile/H00_HDB01_hana-ha-rh81sap-0-u1670561406-primary",
					LDLibraryPath: "/usr/sap/H00/HDB01/exe",
				},
			},
		},
		{
			name: "CommandFailure",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					Error: cmpopts.AnyError,
				}
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "InvalidSapservicesEntry",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: "/Not/SAP/Instance00",
				}
			},
		},
		{
			name: "CommentedLine",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `#LD_LIBRARY_PATH=/usr/sap/DEH/HDB00/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;/usr/sap/DEH/HDB00/exe/sapstartsrv pf=/usr/sap/DEH/SYS/profile/DEH_HDB00_dnwh75ldbci -D -u dehadm
					LD_LIBRARY_PATH=/usr/sap/DEV/ASCS01/exe:$LD_LIBRARY_PATH;export LD_LIBRARY_PATH;sapstartsrv pf=/usr/sap/DEV/SYS/profile/DEV_ASCS01_dnwh75ldbci -D -u devadm
					/usr/sap/DEV/D02/exe/sapstartsrv pf=/usr/sap/DEV/SYS/profile/DEV_D02_dnwh75ldbci -D -u devadm`,
				}
			},
			want: []*instanceInfo{
				&instanceInfo{
					Sid:           "DEV",
					Snr:           "01",
					InstanceName:  "ASCS",
					ProfilePath:   "/usr/sap/DEV/SYS/profile/DEV_ASCS01_dnwh75ldbci",
					LDLibraryPath: "/usr/sap/DEV/ASCS01/exe",
				},
				&instanceInfo{
					Sid:           "DEV",
					Snr:           "02",
					InstanceName:  "D",
					ProfilePath:   "/usr/sap/DEV/SYS/profile/DEV_D02_dnwh75ldbci",
					LDLibraryPath: "/usr/sap/DEV/D02/exe",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := listSAPInstances(context.Background(), test.fakeExec)
			if !cmp.Equal(err, test.wantErr, cmpopts.EquateErrors()) {
				t.Fatalf("listSAPInstances() = %v, want %v", err, test.wantErr)
			}
			// Slice are unordered and comparison can lead to test flakes - Sort the slices.
			diff := cmp.Diff(test.want, got, cmpopts.SortSlices(func(a, b *instanceInfo) bool { return a.Snr < b.Snr }))
			if diff != "" {
				t.Errorf("listSAPInstances() unexpected diff: (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNetweaverInstances(t *testing.T) {
	tests := []struct {
		name     string
		fakeList listInstances
		fakeExec commandlineexecutor.Execute
		want     []*sapb.SAPInstance
		wantErr  error
	}{
		{
			name: "Success",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "DEV",
						Snr:           "00",
						InstanceName:  "ASCS",
						LDLibraryPath: "/usr/sap/DEV/ASCS00/exe",
					},
				}, nil
			},
			want: []*sapb.SAPInstance{
				&sapb.SAPInstance{
					Sapsid:                  "DEV",
					InstanceNumber:          "00",
					User:                    "devadm",
					InstanceId:              "ASCS00",
					Kind:                    sapb.InstanceKind_CS,
					Type:                    sapb.InstanceType_NETWEAVER,
					SapcontrolPath:          "/usr/sap/DEV/ASCS00/exe/sapcontrol",
					NetweaverHttpPort:       "8100",
					LdLibraryPath:           "/usr/sap/DEV/ASCS00/exe",
					NetweaverHealthCheckUrl: "http://localhost:8100/msgserver/text/logon",
					ServiceName:             "SAP-CS",
				},
			},
		},
		{
			name: "listInstanceFailure",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return nil, cmpopts.AnyError
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "HANAInstance",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:          "HDB",
						Snr:          "00",
						InstanceName: "HDB",
					}}, nil
			},
		},
		{
			name: "ICMInstance",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
			fakeList: func(context.Context, commandlineexecutor.Execute) ([]*instanceInfo, error) {
				return []*instanceInfo{
					{
						Sid:           "PTS",
						Snr:           "00",
						InstanceName:  "D",
						LDLibraryPath: "/usr/sap/PTS/D00/exe",
					},
				}, nil
			},
			want: []*sapb.SAPInstance{
				&sapb.SAPInstance{
					Sapsid:                  "PTS",
					InstanceNumber:          "00",
					User:                    "ptsadm",
					InstanceId:              "D00",
					Kind:                    sapb.InstanceKind_APP,
					Type:                    sapb.InstanceType_NETWEAVER,
					SapcontrolPath:          "/usr/sap/PTS/D00/exe/sapcontrol",
					LdLibraryPath:           "/usr/sap/PTS/D00/exe",
					NetweaverHttpPort:       "50000",
					NetweaverHealthCheckUrl: "http://localhost:50000/sap/public/icman/ping",
					ServiceName:             "SAP-ICM-ABAP",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := netweaverInstances(context.Background(), test.fakeList, test.fakeExec)

			if !cmp.Equal(err, test.wantErr, cmpopts.EquateErrors()) {
				t.Fatalf("Unexpected return from netweaverInstances(), got(%v), want(%v)", err, test.wantErr)
			}

			if diff := cmp.Diff(test.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("netweaverInstances() returned unexpected difference in protobuf (-want +got):\n%s.", diff)
			}
		})
	}
}

func TestSAPInitRunning(t *testing.T) {
	tests := []struct {
		name     string
		fakeExec commandlineexecutor.Execute
		want     bool
		wantErr  error
	}{
		{
			name: "sapinitRunning",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: sapInitRunningOutput,
				}
			},
			want: true,
		},
		{
			name: "sapInitStopped",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: sapInitStoppedOutput,
				}
			},
		},
		{
			name: "CmdFailure",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					Error: cmpopts.AnyError,
				}
			},
			wantErr: cmpopts.AnyError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := sapInitRunning(context.Background(), test.fakeExec)

			if !cmp.Equal(gotErr, test.wantErr, cmpopts.EquateErrors()) {
				t.Errorf("Unexpected return from sapInitRunning(), got(%v), want(%v)", got, test.wantErr)
			}
			if got != test.want {
				t.Errorf("sapInitRunning() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestFindPort(t *testing.T) {
	tests := []struct {
		name         string
		inst         *sapb.SAPInstance
		instanceName string
		wantPort     string
		wantType     sapb.InstanceType
		wantKind     sapb.InstanceKind
		fakeExec     commandlineexecutor.Execute
	}{
		{
			name: "SuccessASCS",
			inst: &sapb.SAPInstance{
				Sapsid:         "DEV",
				InstanceNumber: "00",
				User:           "devadm",
				InstanceId:     "ASCS00",
				SapcontrolPath: "/usr/sap/DEV/ASCS00/exe/sapcontrol",
			},
			instanceName: "ASCS",
			wantPort:     "8100",
			wantType:     sapb.InstanceType_NETWEAVER,
			wantKind:     sapb.InstanceKind_CS,
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
		},
		{
			name: "SuccessD",
			inst: &sapb.SAPInstance{
				Sapsid:         "DEV",
				InstanceNumber: "01",
				User:           "devadm",
				InstanceId:     "D01",
				SapcontrolPath: "/usr/sap/DEV/D01/exe/sapcontrol",
			},
			instanceName: "D",
			wantPort:     "50100",
			wantType:     sapb.InstanceType_NETWEAVER,
			wantKind:     sapb.InstanceKind_APP,
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
		},
		{
			name:         "ERSInstance",
			instanceName: "ERS",
			wantType:     sapb.InstanceType_NETWEAVER,
			wantKind:     sapb.InstanceKind_ERS,
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
		},
		{
			name:         "HDBInstance",
			instanceName: "HDB",
			wantType:     sapb.InstanceType_HANA,
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
		},
		{
			name:         "UnknownInstance",
			instanceName: "XYZ",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
		},
		{
			name:         "WebDispatcherAsSeperateSID",
			instanceName: "W01",
			wantKind:     sapb.InstanceKind_APP,
			wantType:     sapb.InstanceType_NETWEAVER,
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotPort, gotType, gotKind := findPort(context.Background(), test.inst, test.instanceName, test.fakeExec)
			if gotPort != test.wantPort || gotType != test.wantType || gotKind != test.wantKind {
				t.Errorf("findPort() returned unexpected values. got(%v, %v, %v), want(%v, %v, %v)",
					gotPort, gotType, gotKind, test.wantPort, test.wantType, test.wantKind)
			}
		})
	}
}

func TestParseHTTPPort(t *testing.T) {
	tests := []struct {
		name     string
		fakeExec commandlineexecutor.Execute
		want     string
		wantErr  error
	}{
		{
			name: "Success",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `10.10.2022 21:32:40\nParameterValue\nOK\nPROT=HTTP,PORT=8100`,
				}
			},
			want: "8100",
		},
		{
			name: "SMTPPortConfigured",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `13.10.2022 20:08:39\nParameterValue\nOK\n\nPROT=SMTP,PORT=0,TIMEOUT=120,PROCTIMEOUT=120`,
				}
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "HTTPPortNotConfigured",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `13.10.2022 20:13:47\nParameterValue\nFAIL: Invalid parameter`,
				}
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "CommandFailure",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					Error: cmpopts.AnyError,
				}
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "HTTPPort0",
			fakeExec: func(context.Context, commandlineexecutor.Params) commandlineexecutor.Result {
				return commandlineexecutor.Result{
					StdOut: `10.10.2022 21:32:40\nParameterValue\nOK\nPROT=HTTP,PORT=0`,
				}
			},
			wantErr: cmpopts.AnyError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotPort, gotErr := parseHTTPPort(context.Background(), commandlineexecutor.Params{}, test.fakeExec)
			if !cmp.Equal(gotErr, test.wantErr, cmpopts.EquateErrors()) {
				t.Errorf("parseHTTPPort() returned error = %v, want %v", gotErr, test.wantErr)
			}
			if gotPort != test.want {
				t.Errorf("parseHTTPPort() returned port= %v, want %v", gotPort, test.want)
			}
		})
	}
}

func TestReadHANACredentials(t *testing.T) {
	tests := []struct {
		name                   string
		hanaConfig             *cpb.HANAMetricsConfig
		gceService             GCEInterface
		wantUser, wantPassword string
		wantErr                error
	}{
		{
			name: "UserAndPasswordInConfigFile",
			hanaConfig: &cpb.HANAMetricsConfig{
				HanaDbUser:     "hdbadm",
				HanaDbPassword: "Dummy",
			},
			wantUser:     "hdbadm",
			wantPassword: "Dummy",
		},
		{
			name:    "NoUserAndPassword",
			wantErr: cmpopts.AnyError,
		},
		{
			name: "DefaultUser",
			hanaConfig: &cpb.HANAMetricsConfig{
				HanaDbPassword: "Dummy",
			},
			wantUser:     "SYSTEM",
			wantPassword: "Dummy",
		},
		{
			name: "PasswordNotInConfigFile",
			hanaConfig: &cpb.HANAMetricsConfig{
				HanaDbUser: "hdbadm",
			},
			wantErr: cmpopts.AnyError,
		},
		{
			name: "PasswordInSecretManager",
			hanaConfig: &cpb.HANAMetricsConfig{
				HanaDbUser:               "hdbadm",
				HanaDbPasswordSecretName: "TESTSECRET",
			},
			gceService: &fake.TestGCE{
				GetSecretResp: []string{"Dummy"},
				GetSecretErr:  []error{nil},
			},
			wantUser:     "hdbadm",
			wantPassword: "Dummy",
		},
		{
			name: "SecretManagerError",
			hanaConfig: &cpb.HANAMetricsConfig{
				HanaDbUser:               "hdbadm",
				HanaDbPasswordSecretName: "TESTSECRET",
			},
			gceService: &fake.TestGCE{
				GetSecretResp: []string{""},
				GetSecretErr:  []error{errors.New("error")},
			},
			wantErr: cmpopts.AnyError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotUser, gotPassword, gotErr := ReadHANACredentials(context.Background(), "test-project", test.hanaConfig, test.gceService)

			if !cmp.Equal(gotErr, test.wantErr, cmpopts.EquateErrors()) {
				t.Errorf("ReadHANACredentials() returned error = %v, want %v", gotErr, test.wantErr)
			}

			if gotUser != test.wantUser {
				t.Errorf("ReadHANACredentials() returned user = %s, want %s", gotUser, test.wantUser)
			}

			if gotPassword != test.wantPassword {
				t.Errorf("ReadHANACredentials() returned Password = %s, want %s", gotPassword, test.wantPassword)
			}
		})
	}
}

func TestBuildURLAndServiceName(t *testing.T) {
	tests := []struct {
		name            string
		instanceName    string
		httpPort        string
		wantURL         string
		wantServiceName string
		wantErr         error
	}{
		{
			name:            "ABAP",
			instanceName:    "DVEBMGS",
			httpPort:        "1234",
			wantURL:         "http://localhost:1234/sap/public/icman/ping",
			wantServiceName: "SAP-ICM-ABAP",
		},
		{
			name:            "JAVA",
			instanceName:    "J",
			httpPort:        "1234",
			wantURL:         "http://localhost:1234/sap/admin/public/images/sap.png",
			wantServiceName: "SAP-ICM-Java",
		},
		{
			name:            "ASCSInstance",
			instanceName:    "ASCS",
			httpPort:        "1234",
			wantURL:         "http://localhost:1234/msgserver/text/logon",
			wantServiceName: "SAP-CS",
		},
		{
			name:     "EmptyPort",
			httpPort: "",
			wantErr:  cmpopts.AnyError,
		},
		{
			name:         "InvalidInstance",
			instanceName: "HDB",
			wantErr:      cmpopts.AnyError,
		},
		{
			name:     "EmptyInstance",
			httpPort: "1234",
			wantErr:  cmpopts.AnyError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotURL, gotServiceName, gotErr := buildURLAndServiceName(test.instanceName, test.httpPort)

			if !cmp.Equal(gotErr, test.wantErr, cmpopts.EquateErrors()) {
				t.Errorf("buildURLAndServiceName() returned error = %v, want %v.", gotErr, test.wantErr)
			}

			if gotURL != test.wantURL {
				t.Errorf("buildURLAndServiceName() returned URL = %s, want %s.", gotURL, test.wantURL)
			}

			if gotServiceName != test.wantServiceName {
				t.Errorf("buildURLAndServiceName() returned service name = %s, want %s.", gotServiceName, test.wantServiceName)
			}
		})
	}
}