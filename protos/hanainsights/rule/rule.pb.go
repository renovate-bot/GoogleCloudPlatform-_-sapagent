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
// source: hanainsights/rule/rule.proto

package rule

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

type EvalNode_EvalType int32

const (
	EvalNode_UNDEFINED EvalNode_EvalType = 0
	EvalNode_OR        EvalNode_EvalType = 1
	EvalNode_AND       EvalNode_EvalType = 2
	EvalNode_EQ        EvalNode_EvalType = 3
	EvalNode_NEQ       EvalNode_EvalType = 4
	EvalNode_LT        EvalNode_EvalType = 5
	EvalNode_LTE       EvalNode_EvalType = 6
	EvalNode_GT        EvalNode_EvalType = 7
	EvalNode_GTE       EvalNode_EvalType = 8
)

// Enum value maps for EvalNode_EvalType.
var (
	EvalNode_EvalType_name = map[int32]string{
		0: "UNDEFINED",
		1: "OR",
		2: "AND",
		3: "EQ",
		4: "NEQ",
		5: "LT",
		6: "LTE",
		7: "GT",
		8: "GTE",
	}
	EvalNode_EvalType_value = map[string]int32{
		"UNDEFINED": 0,
		"OR":        1,
		"AND":       2,
		"EQ":        3,
		"NEQ":       4,
		"LT":        5,
		"LTE":       6,
		"GT":        7,
		"GTE":       8,
	}
)

func (x EvalNode_EvalType) Enum() *EvalNode_EvalType {
	p := new(EvalNode_EvalType)
	*p = x
	return p
}

func (x EvalNode_EvalType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EvalNode_EvalType) Descriptor() protoreflect.EnumDescriptor {
	return file_hanainsights_rule_rule_proto_enumTypes[0].Descriptor()
}

func (EvalNode_EvalType) Type() protoreflect.EnumType {
	return &file_hanainsights_rule_rule_proto_enumTypes[0]
}

func (x EvalNode_EvalType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EvalNode_EvalType.Descriptor instead.
func (EvalNode_EvalType) EnumDescriptor() ([]byte, []int) {
	return file_hanainsights_rule_rule_proto_rawDescGZIP(), []int{3, 0}
}

type Rule struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`     // Optional
	Id     string   `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`         // Required: Unique ID of the rule - must be unique across all rules.
	Labels []string `protobuf:"bytes,3,rep,name=labels,proto3" json:"labels,omitempty"` // Security, High Availability, performance,
	// cost-saving, supportability, reliability, etc.
	Description     string            `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Queries         []*Query          `protobuf:"bytes,5,rep,name=queries,proto3" json:"queries,omitempty"`
	Recommendations []*Recommendation `protobuf:"bytes,6,rep,name=recommendations,proto3" json:"recommendations,omitempty"`
}

func (x *Rule) Reset() {
	*x = Rule{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hanainsights_rule_rule_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Rule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Rule) ProtoMessage() {}

func (x *Rule) ProtoReflect() protoreflect.Message {
	mi := &file_hanainsights_rule_rule_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Rule.ProtoReflect.Descriptor instead.
func (*Rule) Descriptor() ([]byte, []int) {
	return file_hanainsights_rule_rule_proto_rawDescGZIP(), []int{0}
}

func (x *Rule) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Rule) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Rule) GetLabels() []string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *Rule) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Rule) GetQueries() []*Query {
	if x != nil {
		return x.Queries
	}
	return nil
}

func (x *Rule) GetRecommendations() []*Recommendation {
	if x != nil {
		return x.Recommendations
	}
	return nil
}

type Query struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name               string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"` //  Required: Unique within this rule and global knowledgebase.
	Description        string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	DependentOnQueries []string `protobuf:"bytes,3,rep,name=dependent_on_queries,json=dependentOnQueries,proto3" json:"dependent_on_queries,omitempty"` // names of the queries that must be run prior to this.
	Sql                string   `protobuf:"bytes,4,opt,name=sql,proto3" json:"sql,omitempty"`                                                           // SQL query
	Columns            []string `protobuf:"bytes,5,rep,name=columns,proto3" json:"columns,omitempty"`                                                   // Required: Used to build knowledgebase
}

func (x *Query) Reset() {
	*x = Query{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hanainsights_rule_rule_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Query) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Query) ProtoMessage() {}

func (x *Query) ProtoReflect() protoreflect.Message {
	mi := &file_hanainsights_rule_rule_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Query.ProtoReflect.Descriptor instead.
func (*Query) Descriptor() ([]byte, []int) {
	return file_hanainsights_rule_rule_proto_rawDescGZIP(), []int{1}
}

func (x *Query) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Query) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Query) GetDependentOnQueries() []string {
	if x != nil {
		return x.DependentOnQueries
	}
	return nil
}

func (x *Query) GetSql() string {
	if x != nil {
		return x.Sql
	}
	return ""
}

func (x *Query) GetColumns() []string {
	if x != nil {
		return x.Columns
	}
	return nil
}

type Recommendation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`               // Optional
	Id           string    `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`                   // Required: used to uniquely identify a recoomendation.
	Description  string    `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"` // Optional
	Trigger      *EvalNode `protobuf:"bytes,4,opt,name=trigger,proto3" json:"trigger,omitempty"`
	Actions      []*Action `protobuf:"bytes,5,rep,name=actions,proto3" json:"actions,omitempty"`
	ForceTrigger bool      `protobuf:"varint,6,opt,name=force_trigger,json=forceTrigger,proto3" json:"force_trigger,omitempty"` // Optional - for internal testing
	References   []string  `protobuf:"bytes,7,rep,name=references,proto3" json:"references,omitempty"`                          // Optional
}

func (x *Recommendation) Reset() {
	*x = Recommendation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hanainsights_rule_rule_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Recommendation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Recommendation) ProtoMessage() {}

func (x *Recommendation) ProtoReflect() protoreflect.Message {
	mi := &file_hanainsights_rule_rule_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Recommendation.ProtoReflect.Descriptor instead.
func (*Recommendation) Descriptor() ([]byte, []int) {
	return file_hanainsights_rule_rule_proto_rawDescGZIP(), []int{2}
}

func (x *Recommendation) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Recommendation) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Recommendation) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Recommendation) GetTrigger() *EvalNode {
	if x != nil {
		return x.Trigger
	}
	return nil
}

func (x *Recommendation) GetActions() []*Action {
	if x != nil {
		return x.Actions
	}
	return nil
}

func (x *Recommendation) GetForceTrigger() bool {
	if x != nil {
		return x.ForceTrigger
	}
	return false
}

func (x *Recommendation) GetReferences() []string {
	if x != nil {
		return x.References
	}
	return nil
}

type EvalNode struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Lhs        string            `protobuf:"bytes,1,opt,name=lhs,proto3" json:"lhs,omitempty"` // used when type is COMPARISON
	Rhs        string            `protobuf:"bytes,2,opt,name=rhs,proto3" json:"rhs,omitempty"` // used when type is COMPARISON
	Operation  EvalNode_EvalType `protobuf:"varint,3,opt,name=operation,proto3,enum=sapagent.protos.hanainsights.rule.EvalNode_EvalType" json:"operation,omitempty"`
	ChildEvals []*EvalNode       `protobuf:"bytes,4,rep,name=child_evals,json=childEvals,proto3" json:"child_evals,omitempty"` // used when type is OR, AND
}

func (x *EvalNode) Reset() {
	*x = EvalNode{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hanainsights_rule_rule_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EvalNode) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvalNode) ProtoMessage() {}

func (x *EvalNode) ProtoReflect() protoreflect.Message {
	mi := &file_hanainsights_rule_rule_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvalNode.ProtoReflect.Descriptor instead.
func (*EvalNode) Descriptor() ([]byte, []int) {
	return file_hanainsights_rule_rule_proto_rawDescGZIP(), []int{3}
}

func (x *EvalNode) GetLhs() string {
	if x != nil {
		return x.Lhs
	}
	return ""
}

func (x *EvalNode) GetRhs() string {
	if x != nil {
		return x.Rhs
	}
	return ""
}

func (x *EvalNode) GetOperation() EvalNode_EvalType {
	if x != nil {
		return x.Operation
	}
	return EvalNode_UNDEFINED
}

func (x *EvalNode) GetChildEvals() []*EvalNode {
	if x != nil {
		return x.ChildEvals
	}
	return nil
}

type Action struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Statement   string `protobuf:"bytes,3,opt,name=statement,proto3" json:"statement,omitempty"`
	Rollback    string `protobuf:"bytes,4,opt,name=rollback,proto3" json:"rollback,omitempty"` 
}

func (x *Action) Reset() {
	*x = Action{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hanainsights_rule_rule_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Action) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Action) ProtoMessage() {}

func (x *Action) ProtoReflect() protoreflect.Message {
	mi := &file_hanainsights_rule_rule_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Action.ProtoReflect.Descriptor instead.
func (*Action) Descriptor() ([]byte, []int) {
	return file_hanainsights_rule_rule_proto_rawDescGZIP(), []int{4}
}

func (x *Action) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Action) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Action) GetStatement() string {
	if x != nil {
		return x.Statement
	}
	return ""
}

func (x *Action) GetRollback() string {
	if x != nil {
		return x.Rollback
	}
	return ""
}

var File_hanainsights_rule_rule_proto protoreflect.FileDescriptor

var file_hanainsights_rule_rule_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x68, 0x61, 0x6e, 0x61, 0x69, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73, 0x2f, 0x72,
	0x75, 0x6c, 0x65, 0x2f, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x21,
	0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e,
	0x68, 0x61, 0x6e, 0x61, 0x69, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73, 0x2e, 0x72, 0x75, 0x6c,
	0x65, 0x22, 0x85, 0x02, 0x0a, 0x04, 0x52, 0x75, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16,
	0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x42, 0x0a, 0x07, 0x71, 0x75, 0x65, 0x72,
	0x69, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x73, 0x61, 0x70, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x68, 0x61, 0x6e, 0x61,
	0x69, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x51, 0x75,
	0x65, 0x72, 0x79, 0x52, 0x07, 0x71, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x12, 0x5b, 0x0a, 0x0f,
	0x72, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x68, 0x61, 0x6e, 0x61, 0x69, 0x6e, 0x73, 0x69,
	0x67, 0x68, 0x74, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x6d, 0x6d,
	0x65, 0x6e, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0f, 0x72, 0x65, 0x63, 0x6f, 0x6d, 0x6d,
	0x65, 0x6e, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x9b, 0x01, 0x0a, 0x05, 0x51, 0x75,
	0x65, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x30, 0x0a, 0x14, 0x64, 0x65, 0x70,
	0x65, 0x6e, 0x64, 0x65, 0x6e, 0x74, 0x5f, 0x6f, 0x6e, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x69, 0x65,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65,
	0x6e, 0x74, 0x4f, 0x6e, 0x51, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x73,
	0x71, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x71, 0x6c, 0x12, 0x18, 0x0a,
	0x07, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07,
	0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x22, 0xa7, 0x02, 0x0a, 0x0e, 0x52, 0x65, 0x63, 0x6f,
	0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x45, 0x0a, 0x07, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2b, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x73, 0x2e, 0x68, 0x61, 0x6e, 0x61, 0x69, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73,
	0x2e, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x07,
	0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x12, 0x43, 0x0a, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67,
	0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x68, 0x61, 0x6e, 0x61, 0x69,
	0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x41, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x23, 0x0a, 0x0d,
	0x66, 0x6f, 0x72, 0x63, 0x65, 0x5f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0c, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x18,
	0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65,
	0x73, 0x22, 0xaf, 0x02, 0x0a, 0x08, 0x45, 0x76, 0x61, 0x6c, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x6c, 0x68, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6c, 0x68, 0x73,
	0x12, 0x10, 0x0a, 0x03, 0x72, 0x68, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x72,
	0x68, 0x73, 0x12, 0x52, 0x0a, 0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x34, 0x2e, 0x73, 0x61, 0x70, 0x61, 0x67, 0x65, 0x6e, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x68, 0x61, 0x6e, 0x61, 0x69, 0x6e, 0x73, 0x69,
	0x67, 0x68, 0x74, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x4e, 0x6f,
	0x64, 0x65, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09, 0x6f, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4c, 0x0a, 0x0b, 0x63, 0x68, 0x69, 0x6c, 0x64, 0x5f,
	0x65, 0x76, 0x61, 0x6c, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x73, 0x61,
	0x70, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x68, 0x61,
	0x6e, 0x61, 0x69, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73, 0x2e, 0x72, 0x75, 0x6c, 0x65, 0x2e,
	0x45, 0x76, 0x61, 0x6c, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x0a, 0x63, 0x68, 0x69, 0x6c, 0x64, 0x45,
	0x76, 0x61, 0x6c, 0x73, 0x22, 0x5d, 0x0a, 0x08, 0x45, 0x76, 0x61, 0x6c, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x06, 0x0a, 0x02, 0x4f, 0x52, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x4e, 0x44, 0x10, 0x02,
	0x12, 0x06, 0x0a, 0x02, 0x45, 0x51, 0x10, 0x03, 0x12, 0x07, 0x0a, 0x03, 0x4e, 0x45, 0x51, 0x10,
	0x04, 0x12, 0x06, 0x0a, 0x02, 0x4c, 0x54, 0x10, 0x05, 0x12, 0x07, 0x0a, 0x03, 0x4c, 0x54, 0x45,
	0x10, 0x06, 0x12, 0x06, 0x0a, 0x02, 0x47, 0x54, 0x10, 0x07, 0x12, 0x07, 0x0a, 0x03, 0x47, 0x54,
	0x45, 0x10, 0x08, 0x22, 0x78, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x74, 0x61, 0x74, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x74, 0x61, 0x74, 0x65, 0x6d, 0x65, 0x6e,
	0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x42, 0x02, 0x50,
	0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_hanainsights_rule_rule_proto_rawDescOnce sync.Once
	file_hanainsights_rule_rule_proto_rawDescData = file_hanainsights_rule_rule_proto_rawDesc
)

func file_hanainsights_rule_rule_proto_rawDescGZIP() []byte {
	file_hanainsights_rule_rule_proto_rawDescOnce.Do(func() {
		file_hanainsights_rule_rule_proto_rawDescData = protoimpl.X.CompressGZIP(file_hanainsights_rule_rule_proto_rawDescData)
	})
	return file_hanainsights_rule_rule_proto_rawDescData
}

var file_hanainsights_rule_rule_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_hanainsights_rule_rule_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_hanainsights_rule_rule_proto_goTypes = []interface{}{
	(EvalNode_EvalType)(0), // 0: sapagent.protos.hanainsights.rule.EvalNode.EvalType
	(*Rule)(nil),           // 1: sapagent.protos.hanainsights.rule.Rule
	(*Query)(nil),          // 2: sapagent.protos.hanainsights.rule.Query
	(*Recommendation)(nil), // 3: sapagent.protos.hanainsights.rule.Recommendation
	(*EvalNode)(nil),       // 4: sapagent.protos.hanainsights.rule.EvalNode
	(*Action)(nil),         // 5: sapagent.protos.hanainsights.rule.Action
}
var file_hanainsights_rule_rule_proto_depIdxs = []int32{
	2, // 0: sapagent.protos.hanainsights.rule.Rule.queries:type_name -> sapagent.protos.hanainsights.rule.Query
	3, // 1: sapagent.protos.hanainsights.rule.Rule.recommendations:type_name -> sapagent.protos.hanainsights.rule.Recommendation
	4, // 2: sapagent.protos.hanainsights.rule.Recommendation.trigger:type_name -> sapagent.protos.hanainsights.rule.EvalNode
	5, // 3: sapagent.protos.hanainsights.rule.Recommendation.actions:type_name -> sapagent.protos.hanainsights.rule.Action
	0, // 4: sapagent.protos.hanainsights.rule.EvalNode.operation:type_name -> sapagent.protos.hanainsights.rule.EvalNode.EvalType
	4, // 5: sapagent.protos.hanainsights.rule.EvalNode.child_evals:type_name -> sapagent.protos.hanainsights.rule.EvalNode
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_hanainsights_rule_rule_proto_init() }
func file_hanainsights_rule_rule_proto_init() {
	if File_hanainsights_rule_rule_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_hanainsights_rule_rule_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Rule); i {
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
		file_hanainsights_rule_rule_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Query); i {
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
		file_hanainsights_rule_rule_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Recommendation); i {
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
		file_hanainsights_rule_rule_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EvalNode); i {
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
		file_hanainsights_rule_rule_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Action); i {
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
			RawDescriptor: file_hanainsights_rule_rule_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_hanainsights_rule_rule_proto_goTypes,
		DependencyIndexes: file_hanainsights_rule_rule_proto_depIdxs,
		EnumInfos:         file_hanainsights_rule_rule_proto_enumTypes,
		MessageInfos:      file_hanainsights_rule_rule_proto_msgTypes,
	}.Build()
	File_hanainsights_rule_rule_proto = out.File
	file_hanainsights_rule_rule_proto_rawDesc = nil
	file_hanainsights_rule_rule_proto_goTypes = nil
	file_hanainsights_rule_rule_proto_depIdxs = nil
}
