//
//Copyright 2023 Google LLC
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//https://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.6.1
// source: backint/backint.proto

package backint

import (
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LogLevel int32

const (
	LogLevel_LOG_LEVEL_UNSPECIFIED LogLevel = 0
	LogLevel_DEBUG                 LogLevel = 1
	LogLevel_INFO                  LogLevel = 2
	LogLevel_WARNING               LogLevel = 3
	LogLevel_ERROR                 LogLevel = 4
)

// Enum value maps for LogLevel.
var (
	LogLevel_name = map[int32]string{
		0: "LOG_LEVEL_UNSPECIFIED",
		1: "DEBUG",
		2: "INFO",
		3: "WARNING",
		4: "ERROR",
	}
	LogLevel_value = map[string]int32{
		"LOG_LEVEL_UNSPECIFIED": 0,
		"DEBUG":                 1,
		"INFO":                  2,
		"WARNING":               3,
		"ERROR":                 4,
	}
)

func (x LogLevel) Enum() *LogLevel {
	p := new(LogLevel)
	*p = x
	return p
}

func (x LogLevel) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LogLevel) Descriptor() protoreflect.EnumDescriptor {
	return file_backint_backint_proto_enumTypes[0].Descriptor()
}

func (LogLevel) Type() protoreflect.EnumType {
	return &file_backint_backint_proto_enumTypes[0]
}

func (x LogLevel) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LogLevel.Descriptor instead.
func (LogLevel) EnumDescriptor() ([]byte, []int) {
	return file_backint_backint_proto_rawDescGZIP(), []int{0}
}

type Function int32

const (
	Function_FUNCTION_UNSPECIFIED Function = 0
	Function_BACKUP               Function = 1
	Function_RESTORE              Function = 2
	Function_INQUIRE              Function = 3
	Function_DELETE               Function = 4
	Function_DIAGNOSE             Function = 5
)

// Enum value maps for Function.
var (
	Function_name = map[int32]string{
		0: "FUNCTION_UNSPECIFIED",
		1: "BACKUP",
		2: "RESTORE",
		3: "INQUIRE",
		4: "DELETE",
		5: "DIAGNOSE",
	}
	Function_value = map[string]int32{
		"FUNCTION_UNSPECIFIED": 0,
		"BACKUP":               1,
		"RESTORE":              2,
		"INQUIRE":              3,
		"DELETE":               4,
		"DIAGNOSE":             5,
	}
)

func (x Function) Enum() *Function {
	p := new(Function)
	*p = x
	return p
}

func (x Function) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Function) Descriptor() protoreflect.EnumDescriptor {
	return file_backint_backint_proto_enumTypes[1].Descriptor()
}

func (Function) Type() protoreflect.EnumType {
	return &file_backint_backint_proto_enumTypes[1]
}

func (x Function) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Function.Descriptor instead.
func (Function) EnumDescriptor() ([]byte, []int) {
	return file_backint_backint_proto_rawDescGZIP(), []int{1}
}

type StorageClass int32

const (
	StorageClass_STORAGE_CLASS_UNSPECIFIED StorageClass = 0
	StorageClass_STANDARD                  StorageClass = 1
	StorageClass_NEARLINE                  StorageClass = 2
	StorageClass_COLDLINE                  StorageClass = 3
	StorageClass_ARCHIVE                   StorageClass = 4
)

// Enum value maps for StorageClass.
var (
	StorageClass_name = map[int32]string{
		0: "STORAGE_CLASS_UNSPECIFIED",
		1: "STANDARD",
		2: "NEARLINE",
		3: "COLDLINE",
		4: "ARCHIVE",
	}
	StorageClass_value = map[string]int32{
		"STORAGE_CLASS_UNSPECIFIED": 0,
		"STANDARD":                  1,
		"NEARLINE":                  2,
		"COLDLINE":                  3,
		"ARCHIVE":                   4,
	}
)

func (x StorageClass) Enum() *StorageClass {
	p := new(StorageClass)
	*p = x
	return p
}

func (x StorageClass) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (StorageClass) Descriptor() protoreflect.EnumDescriptor {
	return file_backint_backint_proto_enumTypes[2].Descriptor()
}

func (StorageClass) Type() protoreflect.EnumType {
	return &file_backint_backint_proto_enumTypes[2]
}

func (x StorageClass) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use StorageClass.Descriptor instead.
func (StorageClass) EnumDescriptor() ([]byte, []int) {
	return file_backint_backint_proto_rawDescGZIP(), []int{2}
}

type BackintConfiguration struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bucket                  string              `protobuf:"bytes,1,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Retries                 int64               `protobuf:"varint,2,opt,name=retries,proto3" json:"retries,omitempty"`
	ParallelStreams         int64               `protobuf:"varint,3,opt,name=parallel_streams,json=parallelStreams,proto3" json:"parallel_streams,omitempty"`
	Threads                 int64               `protobuf:"varint,4,opt,name=threads,proto3" json:"threads,omitempty"`
	BufferSizeMb            int64               `protobuf:"varint,5,opt,name=buffer_size_mb,json=bufferSizeMb,proto3" json:"buffer_size_mb,omitempty"`
	EncryptionKey           string              `protobuf:"bytes,6,opt,name=encryption_key,json=encryptionKey,proto3" json:"encryption_key,omitempty"`
	Compress                bool                `protobuf:"varint,7,opt,name=compress,proto3" json:"compress,omitempty"`
	KmsKey                  string              `protobuf:"bytes,8,opt,name=kms_key,json=kmsKey,proto3" json:"kms_key,omitempty"`
	ServiceAccountKey       string              `protobuf:"bytes,9,opt,name=service_account_key,json=serviceAccountKey,proto3" json:"service_account_key,omitempty"`
	RateLimitMb             int64               `protobuf:"varint,10,opt,name=rate_limit_mb,json=rateLimitMb,proto3" json:"rate_limit_mb,omitempty"`
	FileReadTimeoutMs       int64               `protobuf:"varint,11,opt,name=file_read_timeout_ms,json=fileReadTimeoutMs,proto3" json:"file_read_timeout_ms,omitempty"`
	DumpData                bool                `protobuf:"varint,12,opt,name=dump_data,json=dumpData,proto3" json:"dump_data,omitempty"`
	LogLevel                LogLevel            `protobuf:"varint,13,opt,name=log_level,json=logLevel,proto3,enum=sapagent.protos.backint.LogLevel" json:"log_level,omitempty"`
	UserId                  string              `protobuf:"bytes,14,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Function                Function            `protobuf:"varint,15,opt,name=function,proto3,enum=sapagent.protos.backint.Function" json:"function,omitempty"`
	InputFile               string              `protobuf:"bytes,16,opt,name=input_file,json=inputFile,proto3" json:"input_file,omitempty"`
	OutputFile              string              `protobuf:"bytes,17,opt,name=output_file,json=outputFile,proto3" json:"output_file,omitempty"`
	ParamFile               string              `protobuf:"bytes,18,opt,name=param_file,json=paramFile,proto3" json:"param_file,omitempty"`
	BackupId                string              `protobuf:"bytes,19,opt,name=backup_id,json=backupId,proto3" json:"backup_id,omitempty"`
	DatabaseObjectCount     int64               `protobuf:"varint,20,opt,name=database_object_count,json=databaseObjectCount,proto3" json:"database_object_count,omitempty"`
	BackupLevel             string              `protobuf:"bytes,21,opt,name=backup_level,json=backupLevel,proto3" json:"backup_level,omitempty"`
	LogDelaySec             int64               `protobuf:"varint,22,opt,name=log_delay_sec,json=logDelaySec,proto3" json:"log_delay_sec,omitempty"`
	LogToCloud              *wrappers.BoolValue `protobuf:"bytes,23,opt,name=log_to_cloud,json=logToCloud,proto3" json:"log_to_cloud,omitempty"`
	RecoveryBucket          string              `protobuf:"bytes,24,opt,name=recovery_bucket,json=recoveryBucket,proto3" json:"recovery_bucket,omitempty"`
	RetryBackoffInitial     int64               `protobuf:"varint,25,opt,name=retry_backoff_initial,json=retryBackoffInitial,proto3" json:"retry_backoff_initial,omitempty"`
	RetryBackoffMax         int64               `protobuf:"varint,26,opt,name=retry_backoff_max,json=retryBackoffMax,proto3" json:"retry_backoff_max,omitempty"`
	RetryBackoffMultiplier  float32             `protobuf:"fixed32,27,opt,name=retry_backoff_multiplier,json=retryBackoffMultiplier,proto3" json:"retry_backoff_multiplier,omitempty"`
	ClientEndpoint          string              `protobuf:"bytes,28,opt,name=client_endpoint,json=clientEndpoint,proto3" json:"client_endpoint,omitempty"`
	FolderPrefix            string              `protobuf:"bytes,29,opt,name=folder_prefix,json=folderPrefix,proto3" json:"folder_prefix,omitempty"`
	RecoveryFolderPrefix    string              `protobuf:"bytes,30,opt,name=recovery_folder_prefix,json=recoveryFolderPrefix,proto3" json:"recovery_folder_prefix,omitempty"`
	XmlMultipartUpload      bool                `protobuf:"varint,31,opt,name=xml_multipart_upload,json=xmlMultipartUpload,proto3" json:"xml_multipart_upload,omitempty"`
	StorageClass            StorageClass        `protobuf:"varint,32,opt,name=storage_class,json=storageClass,proto3,enum=sapagent.protos.backint.StorageClass" json:"storage_class,omitempty"`
	Metadata                map[string]string   `protobuf:"bytes,33,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	DiagnoseFileMaxSizeGb   int64               `protobuf:"varint,34,opt,name=diagnose_file_max_size_gb,json=diagnoseFileMaxSizeGb,proto3" json:"diagnose_file_max_size_gb,omitempty"`
	SendMetricsToMonitoring *wrappers.BoolValue `protobuf:"bytes,35,opt,name=send_metrics_to_monitoring,json=sendMetricsToMonitoring,proto3" json:"send_metrics_to_monitoring,omitempty"`
	ShortenFolderPath       bool                `protobuf:"varint,36,opt,name=shorten_folder_path,json=shortenFolderPath,proto3" json:"shorten_folder_path,omitempty"`
	DiagnoseTmpDirectory    string              `protobuf:"bytes,37,opt,name=diagnose_tmp_directory,json=diagnoseTmpDirectory,proto3" json:"diagnose_tmp_directory,omitempty"`
	CustomTime              string              `protobuf:"bytes,38,opt,name=custom_time,json=customTime,proto3" json:"custom_time,omitempty"` // This updates the customTime metadata entry:
	//	Format: RFC 3339 format - "YYYY-MM-DD'T'HH:MM:SS.SS'Z'" or
	//	"YYYY-MM-DD'T'HH:MM:SS'Z'".
	//	Example: "2024-06-25T13:25:00Z"
	//
	// Reference:
	// https://cloud.google.com/storage/docs/metadata#custom-time.
	// A value of "UTCNow" will set the customTime to the current time in
	// UTC.
	ParallelRecoveryStreams int64 `protobuf:"varint,39,opt,name=parallel_recovery_streams,json=parallelRecoveryStreams,proto3" json:"parallel_recovery_streams,omitempty"`
}

func (x *BackintConfiguration) Reset() {
	*x = BackintConfiguration{}
	if protoimpl.UnsafeEnabled {
		mi := &file_backint_backint_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackintConfiguration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackintConfiguration) ProtoMessage() {}

func (x *BackintConfiguration) ProtoReflect() protoreflect.Message {
	mi := &file_backint_backint_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackintConfiguration.ProtoReflect.Descriptor instead.
func (*BackintConfiguration) Descriptor() ([]byte, []int) {
	return file_backint_backint_proto_rawDescGZIP(), []int{0}
}

func (x *BackintConfiguration) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

func (x *BackintConfiguration) GetRetries() int64 {
	if x != nil {
		return x.Retries
	}
	return 0
}

func (x *BackintConfiguration) GetParallelStreams() int64 {
	if x != nil {
		return x.ParallelStreams
	}
	return 0
}

func (x *BackintConfiguration) GetThreads() int64 {
	if x != nil {
		return x.Threads
	}
	return 0
}

func (x *BackintConfiguration) GetBufferSizeMb() int64 {
	if x != nil {
		return x.BufferSizeMb
	}
	return 0
}

func (x *BackintConfiguration) GetEncryptionKey() string {
	if x != nil {
		return x.EncryptionKey
	}
	return ""
}

func (x *BackintConfiguration) GetCompress() bool {
	if x != nil {
		return x.Compress
	}
	return false
}

func (x *BackintConfiguration) GetKmsKey() string {
	if x != nil {
		return x.KmsKey
	}
	return ""
}

func (x *BackintConfiguration) GetServiceAccountKey() string {
	if x != nil {
		return x.ServiceAccountKey
	}
	return ""
}

func (x *BackintConfiguration) GetRateLimitMb() int64 {
	if x != nil {
		return x.RateLimitMb
	}
	return 0
}

func (x *BackintConfiguration) GetFileReadTimeoutMs() int64 {
	if x != nil {
		return x.FileReadTimeoutMs
	}
	return 0
}

func (x *BackintConfiguration) GetDumpData() bool {
	if x != nil {
		return x.DumpData
	}
	return false
}

func (x *BackintConfiguration) GetLogLevel() LogLevel {
	if x != nil {
		return x.LogLevel
	}
	return LogLevel_LOG_LEVEL_UNSPECIFIED
}

func (x *BackintConfiguration) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *BackintConfiguration) GetFunction() Function {
	if x != nil {
		return x.Function
	}
	return Function_FUNCTION_UNSPECIFIED
}

func (x *BackintConfiguration) GetInputFile() string {
	if x != nil {
		return x.InputFile
	}
	return ""
}

func (x *BackintConfiguration) GetOutputFile() string {
	if x != nil {
		return x.OutputFile
	}
	return ""
}

func (x *BackintConfiguration) GetParamFile() string {
	if x != nil {
		return x.ParamFile
	}
	return ""
}

func (x *BackintConfiguration) GetBackupId() string {
	if x != nil {
		return x.BackupId
	}
	return ""
}

func (x *BackintConfiguration) GetDatabaseObjectCount() int64 {
	if x != nil {
		return x.DatabaseObjectCount
	}
	return 0
}

func (x *BackintConfiguration) GetBackupLevel() string {
	if x != nil {
		return x.BackupLevel
	}
	return ""
}

func (x *BackintConfiguration) GetLogDelaySec() int64 {
	if x != nil {
		return x.LogDelaySec
	}
	return 0
}

func (x *BackintConfiguration) GetLogToCloud() *wrappers.BoolValue {
	if x != nil {
		return x.LogToCloud
	}
	return nil
}

func (x *BackintConfiguration) GetRecoveryBucket() string {
	if x != nil {
		return x.RecoveryBucket
	}
	return ""
}

func (x *BackintConfiguration) GetRetryBackoffInitial() int64 {
	if x != nil {
		return x.RetryBackoffInitial
	}
	return 0
}

func (x *BackintConfiguration) GetRetryBackoffMax() int64 {
	if x != nil {
		return x.RetryBackoffMax
	}
	return 0
}

func (x *BackintConfiguration) GetRetryBackoffMultiplier() float32 {
	if x != nil {
		return x.RetryBackoffMultiplier
	}
	return 0
}

func (x *BackintConfiguration) GetClientEndpoint() string {
	if x != nil {
		return x.ClientEndpoint
	}
	return ""
}

func (x *BackintConfiguration) GetFolderPrefix() string {
	if x != nil {
		return x.FolderPrefix
	}
	return ""
}

func (x *BackintConfiguration) GetRecoveryFolderPrefix() string {
	if x != nil {
		return x.RecoveryFolderPrefix
	}
	return ""
}

func (x *BackintConfiguration) GetXmlMultipartUpload() bool {
	if x != nil {
		return x.XmlMultipartUpload
	}
	return false
}

func (x *BackintConfiguration) GetStorageClass() StorageClass {
	if x != nil {
		return x.StorageClass
	}
	return StorageClass_STORAGE_CLASS_UNSPECIFIED
}

func (x *BackintConfiguration) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *BackintConfiguration) GetDiagnoseFileMaxSizeGb() int64 {
	if x != nil {
		return x.DiagnoseFileMaxSizeGb
	}
	return 0
}

func (x *BackintConfiguration) GetSendMetricsToMonitoring() *wrappers.BoolValue {
	if x != nil {
		return x.SendMetricsToMonitoring
	}
	return nil
}

func (x *BackintConfiguration) GetShortenFolderPath() bool {
	if x != nil {
		return x.ShortenFolderPath
	}
	return false
}

func (x *BackintConfiguration) GetDiagnoseTmpDirectory() string {
	if x != nil {
		return x.DiagnoseTmpDirectory
	}
	return ""
}

func (x *BackintConfiguration) GetCustomTime() string {
	if x != nil {
		return x.CustomTime
	}
	return ""
}

func (x *BackintConfiguration) GetParallelRecoveryStreams() int64 {
	if x != nil {
		return x.ParallelRecoveryStreams
	}
	return 0
}

var File_backint_backint_proto protoreflect.FileDescriptor

var file_backint_backint_proto_rawDesc = []byte{
	0x0a, 0x15, 0x62, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x74, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x69, 0x6e,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x74,
	0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xaf, 0x0e, 0x0a, 0x14, 0x42, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x75, 0x63,
	0x6b, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x07, 0x72, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73, 0x12, 0x29, 0x0a, 0x10, 0x70,
	0x61, 0x72, 0x61, 0x6c, 0x6c, 0x65, 0x6c, 0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x70, 0x61, 0x72, 0x61, 0x6c, 0x6c, 0x65, 0x6c, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x68, 0x72, 0x65, 0x61, 0x64,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x74, 0x68, 0x72, 0x65, 0x61, 0x64, 0x73,
	0x12, 0x24, 0x0a, 0x0e, 0x62, 0x75, 0x66, 0x66, 0x65, 0x72, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x5f,
	0x6d, 0x62, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x62, 0x75, 0x66, 0x66, 0x65, 0x72,
	0x53, 0x69, 0x7a, 0x65, 0x4d, 0x62, 0x12, 0x25, 0x0a, 0x0e, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d,
	0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x4b, 0x65, 0x79, 0x12, 0x1a, 0x0a,
	0x08, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x08, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x12, 0x17, 0x0a, 0x07, 0x6b, 0x6d, 0x73,
	0x5f, 0x6b, 0x65, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6b, 0x6d, 0x73, 0x4b,
	0x65, 0x79, 0x12, 0x2e, 0x0a, 0x13, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x11, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4b,
	0x65, 0x79, 0x12, 0x22, 0x0a, 0x0d, 0x72, 0x61, 0x74, 0x65, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74,
	0x5f, 0x6d, 0x62, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x72, 0x61, 0x74, 0x65, 0x4c,
	0x69, 0x6d, 0x69, 0x74, 0x4d, 0x62, 0x12, 0x2f, 0x0a, 0x14, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x72,
	0x65, 0x61, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x6d, 0x73, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x66, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x61, 0x64, 0x54, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x64, 0x75, 0x6d, 0x70, 0x5f,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x64, 0x75, 0x6d, 0x70,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x3e, 0x0a, 0x09, 0x6c, 0x6f, 0x67, 0x5f, 0x6c, 0x65, 0x76, 0x65,
	0x6c, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65,
	0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x69, 0x6e,
	0x74, 0x2e, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x52, 0x08, 0x6c, 0x6f, 0x67, 0x4c,
	0x65, 0x76, 0x65, 0x6c, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18,
	0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x3d, 0x0a,
	0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x21, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x74, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1d, 0x0a, 0x0a,
	0x69, 0x6e, 0x70, 0x75, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6f,
	0x75, 0x74, 0x70, 0x75, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x1d, 0x0a, 0x0a,
	0x70, 0x61, 0x72, 0x61, 0x6d, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x12, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x62,
	0x61, 0x63, 0x6b, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x49, 0x64, 0x12, 0x32, 0x0a, 0x15, 0x64, 0x61, 0x74, 0x61,
	0x62, 0x61, 0x73, 0x65, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x18, 0x14, 0x20, 0x01, 0x28, 0x03, 0x52, 0x13, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73,
	0x65, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x21, 0x0a, 0x0c,
	0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x15, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12,
	0x22, 0x0a, 0x0d, 0x6c, 0x6f, 0x67, 0x5f, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x5f, 0x73, 0x65, 0x63,
	0x18, 0x16, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x6c, 0x6f, 0x67, 0x44, 0x65, 0x6c, 0x61, 0x79,
	0x53, 0x65, 0x63, 0x12, 0x3c, 0x0a, 0x0c, 0x6c, 0x6f, 0x67, 0x5f, 0x74, 0x6f, 0x5f, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x18, 0x17, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f, 0x6f, 0x6c,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0a, 0x6c, 0x6f, 0x67, 0x54, 0x6f, 0x43, 0x6c, 0x6f, 0x75,
	0x64, 0x12, 0x27, 0x0a, 0x0f, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x5f, 0x62, 0x75,
	0x63, 0x6b, 0x65, 0x74, 0x18, 0x18, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x72, 0x65, 0x63, 0x6f,
	0x76, 0x65, 0x72, 0x79, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x32, 0x0a, 0x15, 0x72, 0x65,
	0x74, 0x72, 0x79, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x5f, 0x69, 0x6e, 0x69, 0x74,
	0x69, 0x61, 0x6c, 0x18, 0x19, 0x20, 0x01, 0x28, 0x03, 0x52, 0x13, 0x72, 0x65, 0x74, 0x72, 0x79,
	0x42, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x12, 0x2a,
	0x0a, 0x11, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x5f,
	0x6d, 0x61, 0x78, 0x18, 0x1a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x72, 0x65, 0x74, 0x72, 0x79,
	0x42, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x4d, 0x61, 0x78, 0x12, 0x38, 0x0a, 0x18, 0x72, 0x65,
	0x74, 0x72, 0x79, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x5f, 0x6d, 0x75, 0x6c, 0x74,
	0x69, 0x70, 0x6c, 0x69, 0x65, 0x72, 0x18, 0x1b, 0x20, 0x01, 0x28, 0x02, 0x52, 0x16, 0x72, 0x65,
	0x74, 0x72, 0x79, 0x42, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x70,
	0x6c, 0x69, 0x65, 0x72, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x1c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x23, 0x0a,
	0x0d, 0x66, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x1d,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x66, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x50, 0x72, 0x65, 0x66,
	0x69, 0x78, 0x12, 0x34, 0x0a, 0x16, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x5f, 0x66,
	0x6f, 0x6c, 0x64, 0x65, 0x72, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x1e, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x14, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x46, 0x6f, 0x6c, 0x64,
	0x65, 0x72, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x30, 0x0a, 0x14, 0x78, 0x6d, 0x6c, 0x5f,
	0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x61, 0x72, 0x74, 0x5f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x18, 0x1f, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x78, 0x6d, 0x6c, 0x4d, 0x75, 0x6c, 0x74, 0x69,
	0x70, 0x61, 0x72, 0x74, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x4a, 0x0a, 0x0d, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x18, 0x20, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x25, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x74, 0x2e, 0x53, 0x74, 0x6f, 0x72,
	0x61, 0x67, 0x65, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x52, 0x0c, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x12, 0x57, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x21, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3b, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67,
	0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x69,
	0x6e, 0x74, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x38, 0x0a, 0x19, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x65, 0x5f, 0x66, 0x69, 0x6c, 0x65,
	0x5f, 0x6d, 0x61, 0x78, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x5f, 0x67, 0x62, 0x18, 0x22, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x15, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x65, 0x46, 0x69, 0x6c, 0x65,
	0x4d, 0x61, 0x78, 0x53, 0x69, 0x7a, 0x65, 0x47, 0x62, 0x12, 0x57, 0x0a, 0x1a, 0x73, 0x65, 0x6e,
	0x64, 0x5f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x5f, 0x74, 0x6f, 0x5f, 0x6d, 0x6f, 0x6e,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x23, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x42, 0x6f, 0x6f, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x17, 0x73, 0x65, 0x6e, 0x64, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x54, 0x6f, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69,
	0x6e, 0x67, 0x12, 0x2e, 0x0a, 0x13, 0x73, 0x68, 0x6f, 0x72, 0x74, 0x65, 0x6e, 0x5f, 0x66, 0x6f,
	0x6c, 0x64, 0x65, 0x72, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x24, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x11, 0x73, 0x68, 0x6f, 0x72, 0x74, 0x65, 0x6e, 0x46, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x50, 0x61,
	0x74, 0x68, 0x12, 0x34, 0x0a, 0x16, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x65, 0x5f, 0x74,
	0x6d, 0x70, 0x5f, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x25, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x14, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x65, 0x54, 0x6d, 0x70, 0x44,
	0x69, 0x72, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x75, 0x73, 0x74,
	0x6f, 0x6d, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x26, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63,
	0x75, 0x73, 0x74, 0x6f, 0x6d, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x3a, 0x0a, 0x19, 0x70, 0x61, 0x72,
	0x61, 0x6c, 0x6c, 0x65, 0x6c, 0x5f, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x5f, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x18, 0x27, 0x20, 0x01, 0x28, 0x03, 0x52, 0x17, 0x70, 0x61,
	0x72, 0x61, 0x6c, 0x6c, 0x65, 0x6c, 0x52, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x73, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x2a, 0x52, 0x0a, 0x08, 0x4c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x19,
	0x0a, 0x15, 0x4c, 0x4f, 0x47, 0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x44, 0x45, 0x42,
	0x55, 0x47, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x49, 0x4e, 0x46, 0x4f, 0x10, 0x02, 0x12, 0x0b,
	0x0a, 0x07, 0x57, 0x41, 0x52, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x45,
	0x52, 0x52, 0x4f, 0x52, 0x10, 0x04, 0x2a, 0x64, 0x0a, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x14, 0x46, 0x55, 0x4e, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06,
	0x42, 0x41, 0x43, 0x4b, 0x55, 0x50, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x45, 0x53, 0x54,
	0x4f, 0x52, 0x45, 0x10, 0x02, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e, 0x51, 0x55, 0x49, 0x52, 0x45,
	0x10, 0x03, 0x12, 0x0a, 0x0a, 0x06, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x10, 0x04, 0x12, 0x0c,
	0x0a, 0x08, 0x44, 0x49, 0x41, 0x47, 0x4e, 0x4f, 0x53, 0x45, 0x10, 0x05, 0x2a, 0x64, 0x0a, 0x0c,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x12, 0x1d, 0x0a, 0x19,
	0x53, 0x54, 0x4f, 0x52, 0x41, 0x47, 0x45, 0x5f, 0x43, 0x4c, 0x41, 0x53, 0x53, 0x5f, 0x55, 0x4e,
	0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x53,
	0x54, 0x41, 0x4e, 0x44, 0x41, 0x52, 0x44, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x4e, 0x45, 0x41,
	0x52, 0x4c, 0x49, 0x4e, 0x45, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x4f, 0x4c, 0x44, 0x4c,
	0x49, 0x4e, 0x45, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x41, 0x52, 0x43, 0x48, 0x49, 0x56, 0x45,
	0x10, 0x04, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_backint_backint_proto_rawDescOnce sync.Once
	file_backint_backint_proto_rawDescData = file_backint_backint_proto_rawDesc
)

func file_backint_backint_proto_rawDescGZIP() []byte {
	file_backint_backint_proto_rawDescOnce.Do(func() {
		file_backint_backint_proto_rawDescData = protoimpl.X.CompressGZIP(file_backint_backint_proto_rawDescData)
	})
	return file_backint_backint_proto_rawDescData
}

var file_backint_backint_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_backint_backint_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_backint_backint_proto_goTypes = []any{
	(LogLevel)(0),                // 0: sapagent.protos.backint.LogLevel
	(Function)(0),                // 1: sapagent.protos.backint.Function
	(StorageClass)(0),            // 2: sapagent.protos.backint.StorageClass
	(*BackintConfiguration)(nil), // 3: sapagent.protos.backint.BackintConfiguration
	nil,                          // 4: sapagent.protos.backint.BackintConfiguration.MetadataEntry
	(*wrappers.BoolValue)(nil),   // 5: google.protobuf.BoolValue
}
var file_backint_backint_proto_depIdxs = []int32{
	0, // 0: sapagent.protos.backint.BackintConfiguration.log_level:type_name -> sapagent.protos.backint.LogLevel
	1, // 1: sapagent.protos.backint.BackintConfiguration.function:type_name -> sapagent.protos.backint.Function
	5, // 2: sapagent.protos.backint.BackintConfiguration.log_to_cloud:type_name -> google.protobuf.BoolValue
	2, // 3: sapagent.protos.backint.BackintConfiguration.storage_class:type_name -> sapagent.protos.backint.StorageClass
	4, // 4: sapagent.protos.backint.BackintConfiguration.metadata:type_name -> sapagent.protos.backint.BackintConfiguration.MetadataEntry
	5, // 5: sapagent.protos.backint.BackintConfiguration.send_metrics_to_monitoring:type_name -> google.protobuf.BoolValue
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_backint_backint_proto_init() }
func file_backint_backint_proto_init() {
	if File_backint_backint_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_backint_backint_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*BackintConfiguration); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_backint_backint_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_backint_backint_proto_goTypes,
		DependencyIndexes: file_backint_backint_proto_depIdxs,
		EnumInfos:         file_backint_backint_proto_enumTypes,
		MessageInfos:      file_backint_backint_proto_msgTypes,
	}.Build()
	File_backint_backint_proto = out.File
	file_backint_backint_proto_rawDesc = nil
	file_backint_backint_proto_goTypes = nil
	file_backint_backint_proto_depIdxs = nil
}
