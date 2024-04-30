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
// 	protoc-gen-go v1.34.0
// 	protoc        v3.6.1
// source: yamlpbtest/yamlpbtest.proto

package yamlpbtest

import (
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

type TestEnum int32

const (
	TestEnum_VAL_UNSPECIFIED TestEnum = 0
	TestEnum_VAL_ONE         TestEnum = 1
	TestEnum_VAL_TWO         TestEnum = 2
)

// Enum value maps for TestEnum.
var (
	TestEnum_name = map[int32]string{
		0: "VAL_UNSPECIFIED",
		1: "VAL_ONE",
		2: "VAL_TWO",
	}
	TestEnum_value = map[string]int32{
		"VAL_UNSPECIFIED": 0,
		"VAL_ONE":         1,
		"VAL_TWO":         2,
	}
)

func (x TestEnum) Enum() *TestEnum {
	p := new(TestEnum)
	*p = x
	return p
}

func (x TestEnum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TestEnum) Descriptor() protoreflect.EnumDescriptor {
	return file_yamlpbtest_yamlpbtest_proto_enumTypes[0].Descriptor()
}

func (TestEnum) Type() protoreflect.EnumType {
	return &file_yamlpbtest_yamlpbtest_proto_enumTypes[0]
}

func (x TestEnum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TestEnum.Descriptor instead.
func (TestEnum) EnumDescriptor() ([]byte, []int) {
	return file_yamlpbtest_yamlpbtest_proto_rawDescGZIP(), []int{0}
}

type NestedTestMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uint32V uint32 `protobuf:"varint,1,opt,name=uint32_v,json=uint32V,proto3" json:"uint32_v,omitempty"`
}

func (x *NestedTestMessage) Reset() {
	*x = NestedTestMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_yamlpbtest_yamlpbtest_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NestedTestMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NestedTestMessage) ProtoMessage() {}

func (x *NestedTestMessage) ProtoReflect() protoreflect.Message {
	mi := &file_yamlpbtest_yamlpbtest_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NestedTestMessage.ProtoReflect.Descriptor instead.
func (*NestedTestMessage) Descriptor() ([]byte, []int) {
	return file_yamlpbtest_yamlpbtest_proto_rawDescGZIP(), []int{0}
}

func (x *NestedTestMessage) GetUint32V() uint32 {
	if x != nil {
		return x.Uint32V
	}
	return 0
}

type OtherNestedTestMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StringV string `protobuf:"bytes,1,opt,name=string_v,json=stringV,proto3" json:"string_v,omitempty"`
}

func (x *OtherNestedTestMessage) Reset() {
	*x = OtherNestedTestMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_yamlpbtest_yamlpbtest_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OtherNestedTestMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OtherNestedTestMessage) ProtoMessage() {}

func (x *OtherNestedTestMessage) ProtoReflect() protoreflect.Message {
	mi := &file_yamlpbtest_yamlpbtest_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OtherNestedTestMessage.ProtoReflect.Descriptor instead.
func (*OtherNestedTestMessage) Descriptor() ([]byte, []int) {
	return file_yamlpbtest_yamlpbtest_proto_rawDescGZIP(), []int{1}
}

func (x *OtherNestedTestMessage) GetStringV() string {
	if x != nil {
		return x.StringV
	}
	return ""
}

type TestMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uint32V uint32               `protobuf:"varint,1,opt,name=uint32_v,json=uint32V,proto3" json:"uint32_v,omitempty"`
	FloatV  float32              `protobuf:"fixed32,2,opt,name=float_v,json=floatV,proto3" json:"float_v,omitempty"`
	StringV string               `protobuf:"bytes,3,opt,name=string_v,json=stringV,proto3" json:"string_v,omitempty"`
	BoolV   bool                 `protobuf:"varint,4,opt,name=bool_v,json=boolV,proto3" json:"bool_v,omitempty"`
	EnumV   TestEnum             `protobuf:"varint,5,opt,name=enum_v,json=enumV,proto3,enum=sapagent.protos.yamlpbtest.TestEnum" json:"enum_v,omitempty"`
	NestedV *NestedTestMessage   `protobuf:"bytes,6,opt,name=nested_v,json=nestedV,proto3" json:"nested_v,omitempty"`
	Uint32R []uint32             `protobuf:"varint,7,rep,packed,name=uint32_r,json=uint32R,proto3" json:"uint32_r,omitempty"`
	NestedR []*NestedTestMessage `protobuf:"bytes,8,rep,name=nested_r,json=nestedR,proto3" json:"nested_r,omitempty"`
}

func (x *TestMessage) Reset() {
	*x = TestMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_yamlpbtest_yamlpbtest_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestMessage) ProtoMessage() {}

func (x *TestMessage) ProtoReflect() protoreflect.Message {
	mi := &file_yamlpbtest_yamlpbtest_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestMessage.ProtoReflect.Descriptor instead.
func (*TestMessage) Descriptor() ([]byte, []int) {
	return file_yamlpbtest_yamlpbtest_proto_rawDescGZIP(), []int{2}
}

func (x *TestMessage) GetUint32V() uint32 {
	if x != nil {
		return x.Uint32V
	}
	return 0
}

func (x *TestMessage) GetFloatV() float32 {
	if x != nil {
		return x.FloatV
	}
	return 0
}

func (x *TestMessage) GetStringV() string {
	if x != nil {
		return x.StringV
	}
	return ""
}

func (x *TestMessage) GetBoolV() bool {
	if x != nil {
		return x.BoolV
	}
	return false
}

func (x *TestMessage) GetEnumV() TestEnum {
	if x != nil {
		return x.EnumV
	}
	return TestEnum_VAL_UNSPECIFIED
}

func (x *TestMessage) GetNestedV() *NestedTestMessage {
	if x != nil {
		return x.NestedV
	}
	return nil
}

func (x *TestMessage) GetUint32R() []uint32 {
	if x != nil {
		return x.Uint32R
	}
	return nil
}

func (x *TestMessage) GetNestedR() []*NestedTestMessage {
	if x != nil {
		return x.NestedR
	}
	return nil
}

var File_yamlpbtest_yamlpbtest_proto protoreflect.FileDescriptor

var file_yamlpbtest_yamlpbtest_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x79, 0x61, 0x6d, 0x6c, 0x70, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x79, 0x61, 0x6d,
	0x6c, 0x70, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x73,
	0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x79,
	0x61, 0x6d, 0x6c, 0x70, 0x62, 0x74, 0x65, 0x73, 0x74, 0x22, 0x2e, 0x0a, 0x11, 0x4e, 0x65, 0x73,
	0x74, 0x65, 0x64, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x19,
	0x0a, 0x08, 0x75, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x5f, 0x76, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x07, 0x75, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x22, 0x33, 0x0a, 0x16, 0x4f, 0x74, 0x68,
	0x65, 0x72, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x22, 0xdf,
	0x02, 0x0a, 0x0b, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x19,
	0x0a, 0x08, 0x75, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x5f, 0x76, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x07, 0x75, 0x69, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x12, 0x17, 0x0a, 0x07, 0x66, 0x6c, 0x6f,
	0x61, 0x74, 0x5f, 0x76, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x06, 0x66, 0x6c, 0x6f, 0x61,
	0x74, 0x56, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x12, 0x15, 0x0a,
	0x06, 0x62, 0x6f, 0x6f, 0x6c, 0x5f, 0x76, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x62,
	0x6f, 0x6f, 0x6c, 0x56, 0x12, 0x3b, 0x0a, 0x06, 0x65, 0x6e, 0x75, 0x6d, 0x5f, 0x76, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x79, 0x61, 0x6d, 0x6c, 0x70, 0x62, 0x74, 0x65, 0x73,
	0x74, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x45, 0x6e, 0x75, 0x6d, 0x52, 0x05, 0x65, 0x6e, 0x75, 0x6d,
	0x56, 0x12, 0x48, 0x0a, 0x08, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x5f, 0x76, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x79, 0x61, 0x6d, 0x6c, 0x70, 0x62, 0x74, 0x65, 0x73, 0x74,
	0x2e, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x52, 0x07, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x56, 0x12, 0x19, 0x0a, 0x08, 0x75,
	0x69, 0x6e, 0x74, 0x33, 0x32, 0x5f, 0x72, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x07, 0x75,
	0x69, 0x6e, 0x74, 0x33, 0x32, 0x52, 0x12, 0x48, 0x0a, 0x08, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64,
	0x5f, 0x72, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67,
	0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x79, 0x61, 0x6d, 0x6c, 0x70,
	0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x54, 0x65, 0x73, 0x74,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x52,
	0x2a, 0x39, 0x0a, 0x08, 0x54, 0x65, 0x73, 0x74, 0x45, 0x6e, 0x75, 0x6d, 0x12, 0x13, 0x0a, 0x0f,
	0x56, 0x41, 0x4c, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10,
	0x00, 0x12, 0x0b, 0x0a, 0x07, 0x56, 0x41, 0x4c, 0x5f, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x12, 0x0b,
	0x0a, 0x07, 0x56, 0x41, 0x4c, 0x5f, 0x54, 0x57, 0x4f, 0x10, 0x02, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_yamlpbtest_yamlpbtest_proto_rawDescOnce sync.Once
	file_yamlpbtest_yamlpbtest_proto_rawDescData = file_yamlpbtest_yamlpbtest_proto_rawDesc
)

func file_yamlpbtest_yamlpbtest_proto_rawDescGZIP() []byte {
	file_yamlpbtest_yamlpbtest_proto_rawDescOnce.Do(func() {
		file_yamlpbtest_yamlpbtest_proto_rawDescData = protoimpl.X.CompressGZIP(file_yamlpbtest_yamlpbtest_proto_rawDescData)
	})
	return file_yamlpbtest_yamlpbtest_proto_rawDescData
}

var file_yamlpbtest_yamlpbtest_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_yamlpbtest_yamlpbtest_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_yamlpbtest_yamlpbtest_proto_goTypes = []interface{}{
	(TestEnum)(0),                  // 0: sapagent.protos.yamlpbtest.TestEnum
	(*NestedTestMessage)(nil),      // 1: sapagent.protos.yamlpbtest.NestedTestMessage
	(*OtherNestedTestMessage)(nil), // 2: sapagent.protos.yamlpbtest.OtherNestedTestMessage
	(*TestMessage)(nil),            // 3: sapagent.protos.yamlpbtest.TestMessage
}
var file_yamlpbtest_yamlpbtest_proto_depIdxs = []int32{
	0, // 0: sapagent.protos.yamlpbtest.TestMessage.enum_v:type_name -> sapagent.protos.yamlpbtest.TestEnum
	1, // 1: sapagent.protos.yamlpbtest.TestMessage.nested_v:type_name -> sapagent.protos.yamlpbtest.NestedTestMessage
	1, // 2: sapagent.protos.yamlpbtest.TestMessage.nested_r:type_name -> sapagent.protos.yamlpbtest.NestedTestMessage
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_yamlpbtest_yamlpbtest_proto_init() }
func file_yamlpbtest_yamlpbtest_proto_init() {
	if File_yamlpbtest_yamlpbtest_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_yamlpbtest_yamlpbtest_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NestedTestMessage); i {
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
		file_yamlpbtest_yamlpbtest_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OtherNestedTestMessage); i {
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
		file_yamlpbtest_yamlpbtest_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestMessage); i {
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
			RawDescriptor: file_yamlpbtest_yamlpbtest_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_yamlpbtest_yamlpbtest_proto_goTypes,
		DependencyIndexes: file_yamlpbtest_yamlpbtest_proto_depIdxs,
		EnumInfos:         file_yamlpbtest_yamlpbtest_proto_enumTypes,
		MessageInfos:      file_yamlpbtest_yamlpbtest_proto_msgTypes,
	}.Build()
	File_yamlpbtest_yamlpbtest_proto = out.File
	file_yamlpbtest_yamlpbtest_proto_rawDesc = nil
	file_yamlpbtest_yamlpbtest_proto_goTypes = nil
	file_yamlpbtest_yamlpbtest_proto_depIdxs = nil
}
