//
//Copyright 2024 Google LLC
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
// 	protoc        v3.6.1
// source: guestactions/guestactions.proto

package guestactions

import (
	_ "github.com/golang/protobuf/ptypes/wrappers"
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

type Command_CommandType int32

const (
	Command_COMMAND_TYPE_UNSPECIFIED Command_CommandType = 0
	Command_SHELL                    Command_CommandType = 1
	Command_VERSION                  Command_CommandType = 2
)

// Enum value maps for Command_CommandType.
var (
	Command_CommandType_name = map[int32]string{
		0: "COMMAND_TYPE_UNSPECIFIED",
		1: "SHELL",
		2: "VERSION",
	}
	Command_CommandType_value = map[string]int32{
		"COMMAND_TYPE_UNSPECIFIED": 0,
		"SHELL":                    1,
		"VERSION":                  2,
	}
)

func (x Command_CommandType) Enum() *Command_CommandType {
	p := new(Command_CommandType)
	*p = x
	return p
}

func (x Command_CommandType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Command_CommandType) Descriptor() protoreflect.EnumDescriptor {
	return file_guestactions_guestactions_proto_enumTypes[0].Descriptor()
}

func (Command_CommandType) Type() protoreflect.EnumType {
	return &file_guestactions_guestactions_proto_enumTypes[0]
}

func (x Command_CommandType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Command_CommandType.Descriptor instead.
func (Command_CommandType) EnumDescriptor() ([]byte, []int) {
	return file_guestactions_guestactions_proto_rawDescGZIP(), []int{1, 0}
}

type GuestAction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Commands []*Command `protobuf:"bytes,1,rep,name=commands,proto3" json:"commands,omitempty"`
}

func (x *GuestAction) Reset() {
	*x = GuestAction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_guestactions_guestactions_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GuestAction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GuestAction) ProtoMessage() {}

func (x *GuestAction) ProtoReflect() protoreflect.Message {
	mi := &file_guestactions_guestactions_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GuestAction.ProtoReflect.Descriptor instead.
func (*GuestAction) Descriptor() ([]byte, []int) {
	return file_guestactions_guestactions_proto_rawDescGZIP(), []int{0}
}

func (x *GuestAction) GetCommands() []*Command {
	if x != nil {
		return x.Commands
	}
	return nil
}

type Command struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CommandType Command_CommandType `protobuf:"varint,1,opt,name=command_type,json=commandType,proto3,enum=sapagent.protos.guestactions.Command_CommandType" json:"command_type,omitempty"`
	Parameters  string              `protobuf:"bytes,2,opt,name=parameters,proto3" json:"parameters,omitempty"`
}

func (x *Command) Reset() {
	*x = Command{}
	if protoimpl.UnsafeEnabled {
		mi := &file_guestactions_guestactions_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Command) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Command) ProtoMessage() {}

func (x *Command) ProtoReflect() protoreflect.Message {
	mi := &file_guestactions_guestactions_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Command.ProtoReflect.Descriptor instead.
func (*Command) Descriptor() ([]byte, []int) {
	return file_guestactions_guestactions_proto_rawDescGZIP(), []int{1}
}

func (x *Command) GetCommandType() Command_CommandType {
	if x != nil {
		return x.CommandType
	}
	return Command_COMMAND_TYPE_UNSPECIFIED
}

func (x *Command) GetParameters() string {
	if x != nil {
		return x.Parameters
	}
	return ""
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CommandResults []*CommandResult `protobuf:"bytes,1,rep,name=command_results,json=commandResults,proto3" json:"command_results,omitempty"`
	ErrorMessage   string           `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_guestactions_guestactions_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_guestactions_guestactions_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_guestactions_guestactions_proto_rawDescGZIP(), []int{2}
}

func (x *Response) GetCommandResults() []*CommandResult {
	if x != nil {
		return x.CommandResults
	}
	return nil
}

func (x *Response) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

type CommandResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Output string `protobuf:"bytes,1,opt,name=output,proto3" json:"output,omitempty"`
}

func (x *CommandResult) Reset() {
	*x = CommandResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_guestactions_guestactions_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommandResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommandResult) ProtoMessage() {}

func (x *CommandResult) ProtoReflect() protoreflect.Message {
	mi := &file_guestactions_guestactions_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommandResult.ProtoReflect.Descriptor instead.
func (*CommandResult) Descriptor() ([]byte, []int) {
	return file_guestactions_guestactions_proto_rawDescGZIP(), []int{3}
}

func (x *CommandResult) GetOutput() string {
	if x != nil {
		return x.Output
	}
	return ""
}

var File_guestactions_guestactions_proto protoreflect.FileDescriptor

var file_guestactions_guestactions_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x67, 0x75, 0x65, 0x73, 0x74, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x67,
	0x75, 0x65, 0x73, 0x74, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x1c, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2e, 0x67, 0x75, 0x65, 0x73, 0x74, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a,
	0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x50, 0x0a, 0x0b, 0x47, 0x75, 0x65, 0x73, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x41,
	0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x25, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2e, 0x67, 0x75, 0x65, 0x73, 0x74, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x73, 0x22, 0xc4, 0x01, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x54, 0x0a,
	0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x31, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x67, 0x75, 0x65, 0x73, 0x74, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61,
	0x6e, 0x64, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74,
	0x65, 0x72, 0x73, 0x22, 0x43, 0x0a, 0x0b, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x1c, 0x0a, 0x18, 0x43, 0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00,
	0x12, 0x09, 0x0a, 0x05, 0x53, 0x48, 0x45, 0x4c, 0x4c, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x56,
	0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x10, 0x02, 0x22, 0x85, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x54, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b,
	0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x67, 0x75, 0x65, 0x73, 0x74, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x43, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x0e, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x22, 0x27, 0x0a, 0x0d, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_guestactions_guestactions_proto_rawDescOnce sync.Once
	file_guestactions_guestactions_proto_rawDescData = file_guestactions_guestactions_proto_rawDesc
)

func file_guestactions_guestactions_proto_rawDescGZIP() []byte {
	file_guestactions_guestactions_proto_rawDescOnce.Do(func() {
		file_guestactions_guestactions_proto_rawDescData = protoimpl.X.CompressGZIP(file_guestactions_guestactions_proto_rawDescData)
	})
	return file_guestactions_guestactions_proto_rawDescData
}

var file_guestactions_guestactions_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_guestactions_guestactions_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_guestactions_guestactions_proto_goTypes = []interface{}{
	(Command_CommandType)(0), // 0: sapagent.protos.guestactions.Command.CommandType
	(*GuestAction)(nil),      // 1: sapagent.protos.guestactions.GuestAction
	(*Command)(nil),          // 2: sapagent.protos.guestactions.Command
	(*Response)(nil),         // 3: sapagent.protos.guestactions.Response
	(*CommandResult)(nil),    // 4: sapagent.protos.guestactions.CommandResult
}
var file_guestactions_guestactions_proto_depIdxs = []int32{
	2, // 0: sapagent.protos.guestactions.GuestAction.commands:type_name -> sapagent.protos.guestactions.Command
	0, // 1: sapagent.protos.guestactions.Command.command_type:type_name -> sapagent.protos.guestactions.Command.CommandType
	4, // 2: sapagent.protos.guestactions.Response.command_results:type_name -> sapagent.protos.guestactions.CommandResult
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_guestactions_guestactions_proto_init() }
func file_guestactions_guestactions_proto_init() {
	if File_guestactions_guestactions_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_guestactions_guestactions_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GuestAction); i {
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
		file_guestactions_guestactions_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Command); i {
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
		file_guestactions_guestactions_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
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
		file_guestactions_guestactions_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommandResult); i {
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
			RawDescriptor: file_guestactions_guestactions_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_guestactions_guestactions_proto_goTypes,
		DependencyIndexes: file_guestactions_guestactions_proto_depIdxs,
		EnumInfos:         file_guestactions_guestactions_proto_enumTypes,
		MessageInfos:      file_guestactions_guestactions_proto_msgTypes,
	}.Build()
	File_guestactions_guestactions_proto = out.File
	file_guestactions_guestactions_proto_rawDesc = nil
	file_guestactions_guestactions_proto_goTypes = nil
	file_guestactions_guestactions_proto_depIdxs = nil
}