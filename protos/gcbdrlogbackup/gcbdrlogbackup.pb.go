//
//Copyright 2025 Google LLC
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
// 	protoc-gen-go v1.33.0
// 	protoc        v4.23.4
// source: protos/gcbdrlogbackup/gcbdrlogbackup.proto

package gcbdrlogbackup

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LogBackupResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status *wrapperspb.BoolValue `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	JobId  string                `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (x *LogBackupResponse) Reset() {
	*x = LogBackupResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogBackupResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogBackupResponse) ProtoMessage() {}

func (x *LogBackupResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogBackupResponse.ProtoReflect.Descriptor instead.
func (*LogBackupResponse) Descriptor() ([]byte, []int) {
	return file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescGZIP(), []int{0}
}

func (x *LogBackupResponse) GetStatus() *wrapperspb.BoolValue {
	if x != nil {
		return x.Status
	}
	return nil
}

func (x *LogBackupResponse) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

var File_protos_gcbdrlogbackup_gcbdrlogbackup_proto protoreflect.FileDescriptor

var file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x67, 0x63, 0x62, 0x64, 0x72, 0x6c, 0x6f,
	0x67, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x2f, 0x67, 0x63, 0x62, 0x64, 0x72, 0x6c, 0x6f, 0x67,
	0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1e, 0x73, 0x61,
	0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x67, 0x63,
	0x62, 0x64, 0x72, 0x6c, 0x6f, 0x67, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x1a, 0x1e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72,
	0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5e, 0x0a, 0x11,
	0x4c, 0x6f, 0x67, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x32, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x42, 0x3f, 0x5a, 0x3d,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x47, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x73,
	0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x67,
	0x63, 0x62, 0x64, 0x72, 0x6c, 0x6f, 0x67, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescOnce sync.Once
	file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescData = file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDesc
)

func file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescGZIP() []byte {
	file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescOnce.Do(func() {
		file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescData)
	})
	return file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDescData
}

var file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_goTypes = []interface{}{
	(*LogBackupResponse)(nil),    // 0: sapagent.protos.gcbdrlogbackup.LogBackupResponse
	(*wrapperspb.BoolValue)(nil), // 1: google.protobuf.BoolValue
}
var file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_depIdxs = []int32{
	1, // 0: sapagent.protos.gcbdrlogbackup.LogBackupResponse.status:type_name -> google.protobuf.BoolValue
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_init() }
func file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_init() {
	if File_protos_gcbdrlogbackup_gcbdrlogbackup_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogBackupResponse); i {
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
			RawDescriptor: file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_goTypes,
		DependencyIndexes: file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_depIdxs,
		MessageInfos:      file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_msgTypes,
	}.Build()
	File_protos_gcbdrlogbackup_gcbdrlogbackup_proto = out.File
	file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_rawDesc = nil
	file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_goTypes = nil
	file_protos_gcbdrlogbackup_gcbdrlogbackup_proto_depIdxs = nil
}
