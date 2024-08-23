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

// Package staginghanadiskrestore implements one time execution for HANA Disk based restore workflow.
package staginghanadiskrestore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"flag"
	backoff "github.com/cenkalti/backoff/v4"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"github.com/google/subcommands"
	"github.com/GoogleCloudPlatform/sapagent/internal/hanabackup"
	"github.com/GoogleCloudPlatform/sapagent/internal/instanceinfo"
	"github.com/GoogleCloudPlatform/sapagent/internal/onetime"
	"github.com/GoogleCloudPlatform/sapagent/internal/usagemetrics"
	"github.com/GoogleCloudPlatform/sapagent/internal/utils/instantsnapshotgroup"
	"github.com/GoogleCloudPlatform/sapagent/shared/cloudmonitoring"
	"github.com/GoogleCloudPlatform/sapagent/shared/commandlineexecutor"
	"github.com/GoogleCloudPlatform/sapagent/shared/gce"
	"github.com/GoogleCloudPlatform/sapagent/shared/log"
	"github.com/GoogleCloudPlatform/sapagent/shared/timeseries"

	mrpb "google.golang.org/genproto/googleapis/monitoring/v3"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	cpb "github.com/GoogleCloudPlatform/sapagent/protos/configuration"
	ipb "github.com/GoogleCloudPlatform/sapagent/protos/instanceinfo"
)

type (
	// getDataPaths provides testable replacement for hanabackup.CheckDataDir
	getDataPaths func(context.Context, commandlineexecutor.Execute) (string, string, string, error)

	// getLogPaths provides testable replacement for hanabackup.CheckLogDir
	getLogPaths func(context.Context, commandlineexecutor.Execute) (string, string, string, error)

	// waitForIndexServerToStopWithRetry provides testable replacement for hanabackup.WaitForIndexServerToStopWithRetry
	waitForIndexServerToStopWithRetry func(ctx context.Context, user string, exec commandlineexecutor.Execute) error

	// metricClientCreator provides testable replacement for monitoring.NewMetricClient API.
	metricClientCreator func(context.Context, ...option.ClientOption) (*monitoring.MetricClient, error)

	// gceServiceFunc provides testable replacement for gce.New API.
	gceServiceFunc func(context.Context) (*gce.GCE, error)

	// computeServiceFunc provides testable replacement for compute.Service API
	computeServiceFunc func(context.Context) (*compute.Service, error)

	// gceInterface is the testable equivalent for gce.GCE for secret manager access.
	gceInterface interface {
		GetInstance(project, zone, instance string) (*compute.Instance, error)
		ListZoneOperations(project, zone, filter string, maxResults int64) (*compute.OperationList, error)
		GetDisk(project, zone, name string) (*compute.Disk, error)
		ListDisks(project, zone, filter string) (*compute.DiskList, error)

		DiskAttachedToInstance(projectID, zone, instanceName, diskName string) (string, bool, error)
		AttachDisk(ctx context.Context, diskName string, cp *ipb.CloudProperties, project, dataDiskZone string) error
		DetachDisk(ctx context.Context, cp *ipb.CloudProperties, project, dataDiskZone, dataDiskName, dataDiskDeviceName string) error
		WaitForDiskOpCompletionWithRetry(ctx context.Context, op *compute.Operation, project, dataDiskZone string) error
	}

	// ISGInterface is the testable equivalent for ISGService for ISG operations.
	ISGInterface interface {
		GetResponse(ctx context.Context, method string, baseURL string, data []byte) ([]byte, error)
		CreateISG(ctx context.Context, project, zone string, data []byte) error
		DescribeInstantSnapshots(ctx context.Context, project, zone, isgName string) ([]instantsnapshotgroup.ISItem, error)
		DescribeStandardSnapshots(ctx context.Context, project, zone, isgName string) ([]*compute.Snapshot, error)
		TruncateName(ctx context.Context, src, suffix string) string
		NewService() error
	}
)

const (
	metricPrefix = "workload.googleapis.com/sap/agent/"
)

var (
	workflowStartTime time.Time
)

// Restorer has args for staginghanadiskrestore subcommands
type Restorer struct {
	Project, Sid, HanaSidAdm, DataDiskName, DataDiskDeviceName string
	DataDiskZone, SourceSnapshot, GroupSnapshot, NewDiskType   string
	disks                                                      []*ipb.Disk
	DataDiskVG                                                 string
	gceService                                                 gceInterface
	computeService                                             *compute.Service
	isgService                                                 ISGInterface
	cgName                                                     string
	baseDataPath, baseLogPath                                  string
	logicalDataPath, logicalLogPath                            string
	physicalDataPath, physicalLogPath                          string
	labelsOnDetachedDisk                                       string
	timeSeriesCreator                                          cloudmonitoring.TimeSeriesCreator
	help                                                       bool
	SendToMonitoring                                           bool
	SkipDBSnapshotForChangeDiskType                            bool
	HANAChangeDiskTypeOTEName                                  string
	LogLevel, LogPath                                          string
	ForceStopHANA                                              bool
	isGroupSnapshot                                            bool
	NewdiskName                                                string
	CSEKKeyFile                                                string
	ProvisionedIops, ProvisionedThroughput, DiskSizeGb         int64
	IIOTEParams                                                *onetime.InternallyInvokedOTE
	oteLogger                                                  *onetime.OTELogger
}

// Name implements the subcommand interface for staginghanadiskrestore.
func (*Restorer) Name() string { return "staginghanadiskrestore" }

// Synopsis implements the subcommand interface for staginghanadiskrestore.
func (*Restorer) Synopsis() string {
	return "invoke HANA staginghanadiskrestore using workflow to restore from disk snapshot"
}

// Usage implements the subcommand interface for staginghanadiskrestore.
func (*Restorer) Usage() string {
	return `Usage: staginghanadiskrestore -sid=<HANA-sid> -source-snapshot=<snapshot-name>
  -data-disk-name=<disk-name> -data-disk-zone=<disk-zone> -new-disk-name=<name-less-than-63-chars>
  [-project=<project-name>] [-new-disk-type=<Type of the new disk>] [-force-stop-hana=<true|false>]
  [-hana-sidadm=<hana-sid-user-name>] [-provisioned-iops=<Integer value between 10,000 and 120,000>]
  [-provisioned-throughput=<Integer value between 1 and 7,124>] [-disk-size-gb=<New disk size in GB>]
  [-send-metrics-to-monitoring]=<true|false>
  [csek-key-file]=<path-to-key-file>]
  [-h] [-loglevel=<debug|info|warn|error>] [-log-path=<log-path>]` + "\n"
}

// SetFlags implements the subcommand interface for staginghanadiskrestore.
func (r *Restorer) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&r.Sid, "sid", "", "HANA sid. (required)")
	fs.StringVar(&r.DataDiskName, "data-disk-name", "", "Current disk name. (optional) Default: Disk backing up /hana/data")
	fs.StringVar(&r.DataDiskZone, "data-disk-zone", "", "Current disk zone. (optional) Default: Same zone as current instance")
	fs.StringVar(&r.SourceSnapshot, "source-snapshot", "", "Source disk snapshot to restore from. (optional) either source-snapshot or group-snapshot must be provided")
	fs.StringVar(&r.GroupSnapshot, "group-snapshot", "", "Name of group snapshot to restore from. (optional) either source-snapshot or group-snapshot must be provided")
	fs.StringVar(&r.NewdiskName, "new-disk-name", "", "New disk name. (required) must be less than 63 characters long")
	fs.StringVar(&r.Project, "project", "", "GCP project. (optional) Default: project corresponding to this instance")
	fs.StringVar(&r.NewDiskType, "new-disk-type", "", "Type of the new disk. (optional) Default: same type as disk passed in data-disk-name.")
	fs.StringVar(&r.HanaSidAdm, "hana-sidadm", "", "HANA sidadm username. (optional) Default: <sid>adm")
	fs.StringVar(&r.labelsOnDetachedDisk, "labels-on-detached-disk", "", "Labels to be appended to detached disks. (optional) Default: empty. Accepts comma separated key-value pairs, like \"key1=value1,key2=value2\"")
	fs.BoolVar(&r.ForceStopHANA, "force-stop-hana", false, "Forcefully stop HANA using `HDB kill` before attempting restore.(optional) Default: false.")
	fs.Int64Var(&r.DiskSizeGb, "disk-size-gb", 0, "New disk size in GB, must not be less than the size of the source (optional)")
	fs.Int64Var(&r.ProvisionedIops, "provisioned-iops", 0, "Number of I/O operations per second that the disk can handle. (optional)")
	fs.Int64Var(&r.ProvisionedThroughput, "provisioned-throughput", 0, "Number of throughput mb per second that the disk can handle. (optional)")
	fs.BoolVar(&r.SendToMonitoring, "send-metrics-to-monitoring", true, "Send restore related metrics to cloud monitoring. (optional) Default: true")
	fs.StringVar(&r.CSEKKeyFile, "csek-key-file", "", `Path to a Customer-Supplied Encryption Key (CSEK) key file for the source snapshot. (required if source snapshot is encrypted)`)
	fs.StringVar(&r.LogPath, "log-path", "", "The log path to write the log file (optional), default value is /var/log/google-cloud-sap-agent/staginghanadiskrestore.log")
	fs.BoolVar(&r.help, "h", false, "Displays help")
	fs.StringVar(&r.LogLevel, "loglevel", "info", "Sets the logging level")
}

// Execute implements the subcommand interface for staginghanadiskrestore.
func (r *Restorer) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	// Help will return before the args are parsed.
	_, cp, exitStatus, completed := onetime.Init(ctx, onetime.InitOptions{
		Name:     r.Name(),
		Help:     r.help,
		Fs:       f,
		IIOTE:    r.IIOTEParams,
		LogLevel: r.LogLevel,
		LogPath:  r.LogPath,
	}, args...)
	if !completed {
		return exitStatus
	}

	return r.Run(ctx, onetime.CreateRunOptions(cp, false))
}

// Run performs the functionality specified by the staginghanadiskrestore subcommand.
func (r *Restorer) Run(ctx context.Context, runOpts *onetime.RunOptions) subcommands.ExitStatus {
	r.oteLogger = onetime.CreateOTELogger(runOpts.DaemonMode)
	return r.restoreHandler(ctx, monitoring.NewMetricClient, gce.NewGCEClient, onetime.NewComputeService, runOpts.CloudProperties, hanabackup.CheckDataDir, hanabackup.CheckLogDir)
}

// validateParameters validates the parameters passed to the restore subcommand.
func (r *Restorer) validateParameters(os string, cp *ipb.CloudProperties) error {
	if r.SkipDBSnapshotForChangeDiskType {
		log.Logger.Debug("Skip DB Snapshot for Change Disk Type")
		return nil
	}
	if os == "windows" {
		return fmt.Errorf("disk snapshot restore is only supported on Linux systems")
	}

	// Checking if sufficient arguments are passed for either group snapshot or single snapshot.
	// Only SID is required for restoring from groupSnapshot.
	// DataDiskName and NewdiskNames are fetched and respectively created
	// from individual snapshots mapped to groupSnapshot.
	restoreFromGroupSnapshot := !(r.Sid == "" || r.GroupSnapshot == "")
	restoreFromSingleSnapshot := !(r.Sid == "" || r.SourceSnapshot == "" || r.NewdiskName == "")

	if restoreFromGroupSnapshot == true && restoreFromSingleSnapshot == true {
		return fmt.Errorf("either source-snapshot or group-snapshot must be provided, not both. Usage: %s", r.Usage())
	} else if restoreFromGroupSnapshot == false && restoreFromSingleSnapshot == false {
		return fmt.Errorf("required arguments not passed. Usage: %s", r.Usage())
	}
	if len(r.NewdiskName) > 63 {
		return fmt.Errorf("the new-disk-name is longer than 63 chars which is not supported, please provide a shorter name")
	}

	if r.Project == "" {
		r.Project = cp.GetProjectId()
	}
	if r.HanaSidAdm == "" {
		r.HanaSidAdm = strings.ToLower(r.Sid) + "adm"
	}

	if restoreFromGroupSnapshot {
		r.isgService = &instantsnapshotgroup.ISGService{}
		if err := r.isgService.NewService(); err != nil {
			return fmt.Errorf("failed to create ISG service, err: %w", err)
		}
		r.isGroupSnapshot = true
	}

	log.Logger.Debug("Parameter validation successful.")
	log.Logger.Infof("List of parameters to be used: %+v", r)

	return nil
}

// restoreHandler is the main handler for the restore subcommand.
func (r *Restorer) restoreHandler(ctx context.Context, mcc metricClientCreator, gceServiceCreator gceServiceFunc, computeServiceCreator computeServiceFunc, cp *ipb.CloudProperties, checkDataDir getDataPaths, checkLogDir getLogPaths) subcommands.ExitStatus {
	var err error
	if err = r.validateParameters(runtime.GOOS, cp); err != nil {
		log.Print(err.Error())
		return subcommands.ExitUsageError
	}
	if r.isGroupSnapshot {
		ctx = context.WithValue(ctx, instantsnapshotgroup.EnvKey("env"), "staging")
	}

	r.timeSeriesCreator, err = mcc(ctx)
	if err != nil {
		log.CtxLogger(ctx).Errorw("Failed to create Cloud Monitoring metric client", "error", err)
		return subcommands.ExitFailure
	}

	r.gceService, err = gceServiceCreator(ctx)
	if err != nil {
		r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: Failed to create GCE service", err)
		return subcommands.ExitFailure
	}

	log.CtxLogger(ctx).Infow("Starting HANA disk snapshot restore", "sid", r.Sid)

	if r.computeService, err = computeServiceCreator(ctx); err != nil {
		r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: Failed to create compute service,", err)
		return subcommands.ExitFailure
	}

	if err := r.checkPreConditions(ctx, cp, checkDataDir, checkLogDir); err != nil {
		r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: Pre-restore check failed,", err)
		return subcommands.ExitFailure
	}
	if !r.SkipDBSnapshotForChangeDiskType {
		if err := r.prepare(ctx, cp, hanabackup.WaitForIndexServerToStopWithRetry, commandlineexecutor.ExecuteCommand); err != nil {
			r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: HANA restore prepare failed,", err)
			return subcommands.ExitFailure
		}
	} else {
		if err := r.prepareForHANAChangeDiskType(ctx, cp); err != nil {
			r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: HANA restore prepare failed,", err)
			return subcommands.ExitFailure
		}
	}
	// Rescanning to prevent any volume group naming conflicts
	// with restored disk's volume group.
	hanabackup.RescanVolumeGroups(ctx)

	workflowStartTime = time.Now()
	if !r.isGroupSnapshot {
		if err := r.diskRestore(ctx, commandlineexecutor.ExecuteCommand, cp); err != nil {
			return subcommands.ExitFailure
		}
	} else {
		if err := r.groupRestore(ctx, cp); err != nil {
			return subcommands.ExitFailure
		}
	}
	workflowDur := time.Since(workflowStartTime)
	defer r.sendDurationToCloudMonitoring(ctx, metricPrefix+r.Name()+"/totaltime", workflowDur, cloudmonitoring.NewDefaultBackOffIntervals(), cp)
	r.oteLogger.LogMessageToFileAndConsole(ctx, "SUCCESS: HANA restore from disk snapshot successful. Please refer https://cloud.google.com/solutions/sap/docs/agent-for-sap/latest/disk-snapshot-backup-recovery#recover_to_specific_point-in-time for next steps.")
	if r.labelsOnDetachedDisk != "" {
		if !r.isGroupSnapshot {
			if err := r.appendLabelsToDetachedDisk(ctx, r.DataDiskName); err != nil {
				return subcommands.ExitFailure
			}
		} else {
			for _, d := range r.disks {
				if err := r.appendLabelsToDetachedDisk(ctx, d.DiskName); err != nil {
					return subcommands.ExitFailure
				}
			}
		}
	}
	return subcommands.ExitSuccess
}

func (r *Restorer) fetchVG(ctx context.Context, cp *ipb.CloudProperties, exec commandlineexecutor.Execute, physicalDataPath string) (string, error) {
	result := exec(ctx, commandlineexecutor.Params{
		Executable:  "/sbin/pvs",
		ArgsToSplit: physicalDataPath,
	})
	if result.Error != nil {
		return "", fmt.Errorf("failure fetching VG, stderr: %s, err: %s", result.StdErr, result.Error)
	}

	// A physical volume can only be a part of one volume group at a time.
	// A valid output looks like this:
	// PV         VG    Fmt  Attr PSize   PFree
	// /dev/sdd   my_vg lvm2 a--  500.00g 300.00g
	lines := strings.Split(result.StdOut, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("failure fetching VG, disk does not belong to any vg")
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return "", fmt.Errorf("failure fetching VG, disk does not belong to any vg")
	}
	return fields[1], nil
}

// prepare stops HANA, unmounts data directory and detaches old data disk.
func (r *Restorer) prepare(ctx context.Context, cp *ipb.CloudProperties, waitForIndexServerStop waitForIndexServerToStopWithRetry, exec commandlineexecutor.Execute) error {
	mountPath, err := hanabackup.ReadDataDirMountPath(ctx, r.baseDataPath, exec)
	if err != nil {
		return fmt.Errorf("failed to read data directory mount path: %v", err)
	}
	if err := hanabackup.StopHANA(ctx, r.ForceStopHANA, r.HanaSidAdm, r.Sid, exec); err != nil {
		return fmt.Errorf("failed to stop HANA: %v", err)
	}
	if err := waitForIndexServerStop(ctx, r.HanaSidAdm, exec); err != nil {
		return fmt.Errorf("hdbindexserver process still running after HANA is stopped: %v", err)
	}

	if err := hanabackup.Unmount(ctx, mountPath, exec); err != nil {
		return fmt.Errorf("failed to unmount data directory: %v", err)
	}

	if !r.isGroupSnapshot {
		vg, err := r.fetchVG(ctx, cp, exec, r.physicalDataPath)
		if err != nil {
			return err
		}
		r.DataDiskVG = vg

		log.CtxLogger(ctx).Info("Detaching old data disk", "disk", r.DataDiskName, "physicalDataPath", r.physicalDataPath)
		if err := r.gceService.DetachDisk(ctx, cp, r.Project, r.DataDiskZone, r.DataDiskName, r.DataDiskDeviceName); err != nil {
			// If detach fails, rescan the volume groups to ensure the directories are mounted.
			hanabackup.RescanVolumeGroups(ctx)
			return fmt.Errorf("failed to detach old data disk: %v", err)
		}
	} else {
		if err := r.validateDisksBelongToCG(ctx); err != nil {
			return err
		}

		disksDetached := []*ipb.Disk{}
		for _, d := range r.disks {
			log.CtxLogger(ctx).Info("Detaching old data disk", "disk", d.DiskName, "physicalDataPath", r.physicalDataPath)
			if err := r.detachDisk(ctx, d.DiskName, cp.GetInstanceName()); err != nil {
				log.CtxLogger(ctx).Error("failed to detach old data disk: %v", err)
				// Reattaching detached disks.
				for _, disk := range disksDetached {
					if err := r.attachDisk(ctx, disk.DiskName, cp.GetInstanceName()); err != nil {
						return fmt.Errorf("failed to attach old data disk that was detached earlier: %v", err)
					}
				}

				// If detach fails, rescan the volume groups to ensure the directories are mounted.
				hanabackup.RescanVolumeGroups(ctx)
				return fmt.Errorf("failed to detach old data disk: %v", err)
			}
			if err := r.modifyDiskInCG(ctx, d.DiskName, false); err != nil {
				log.CtxLogger(ctx).Error("failed to remove old disk from CG: %v", err)
				// Reattaching detached disks.
				for _, disk := range disksDetached {
					if err := r.attachDisk(ctx, disk.DiskName, cp.GetInstanceName()); err != nil {
						return fmt.Errorf("failed to attach old data disk that was detached earlier: %v", err)
					}
				}
			}

			disksDetached = append(disksDetached, d)
		}
	}

	log.CtxLogger(ctx).Info("HANA restore prepare succeeded.")
	return nil
}

func (r *Restorer) prepareForHANAChangeDiskType(ctx context.Context, cp *ipb.CloudProperties) error {
	mountPath, err := hanabackup.ReadDataDirMountPath(ctx, r.baseDataPath, commandlineexecutor.ExecuteCommand)
	if err != nil {
		return fmt.Errorf("failed to read data directory mount path: %v", err)
	}
	if err := hanabackup.Unmount(ctx, mountPath, commandlineexecutor.ExecuteCommand); err != nil {
		return fmt.Errorf("failed to unmount data directory: %v", err)
	}
	if err := r.gceService.DetachDisk(ctx, cp, r.Project, r.DataDiskZone, r.DataDiskName, r.DataDiskDeviceName); err != nil {
		// If detach fails, rescan the volume groups to ensure the directories are mounted.
		hanabackup.RescanVolumeGroups(ctx)
		return fmt.Errorf("failed to detach old data disk: %v", err)
	}
	log.CtxLogger(ctx).Info("HANA restore prepareForHANAChangeDiskType succeeded.")
	return nil
}

// diskRestore creates a new data disk restored from a single snapshot and attaches it to the instance.
func (r *Restorer) diskRestore(ctx context.Context, exec commandlineexecutor.Execute, cp *ipb.CloudProperties) error {
	snapShotKey := ""
	if r.CSEKKeyFile != "" {
		r.oteLogger.LogUsageAction(usagemetrics.EncryptedSnapshotRestore)

		snapShotURI := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/snapshots/%s", r.Project, r.DataDiskZone, r.SourceSnapshot)
		key, err := hanabackup.ReadKey(r.CSEKKeyFile, snapShotURI, os.ReadFile)
		if err != nil {
			r.oteLogger.LogUsageError(usagemetrics.EncryptedSnapshotRestoreFailure)
			return err
		}
		snapShotKey = key
	}

	if err := r.restoreFromSnapshot(ctx, exec, cp, snapShotKey, r.NewdiskName, r.SourceSnapshot); err != nil {
		r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: HANA restore from snapshot failed,", err)
		r.gceService.AttachDisk(ctx, r.DataDiskName, cp, r.Project, r.DataDiskZone)
		hanabackup.RescanVolumeGroups(ctx)
		return err
	}

	hanabackup.RescanVolumeGroups(ctx)
	log.CtxLogger(ctx).Info("HANA restore from snapshot succeeded.")
	return nil
}

// groupRestore creates several new HANA data disks from snapshots belonging to given group snapshot and attaches them to the instance.
func (r *Restorer) groupRestore(ctx context.Context, cp *ipb.CloudProperties) error {
	snapShotKey := ""
	if r.CSEKKeyFile != "" {
		r.oteLogger.LogUsageAction(usagemetrics.EncryptedSnapshotRestore)

		snapShotURI := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/snapshots/%s", r.Project, r.DataDiskZone, r.GroupSnapshot)
		key, err := hanabackup.ReadKey(r.CSEKKeyFile, snapShotURI, os.ReadFile)
		if err != nil {
			r.oteLogger.LogUsageError(usagemetrics.EncryptedSnapshotRestoreFailure)
			return err
		}
		snapShotKey = key
	}

	if err := r.restoreFromGroupSnapshot(ctx, commandlineexecutor.ExecuteCommand, cp, snapShotKey); err != nil {
		r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: HANA restore from group snapshot failed,", err)
		for _, d := range r.disks {
			if r.isGroupSnapshot {
				r.attachDisk(ctx, d.DiskName, cp.GetInstanceName())
			} else {
				r.gceService.AttachDisk(ctx, d.DiskName, cp, r.Project, r.DataDiskZone)
			}
		}
		hanabackup.RescanVolumeGroups(ctx)
		return err
	}

	hanabackup.RescanVolumeGroups(ctx)
	log.CtxLogger(ctx).Info("HANA restore from group snapshot succeeded.")
	return nil
}

// restoreFromSnapshot creates a new HANA data disk and attaches it to the instance.
func (r *Restorer) restoreFromSnapshot(ctx context.Context, exec commandlineexecutor.Execute, cp *ipb.CloudProperties, snapshotKey, newDiskName, sourceSnapshot string) error {
	if !r.isGroupSnapshot {
		if r.computeService == nil {
			return fmt.Errorf("compute service is nil")
		}
		snapshot, err := r.computeService.Snapshots.Get(r.Project, r.SourceSnapshot).Do()
		if err != nil {
			return fmt.Errorf("failed to check if source-snapshot=%v is present: %v", r.SourceSnapshot, err)
		}
		if r.DiskSizeGb == 0 {
			r.DiskSizeGb = snapshot.DiskSizeGb
		}
	} else {
		snapshots, err := r.isgService.DescribeStandardSnapshots(ctx, r.Project, r.DataDiskZone, r.GroupSnapshot)
		if err != nil {
			return fmt.Errorf("failed to check if group-snapshot=%v is present: %v", r.GroupSnapshot, err)
		}
		if r.DiskSizeGb == 0 {
			r.DiskSizeGb = snapshots[0].DiskSizeGb
		}
	}
	disk := &compute.Disk{
		Name:                        newDiskName,
		Type:                        r.NewDiskType,
		Zone:                        r.DataDiskZone,
		SourceSnapshot:              fmt.Sprintf("projects/%s/global/snapshots/%s", r.Project, sourceSnapshot),
		SourceSnapshotEncryptionKey: &compute.CustomerEncryptionKey{RsaEncryptedKey: snapshotKey},
	}
	if r.DiskSizeGb > 0 {
		disk.SizeGb = r.DiskSizeGb
	}
	if r.ProvisionedIops > 0 {
		disk.ProvisionedIops = r.ProvisionedIops
	}
	if r.ProvisionedThroughput > 0 {
		disk.ProvisionedThroughput = r.ProvisionedThroughput
	}
	log.Logger.Infow("Inserting new HANA disk from source snapshot", "diskName", newDiskName, "sourceSnapshot", r.SourceSnapshot)

	op, err := r.computeService.Disks.Insert(r.Project, r.DataDiskZone, disk).Do()
	if err != nil {
		r.oteLogger.LogErrorToFileAndConsole(ctx, "ERROR: HANA restore from snapshot failed,", err)
		return fmt.Errorf("failed to insert new data disk: %v", err)
	}
	if err := r.gceService.WaitForDiskOpCompletionWithRetry(ctx, op, r.Project, r.DataDiskZone); err != nil {
		r.oteLogger.LogErrorToFileAndConsole(ctx, "insert data disk failed", err)
		return fmt.Errorf("insert data disk operation failed: %v", err)
	}

	if err := r.gceService.AttachDisk(ctx, newDiskName, cp, r.Project, r.DataDiskZone); err != nil {
		return fmt.Errorf("failed to attach new data disk to instance: %v", err)
	}

	dev, ok, err := r.gceService.DiskAttachedToInstance(r.Project, r.DataDiskZone, cp.GetInstanceName(), newDiskName)
	if err != nil {
		return fmt.Errorf("failed to check if new disk %v is attached to the instance", newDiskName)
	}
	if !ok {
		return fmt.Errorf("newly created disk %v is not attached to the instance", newDiskName)
	}
	// Introducing sleep to let symlinks for the new disk to be created.
	time.Sleep(5 * time.Second)

	if r.DataDiskVG != "" {
		if err := r.renameLVM(ctx, exec, cp, dev, newDiskName); err != nil {
			log.CtxLogger(ctx).Info("Removing newly attached restored disk")
			dev, _, _ := r.gceService.DiskAttachedToInstance(r.Project, r.DataDiskZone, cp.GetInstanceName(), newDiskName)
			if err := r.gceService.DetachDisk(ctx, cp, r.Project, r.DataDiskZone, newDiskName, dev); err != nil {
				log.CtxLogger(ctx).Info("Failed to detach newly attached restored disk: %v", err)
				return err
			}
			return err
		}
	}
	log.Logger.Info("New disk created from snapshot successfully attached to the instance.")
	return nil
}

// renameLVM renames the LVM volume group of the newly restored disk to
// that of the target disk.
func (r *Restorer) renameLVM(ctx context.Context, exec commandlineexecutor.Execute, cp *ipb.CloudProperties, deviceName string, diskName string) error {
	var err error
	var instanceProperties *ipb.InstanceProperties
	instanceInfoReader := instanceinfo.New(&instanceinfo.PhysicalPathReader{OS: runtime.GOOS}, r.gceService)
	if _, instanceProperties, err = instanceInfoReader.ReadDiskMapping(ctx, &cpb.Configuration{CloudProperties: cp}); err != nil {
		return err
	}

	log.CtxLogger(ctx).Infow("Reading disk mapping to fetch physical mapping of newly attached disk", "ip", instanceProperties)
	var restoredDiskPV string
	for _, d := range instanceProperties.GetDisks() {
		if d.GetDeviceName() == deviceName {
			restoredDiskPV = fmt.Sprintf("/dev/%s", d.GetMapping())
		}
	}

	restoredDiskVG, err := r.fetchVG(ctx, cp, exec, restoredDiskPV)
	log.CtxLogger(ctx).Infow("Fetching vg", "restoredDiskVG", restoredDiskVG, "err", err)
	if err != nil {
		return err
	}

	if restoredDiskVG != r.DataDiskVG {
		result := exec(ctx, commandlineexecutor.Params{
			Executable:  "/sbin/vgrename",
			ArgsToSplit: fmt.Sprintf("%s %s", restoredDiskVG, r.DataDiskVG),
		})
		if result.Error != nil {
			log.CtxLogger(ctx).Errorw("Failed to rename volume group of restored disk", "err", result.StdErr)
			return fmt.Errorf("failed to rename volume group of restored disk '%s' from %s to %s: %v", restoredDiskPV, restoredDiskVG, r.DataDiskVG, result.StdErr)
		}
		log.CtxLogger(ctx).Infow("Renaming volume group of restored disk", "Name of TargetDisk VG", r.DataDiskVG, "Name of RestoredDisk VG", restoredDiskVG)
	}

	return nil
}

// restoreFromGroupSnapshot creates several new HANA data disks from snapshots belonging
// to given group snapshot and attaches them to the instance.
func (r *Restorer) restoreFromGroupSnapshot(ctx context.Context, exec commandlineexecutor.Execute, cp *ipb.CloudProperties, snapshotKey string) error {
	snapshots, err := r.isgService.DescribeStandardSnapshots(ctx, r.Project, r.DataDiskZone, r.GroupSnapshot)
	if err != nil {
		return fmt.Errorf("failed to describe ISG: %v", err)
	}
	log.CtxLogger(ctx).Debugw("ISG", "isg", snapshots)

	for _, snapshot := range snapshots {
		timestamp := time.Now().Unix()
		sourceDiskName := r.isgService.TruncateName(ctx, snapshot.Name, fmt.Sprintf("%d", timestamp))

		if err := r.stagingRestoreFromSnapshot(ctx, exec, cp, snapshotKey, sourceDiskName, snapshot.Name); err != nil {
			return err
		}
	}
	return nil
}

// checkPreConditions checks if the HANA data and log disks are on the same physical disk.
// Also verifies that the data disk is attached to the instance.
func (r *Restorer) checkPreConditions(ctx context.Context, cp *ipb.CloudProperties, checkDataDir getDataPaths, checkLogDir getLogPaths) error {
	var err error
	if r.baseDataPath, r.logicalDataPath, r.physicalDataPath, err = checkDataDir(ctx, commandlineexecutor.ExecuteCommand); err != nil {
		return err
	}
	if r.baseLogPath, r.logicalLogPath, r.physicalLogPath, err = checkLogDir(ctx, commandlineexecutor.ExecuteCommand); err != nil {
		return err
	}
	log.CtxLogger(ctx).Infow("Checking preconditions", "Data directory", r.baseDataPath, "Data file system",
		r.logicalDataPath, "Data physical volume", r.physicalDataPath, "Log directory", r.baseLogPath,
		"Log file system", r.logicalLogPath, "Log physical volume", r.physicalLogPath)

	if r.physicalDataPath == r.physicalLogPath {
		return fmt.Errorf("unsupported: HANA data and HANA log are on the same physical disk - %s", r.physicalDataPath)
	}

	if r.DataDiskName == "" || r.DataDiskZone == "" || r.isGroupSnapshot {
		if err := r.readDiskMapping(ctx, cp, &instanceinfo.PhysicalPathReader{OS: runtime.GOOS}); err != nil {
			return fmt.Errorf("failed to read disks backing /hana/data: %v", err)
		}
	}

	// Verify the disk is attached to the instance.
	if !r.isGroupSnapshot {
		dev, ok, err := r.gceService.DiskAttachedToInstance(r.Project, r.DataDiskZone, cp.GetInstanceName(), r.DataDiskName)
		if err != nil {
			return fmt.Errorf("failed to verify if disk %v is attached to the instance", r.DataDiskName)
		}
		if !ok {
			return fmt.Errorf("the disk data-disk-name=%v is not attached to the instance, please pass the current data disk name", r.DataDiskName)
		}
		r.DataDiskDeviceName = dev
	} else {
		// TODO: Update this when ISG APIs are in prod.
		// Commenting this out for now as it is fails due to staging resources being inaccessible.
		// for _, d := range r.disks {
		// 	_, ok, err := r.gceService.DiskAttachedToInstance(r.Project, r.DataDiskZone, cp.GetInstanceName(), d.GetDiskName())
		// 	if err != nil {
		// 		return fmt.Errorf("failed to verify if disk %v is attached to the instance", d.GetDiskName())
		// 	}
		// 	if !ok {
		// 		return fmt.Errorf("the disk data-disk-name=%v is not attached to the instance", d.GetDiskName())
		// 	}
		// }
	}

	// Verify the snapshot is present.
	if !r.isGroupSnapshot {
		snapshot, err := r.computeService.Snapshots.Get(r.Project, r.SourceSnapshot).Do()
		if err != nil {
			return fmt.Errorf("failed to check if source-snapshot=%v is present: %v", r.SourceSnapshot, err)
		}
		r.extractLabels(ctx, snapshot)
	} else {
		// TODO: Update this when ISG APIs are in prod.
		// if _, err = r.isgService.DescribeISG(ctx, r.Project, r.DataDiskZone, r.GroupSnapshot); err != nil {
		// 	return fmt.Errorf("failed to check if group-snapshot=%v is present: %v", r.GroupSnapshot, err)
		// }
	}

	if r.NewDiskType == "" {
		if !r.isGroupSnapshot {
			d, err := r.computeService.Disks.Get(r.Project, r.DataDiskZone, r.DataDiskName).Do()
			if err != nil {
				return fmt.Errorf("failed to read data disk type: %v", err)
			}
			r.NewDiskType = d.Type
			log.CtxLogger(ctx).Infow("New disk type will be same as the data-disk-name", "diskType", r.NewDiskType)
		} else {
			baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%v/zones/%v/disks/%v", r.Project, r.DataDiskZone, r.disks[0].GetDiskName())
			bodyBytes, err := r.isgService.GetResponse(ctx, "GET", baseURL, nil)
			if err != nil {
				log.CtxLogger(ctx).Errorw("Error", "err getting disk", err)
				return fmt.Errorf("failed to get disk, err: %w", err)
			}

			var disk *compute.Disk
			if err := json.Unmarshal([]byte(bodyBytes), &disk); err != nil {
				return fmt.Errorf("failed to unmarshal response body, err: %w", err)
			}
			r.NewDiskType = disk.Type
			log.CtxLogger(ctx).Infow("New disk type will be same as the disk type previously backing up /hana/data", "diskType", r.NewDiskType)
		}
	} else {
		r.NewDiskType = fmt.Sprintf("https://www.googleapis.com/compute/staging_alpha/projects/%v/zones/%v/diskTypes/%v", r.Project, r.DataDiskZone, r.NewDiskType)
	}
	return nil
}

func (r *Restorer) extractLabels(ctx context.Context, snapshot *compute.Snapshot) {
	for key, value := range snapshot.Labels {
		switch key {
		case "goog-sapagent-provisioned-iops":
			iops, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.CtxLogger(ctx).Errorw("failed to parse provisioned-iops=%v: %v from snapshot label", value, err)
			}
			if r.ProvisionedIops == 0 {
				r.ProvisionedIops = iops
			}
		case "goog-sapagent-provisioned-throughput":
			tpt, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.CtxLogger(ctx).Errorw("failed to parse provisioned-throughput from snapshot label=%v: %v", value, err)
			}
			if r.ProvisionedThroughput == 0 {
				r.ProvisionedThroughput = tpt
			}
		}
	}
}

func (r *Restorer) sendDurationToCloudMonitoring(ctx context.Context, mtype string, dur time.Duration, bo *cloudmonitoring.BackOffIntervals, cp *ipb.CloudProperties) bool {
	if !r.SendToMonitoring {
		return false
	}
	log.CtxLogger(ctx).Infow("Sending HANA disk snapshot duration to cloud monitoring", "duration", dur)
	ts := []*mrpb.TimeSeries{
		timeseries.BuildFloat64(timeseries.Params{
			CloudProp:    timeseries.ConvertCloudProperties(cp),
			MetricType:   mtype,
			Timestamp:    tspb.Now(),
			Float64Value: dur.Seconds(),
			MetricLabels: map[string]string{
				"sid":           r.Sid,
				"snapshot_name": r.SourceSnapshot,
			},
		}),
	}
	if _, _, err := cloudmonitoring.SendTimeSeries(ctx, ts, r.timeSeriesCreator, bo, r.Project); err != nil {
		log.CtxLogger(ctx).Debugw("Error sending duration metric to cloud monitoring", "error", err.Error())
		return false
	}
	return true
}

func (r *Restorer) readDiskMapping(ctx context.Context, cp *ipb.CloudProperties, diskMapper instanceinfo.DiskMapper) error {
	var instanceProperties *ipb.InstanceProperties
	var err error

	instanceInfoReader := instanceinfo.New(diskMapper, r.gceService)
	if _, instanceProperties, err = instanceInfoReader.ReadDiskMapping(ctx, &cpb.Configuration{CloudProperties: cp}); err != nil {
		return err
	}

	log.CtxLogger(ctx).Debugw("Reading disk mapping", "ip", instanceProperties)
	for _, d := range instanceProperties.GetDisks() {
		if strings.Contains(r.physicalDataPath, d.GetMapping()) {
			log.CtxLogger(ctx).Debugw("Found disk mapping", "physicalPath", r.physicalDataPath, "diskName", d.GetDiskName())
			if r.isGroupSnapshot {
				r.disks = append(r.disks, d)
				r.DataDiskZone = cp.GetZone()
			} else {
				if r.DataDiskName != "" && r.DataDiskName != d.GetDiskName() {
					log.CtxLogger(ctx).Debugw("Disk name does not match provided disk's name, skipping", "DataDiskName", r.DataDiskName, "disk", d.GetDiskName())
					continue
				}
				r.DataDiskName = d.GetDiskName()
				r.DataDiskZone = cp.GetZone()
			}
		}
	}
	return nil
}

// appendLabelsToDetachedDisk appends and sets labels to the detached disk.
func (r *Restorer) appendLabelsToDetachedDisk(ctx context.Context, diskName string) (err error) {
	var disk *compute.Disk
	if disk, err = r.computeService.Disks.Get(r.Project, r.DataDiskZone, diskName).Do(); err != nil {
		return fmt.Errorf("failed to get disk: %v", err)
	}
	labelFingerprint := disk.LabelFingerprint
	labels, err := r.appendLabels(disk.Labels)
	if err != nil {
		return err
	}

	setLabelRequest := &compute.ZoneSetLabelsRequest{
		LabelFingerprint: labelFingerprint,
		Labels:           labels,
	}

	op, err := r.computeService.Disks.SetLabels(r.Project, r.DataDiskZone, diskName, setLabelRequest).Do()
	if err != nil {
		return fmt.Errorf("failed to append labels on detached disk: %v", err)
	}
	if err = r.gceService.WaitForDiskOpCompletionWithRetry(ctx, op, r.Project, r.DataDiskZone); err != nil {
		return fmt.Errorf("failed to append labels on detached disk: %v", err)
	}
	return nil
}

func (r *Restorer) appendLabels(labels map[string]string) (map[string]string, error) {
	pairs := strings.Split(strings.ReplaceAll(r.labelsOnDetachedDisk, " ", ""), ",")

	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("failed to parse labels on detached disk: %v", pair)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		labels[key] = value
	}

	return labels, nil
}

func (r *Restorer) stagingRestoreFromSnapshot(ctx context.Context, exec commandlineexecutor.Execute, cp *ipb.CloudProperties, snapshotKey, newDiskName, sourceSnapshot string) error {
	snapshot, err := r.GetStandardSnapshot(ctx, sourceSnapshot)
	if err != nil {
		return fmt.Errorf("failed to get standard snapshot: %v", err)
	}
	if r.DiskSizeGb == 0 {
		r.DiskSizeGb = snapshot.DiskSizeGb
	}

	_, err = r.CreateDisk(ctx, snapshotKey, newDiskName, sourceSnapshot)
	if err != nil {
		return fmt.Errorf("failed to create disk: %v", err)
	}
	log.Logger.Infow("Inserting new HANA disk from source snapshot", "diskName", newDiskName, "sourceSnapshot", sourceSnapshot)

	if err := r.attachDisk(ctx, newDiskName, cp.GetInstanceName()); err != nil {
		return fmt.Errorf("failed to attach new data disk to instance: %v", err)
	}

	dev, ok, err := r.diskAttachedToInstance(ctx, cp.GetInstanceName(), newDiskName)
	if err != nil {
		return fmt.Errorf("failed to check if new disk %v is attached to the instance", newDiskName)
	}
	if !ok {
		return fmt.Errorf("newly created disk %v is not attached to the instance", newDiskName)
	}
	// Introducing sleep to let symlinks for the new disk to be created.
	time.Sleep(5 * time.Second)

	if r.DataDiskVG != "" {
		if err := r.renameLVM(ctx, exec, cp, dev, newDiskName); err != nil {
			log.CtxLogger(ctx).Info("Removing newly attached restored disk")
			if err := r.detachDisk(ctx, newDiskName, cp.GetInstanceName()); err != nil {
				log.CtxLogger(ctx).Info("Failed to detach newly attached restored disk: %v", err)
				return err
			}
			return err
		}
	}

	log.CtxLogger(ctx).Info("Adding newly attached disk to consistency group ", r.cgName)
	if err := r.modifyDiskInCG(ctx, newDiskName, true); err != nil {
		log.CtxLogger(ctx).Warnw("Failed to add newly attached disk to consistency group, please retry after restore succeeds", "err", err)
	}
	log.Logger.Info("New disk created from snapshot successfully attached to the instance.")
	return nil
}

func (r *Restorer) GetStandardSnapshot(ctx context.Context, sourceSnapshot string) (*compute.Snapshot, error) {
	var snapshot *compute.Snapshot
	baseURL := fmt.Sprintf("https://www.googleapis.com/compute/staging_alpha/projects/%s/global/snapshots/%s", r.Project, sourceSnapshot)
	bodyBytes, err := r.isgService.GetResponse(ctx, "GET", baseURL, nil)
	if err != nil {
		log.CtxLogger(ctx).Errorw("Error", "err getting snapshot", err)
		return nil, fmt.Errorf("failed to get snapshot, err: %w", err)
	}
	if err := json.Unmarshal([]byte(bodyBytes), &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body, err: %w", err)
	}
	return snapshot, nil
}

func (r *Restorer) CreateDisk(ctx context.Context, snapshotKey, diskName, sourceSnapshot string) (*compute.Disk, error) {
	disk := map[string]any{
		"name":                        diskName,
		"type":                        r.NewDiskType,
		"zone":                        r.DataDiskZone,
		"sourceSnapshot":              fmt.Sprintf("https://www.googleapis.com/compute/staging_alpha/projects/%s/global/snapshots/%s", r.Project, sourceSnapshot),
		"sourceSnapshotEncryptionKey": &compute.CustomerEncryptionKey{RsaEncryptedKey: snapshotKey},
	}
	if r.DiskSizeGb > 0 {
		disk["sizeGb"] = r.DiskSizeGb
	}
	if r.ProvisionedIops > 0 {
		disk["provisionedIops"] = r.ProvisionedIops
	}
	if r.ProvisionedThroughput > 0 {
		disk["provisionedThroughput"] = r.ProvisionedThroughput
	}

	baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/disks", r.Project, r.DataDiskZone)
	data, err := json.Marshal(disk)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json, err: %w", err)
	}
	bodyBytes, err := r.isgService.GetResponse(ctx, "POST", baseURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create disk, err: %w", err)
	}
	log.CtxLogger(ctx).Infow("Disk creating", "diskName", diskName, "bodyBytes", string(bodyBytes))
	if err := r.waitForDiskCreateCompletionWithRetry(ctx, diskName); err != nil {
		return nil, fmt.Errorf("failed to create disk, err: %w", err)
	}

	return nil, nil
}

func (r *Restorer) detachDisk(ctx context.Context, diskName, instanceName string) error {
	baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/instances/%s/detachDisk", r.Project, r.DataDiskZone, instanceName)
	reqBody := map[string]any{
		"deviceName": diskName,
	}
	log.CtxLogger(ctx).Debugw("DetachDisk", "baseURL", baseURL, "reqBody", reqBody)
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal json, err: %w", err)
	}

	bodyBytes, err := r.isgService.GetResponse(ctx, "POST", baseURL, data)
	log.CtxLogger(ctx).Debugw("DetachDisk", "bodyBytes", string(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to detach disk, err: %w", err)
	}
	time.Sleep(20 * time.Second)
	return nil
}

func (r *Restorer) attachDisk(ctx context.Context, diskName, instanceName string) error {
	baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/instances/%s/attachDisk", r.Project, r.DataDiskZone, instanceName)
	reqBody := map[string]any{
		"deviceName": diskName,
		"source":     fmt.Sprintf("https://www.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/disks/%s", r.Project, r.DataDiskZone, diskName),
	}
	log.CtxLogger(ctx).Debugw("AttachDisk", "baseURL", baseURL, "reqBody", reqBody)
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal json, err: %w", err)
	}

	bodyBytes, err := r.isgService.GetResponse(ctx, "POST", baseURL, data)
	log.CtxLogger(ctx).Debugw("AttachDisk", "bodyBytes", string(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to attach disk, err: %w", err)
	}
	time.Sleep(20 * time.Second)
	return nil
}

func (r *Restorer) diskAttachedToInstance(ctx context.Context, instanceName, diskName string) (string, bool, error) {
	baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/instances/%s", r.Project, r.DataDiskZone, instanceName)
	bodyBytes, err := r.isgService.GetResponse(ctx, "GET", baseURL, nil)
	if err != nil {
		return "", false, fmt.Errorf("failed to get instance, err: %w", err)
	}

	var instance *compute.Instance
	if err := json.Unmarshal(bodyBytes, &instance); err != nil {
		return "", false, fmt.Errorf("failed to unmarshal json, err: %w", err)
	}
	log.CtxLogger(ctx).Debugw("DiskAttachedToInstance", "instance", instance)
	for _, disk := range instance.Disks {
		log.CtxLogger(ctx).Debugw("Getting disk source", "diskSource", disk.Source)
		if strings.Contains(disk.Source, diskName) {
			return disk.DeviceName, true, nil
		}
	}
	return "", false, nil
}

func (r *Restorer) diskExists(ctx context.Context, diskName string) error {
	baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/disks/%s", r.Project, r.DataDiskZone, diskName)
	bodyBytes, err := r.isgService.GetResponse(ctx, "GET", baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to get disk, err: %w", err)
	}
	var response map[string]any
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("failed to unmarshal json, err: %w", err)
	}
	if err := response["error"]; err != nil {
		return fmt.Errorf("could not unmarshal, disk not found, err: %v", err)
	}

	if response["status"] != "READY" {
		return fmt.Errorf("disk is not ready, err: %w", err)
	}
	return nil
}

func (r *Restorer) waitForDiskCreateCompletionWithRetry(ctx context.Context, diskName string) error {
	constantBackoff := backoff.NewConstantBackOff(1 * time.Second)
	bo := backoff.WithContext(backoff.WithMaxRetries(constantBackoff, 300), ctx)
	return backoff.Retry(func() error {
		return r.diskExists(ctx, diskName)
	}, bo)
}

func (r *Restorer) validateDisksBelongToCG(ctx context.Context) error {
	disksTraversed := []string{}
	for _, d := range r.disks {
		var cg string
		var err error

		cg, err = r.readConsistencyGroup(ctx, d.DiskName)
		if err != nil {
			return err
		}

		if r.cgName != "" && cg != r.cgName {
			return fmt.Errorf("all disks should belong to the same consistency group, however disk %s belongs to %s, while other disks %s belong to %s", d, cg, disksTraversed, r.cgName)
		}
		disksTraversed = append(disksTraversed, d.DiskName)
		r.cgName = cg
	}

	return nil
}

// cgPath returns the name of the consistency group (CG) from the resource policies.
func cgPath(policies []string) string {
	// Example policy: https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/resourcePolicies/my-cg
	for _, policyLink := range policies {
		parts := strings.Split(policyLink, "/")
		if len(parts) >= 10 && parts[9] == "resourcePolicies" {
			return parts[10]
		}
	}
	return ""
}

// TODO: Update this when ISG APIs are in prod.
// readConsistencyGroup reads the consistency group (CG) from the resource policies of the disk.
func (r *Restorer) readConsistencyGroup(ctx context.Context, diskName string) (string, error) {
	baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/disks/%s", r.Project, r.DataDiskZone, diskName)
	bodyBytes, err := r.isgService.GetResponse(ctx, "GET", baseURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to read consistency group of disk, err: %w", err)
	}

	var disk compute.Disk
	if err := json.Unmarshal([]byte(bodyBytes), &disk); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body, err: %w", err)
	}

	if cgPath := cgPath(disk.ResourcePolicies); cgPath != "" {
		log.CtxLogger(ctx).Infow("Found disk to conistency group mapping", "disk", disk, "cg", cgPath)
		return cgPath, nil
	}
	return "", fmt.Errorf("failed to find consistency group for disk %v", disk)
}

func (r *Restorer) modifyDiskInCG(ctx context.Context, diskName string, add bool) error {
	action := "addResourcePolicies"
	if !add {
		action = "removeResourcePolicies"
	}
	baseURL := fmt.Sprintf("https://compute.googleapis.com/compute/staging_alpha/projects/%s/zones/%s/disks/%s/%s", r.Project, r.DataDiskZone, diskName, action)

	parts := strings.Split(r.DataDiskZone, "-")
	if len(parts) < 3 {
		return fmt.Errorf("invalid zone, cannot fetch region from it: %s", r.DataDiskZone)
	}
	region := strings.Join(parts[:len(parts)-1], "-")
	reqBody := map[string]any{
		"resourcePolicies": fmt.Sprintf("projects/%s/regions/%s/resourcePolicies/%s", r.Project, region, r.cgName),
	}

	log.CtxLogger(ctx).Debugw(action, "baseURL", baseURL, "reqBody", reqBody)
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal json, err: %w", err)
	}

	bodyBytes, err := r.isgService.GetResponse(ctx, "POST", baseURL, data)
	if err != nil {
		return fmt.Errorf("failed to modify consistency group, err: %w", err)
	}
	log.CtxLogger(ctx).Debugw(action, "bodyBytes", string(bodyBytes))
	return nil
}
