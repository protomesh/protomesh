// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: api/types/v1/process.proto

package typesv1

import (
	_ "github.com/protomesh/protoc-gen-terraform/proto/terraform"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Trigger_IDBuilder int32

const (
	// Don't add suffix
	Trigger_ID_BUILDER_ONLY_PREFIX Trigger_IDBuilder = 0
	// Generate a new random id each time
	Trigger_ID_BUILDER_RANDOM Trigger_IDBuilder = 1
	// Generate a unique ID for this workflow
	Trigger_ID_BUILDER_UNIQUE Trigger_IDBuilder = 2
)

// Enum value maps for Trigger_IDBuilder.
var (
	Trigger_IDBuilder_name = map[int32]string{
		0: "ID_BUILDER_ONLY_PREFIX",
		1: "ID_BUILDER_RANDOM",
		2: "ID_BUILDER_UNIQUE",
	}
	Trigger_IDBuilder_value = map[string]int32{
		"ID_BUILDER_ONLY_PREFIX": 0,
		"ID_BUILDER_RANDOM":      1,
		"ID_BUILDER_UNIQUE":      2,
	}
)

func (x Trigger_IDBuilder) Enum() *Trigger_IDBuilder {
	p := new(Trigger_IDBuilder)
	*p = x
	return p
}

func (x Trigger_IDBuilder) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Trigger_IDBuilder) Descriptor() protoreflect.EnumDescriptor {
	return file_api_types_v1_process_proto_enumTypes[0].Descriptor()
}

func (Trigger_IDBuilder) Type() protoreflect.EnumType {
	return &file_api_types_v1_process_proto_enumTypes[0]
}

func (x Trigger_IDBuilder) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Trigger_IDBuilder.Descriptor instead.
func (Trigger_IDBuilder) EnumDescriptor() ([]byte, []int) {
	return file_api_types_v1_process_proto_rawDescGZIP(), []int{0, 0}
}

type Trigger_IfRunningAction int32

const (
	// Abort the current event to keep the running
	Trigger_IF_RUNNING_ABORT Trigger_IfRunningAction = 0
	// Cancel the running workflow and start the current event
	Trigger_IF_RUNNING_OVERLAP Trigger_IfRunningAction = 1
)

// Enum value maps for Trigger_IfRunningAction.
var (
	Trigger_IfRunningAction_name = map[int32]string{
		0: "IF_RUNNING_ABORT",
		1: "IF_RUNNING_OVERLAP",
	}
	Trigger_IfRunningAction_value = map[string]int32{
		"IF_RUNNING_ABORT":   0,
		"IF_RUNNING_OVERLAP": 1,
	}
)

func (x Trigger_IfRunningAction) Enum() *Trigger_IfRunningAction {
	p := new(Trigger_IfRunningAction)
	*p = x
	return p
}

func (x Trigger_IfRunningAction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Trigger_IfRunningAction) Descriptor() protoreflect.EnumDescriptor {
	return file_api_types_v1_process_proto_enumTypes[1].Descriptor()
}

func (Trigger_IfRunningAction) Type() protoreflect.EnumType {
	return &file_api_types_v1_process_proto_enumTypes[1]
}

func (x Trigger_IfRunningAction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Trigger_IfRunningAction.Descriptor instead.
func (Trigger_IfRunningAction) EnumDescriptor() ([]byte, []int) {
	return file_api_types_v1_process_proto_rawDescGZIP(), []int{0, 1}
}

type Trigger_OnDropAction int32

const (
	Trigger_ON_DROP_DO_NOTHING Trigger_OnDropAction = 0
	Trigger_ON_DROP_CANCEL     Trigger_OnDropAction = 1
	Trigger_ON_DROP_TERMINATE  Trigger_OnDropAction = 2
)

// Enum value maps for Trigger_OnDropAction.
var (
	Trigger_OnDropAction_name = map[int32]string{
		0: "ON_DROP_DO_NOTHING",
		1: "ON_DROP_CANCEL",
		2: "ON_DROP_TERMINATE",
	}
	Trigger_OnDropAction_value = map[string]int32{
		"ON_DROP_DO_NOTHING": 0,
		"ON_DROP_CANCEL":     1,
		"ON_DROP_TERMINATE":  2,
	}
)

func (x Trigger_OnDropAction) Enum() *Trigger_OnDropAction {
	p := new(Trigger_OnDropAction)
	*p = x
	return p
}

func (x Trigger_OnDropAction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Trigger_OnDropAction) Descriptor() protoreflect.EnumDescriptor {
	return file_api_types_v1_process_proto_enumTypes[2].Descriptor()
}

func (Trigger_OnDropAction) Type() protoreflect.EnumType {
	return &file_api_types_v1_process_proto_enumTypes[2]
}

func (x Trigger_OnDropAction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Trigger_OnDropAction.Descriptor instead.
func (Trigger_OnDropAction) EnumDescriptor() ([]byte, []int) {
	return file_api_types_v1_process_proto_rawDescGZIP(), []int{0, 2}
}

type Trigger struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the workflow to trigger
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Task queue on temporal to send workflow tasks
	TaskQueue string `protobuf:"bytes,2,opt,name=task_queue,json=taskQueue,proto3" json:"task_queue,omitempty"`
	IdPrefix  string `protobuf:"bytes,3,opt,name=id_prefix,json=idPrefix,proto3" json:"id_prefix,omitempty"`
	// Types that are assignable to IdSuffix:
	//
	//	*Trigger_ExactIdSuffix
	//	*Trigger_IdSuffixBuilder
	IdSuffix isTrigger_IdSuffix `protobuf_oneof:"id_suffix"`
	// Types that are assignable to IfRunning:
	//
	//	*Trigger_IfRunningAction_
	IfRunning        isTrigger_IfRunning  `protobuf_oneof:"if_running"`
	CronSchedule     string               `protobuf:"bytes,7,opt,name=cron_schedule,json=cronSchedule,proto3" json:"cron_schedule,omitempty"`
	ExecutionTimeout *durationpb.Duration `protobuf:"bytes,8,opt,name=execution_timeout,json=executionTimeout,proto3" json:"execution_timeout,omitempty"`
	RunTimeout       *durationpb.Duration `protobuf:"bytes,9,opt,name=run_timeout,json=runTimeout,proto3" json:"run_timeout,omitempty"`
	TaskTimeout      *durationpb.Duration `protobuf:"bytes,10,opt,name=task_timeout,json=taskTimeout,proto3" json:"task_timeout,omitempty"`
	Arguments        *structpb.Value      `protobuf:"bytes,11,opt,name=arguments,proto3" json:"arguments,omitempty"`
	RetryPolicy      *Trigger_RetryPolicy `protobuf:"bytes,12,opt,name=retry_policy,json=retryPolicy,proto3" json:"retry_policy,omitempty"`
	// Types that are assignable to OnDrop:
	//
	//	*Trigger_OnDropAction_
	OnDrop isTrigger_OnDrop `protobuf_oneof:"on_drop"`
}

func (x *Trigger) Reset() {
	*x = Trigger{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_types_v1_process_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Trigger) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Trigger) ProtoMessage() {}

func (x *Trigger) ProtoReflect() protoreflect.Message {
	mi := &file_api_types_v1_process_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Trigger.ProtoReflect.Descriptor instead.
func (*Trigger) Descriptor() ([]byte, []int) {
	return file_api_types_v1_process_proto_rawDescGZIP(), []int{0}
}

func (x *Trigger) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Trigger) GetTaskQueue() string {
	if x != nil {
		return x.TaskQueue
	}
	return ""
}

func (x *Trigger) GetIdPrefix() string {
	if x != nil {
		return x.IdPrefix
	}
	return ""
}

func (m *Trigger) GetIdSuffix() isTrigger_IdSuffix {
	if m != nil {
		return m.IdSuffix
	}
	return nil
}

func (x *Trigger) GetExactIdSuffix() string {
	if x, ok := x.GetIdSuffix().(*Trigger_ExactIdSuffix); ok {
		return x.ExactIdSuffix
	}
	return ""
}

func (x *Trigger) GetIdSuffixBuilder() Trigger_IDBuilder {
	if x, ok := x.GetIdSuffix().(*Trigger_IdSuffixBuilder); ok {
		return x.IdSuffixBuilder
	}
	return Trigger_ID_BUILDER_ONLY_PREFIX
}

func (m *Trigger) GetIfRunning() isTrigger_IfRunning {
	if m != nil {
		return m.IfRunning
	}
	return nil
}

func (x *Trigger) GetIfRunningAction() Trigger_IfRunningAction {
	if x, ok := x.GetIfRunning().(*Trigger_IfRunningAction_); ok {
		return x.IfRunningAction
	}
	return Trigger_IF_RUNNING_ABORT
}

func (x *Trigger) GetCronSchedule() string {
	if x != nil {
		return x.CronSchedule
	}
	return ""
}

func (x *Trigger) GetExecutionTimeout() *durationpb.Duration {
	if x != nil {
		return x.ExecutionTimeout
	}
	return nil
}

func (x *Trigger) GetRunTimeout() *durationpb.Duration {
	if x != nil {
		return x.RunTimeout
	}
	return nil
}

func (x *Trigger) GetTaskTimeout() *durationpb.Duration {
	if x != nil {
		return x.TaskTimeout
	}
	return nil
}

func (x *Trigger) GetArguments() *structpb.Value {
	if x != nil {
		return x.Arguments
	}
	return nil
}

func (x *Trigger) GetRetryPolicy() *Trigger_RetryPolicy {
	if x != nil {
		return x.RetryPolicy
	}
	return nil
}

func (m *Trigger) GetOnDrop() isTrigger_OnDrop {
	if m != nil {
		return m.OnDrop
	}
	return nil
}

func (x *Trigger) GetOnDropAction() Trigger_OnDropAction {
	if x, ok := x.GetOnDrop().(*Trigger_OnDropAction_); ok {
		return x.OnDropAction
	}
	return Trigger_ON_DROP_DO_NOTHING
}

type isTrigger_IdSuffix interface {
	isTrigger_IdSuffix()
}

type Trigger_ExactIdSuffix struct {
	// Use this exact id for the workflow id
	ExactIdSuffix string `protobuf:"bytes,4,opt,name=exact_id_suffix,json=exactIdSuffix,proto3,oneof"`
}

type Trigger_IdSuffixBuilder struct {
	IdSuffixBuilder Trigger_IDBuilder `protobuf:"varint,5,opt,name=id_suffix_builder,json=idSuffixBuilder,proto3,enum=protomesh.types.v1.Trigger_IDBuilder,oneof"`
}

func (*Trigger_ExactIdSuffix) isTrigger_IdSuffix() {}

func (*Trigger_IdSuffixBuilder) isTrigger_IdSuffix() {}

type isTrigger_IfRunning interface {
	isTrigger_IfRunning()
}

type Trigger_IfRunningAction_ struct {
	IfRunningAction Trigger_IfRunningAction `protobuf:"varint,6,opt,name=if_running_action,json=ifRunningAction,proto3,enum=protomesh.types.v1.Trigger_IfRunningAction,oneof"`
}

func (*Trigger_IfRunningAction_) isTrigger_IfRunning() {}

type isTrigger_OnDrop interface {
	isTrigger_OnDrop()
}

type Trigger_OnDropAction_ struct {
	OnDropAction Trigger_OnDropAction `protobuf:"varint,13,opt,name=on_drop_action,json=onDropAction,proto3,enum=protomesh.types.v1.Trigger_OnDropAction,oneof"`
}

func (*Trigger_OnDropAction_) isTrigger_OnDrop() {}

type Trigger_RetryPolicy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InitialInterval    *durationpb.Duration `protobuf:"bytes,1,opt,name=initial_interval,json=initialInterval,proto3" json:"initial_interval,omitempty"`
	MaximumBackoff     *durationpb.Duration `protobuf:"bytes,2,opt,name=maximum_backoff,json=maximumBackoff,proto3" json:"maximum_backoff,omitempty"`
	MaximumAttempts    int32                `protobuf:"varint,3,opt,name=maximum_attempts,json=maximumAttempts,proto3" json:"maximum_attempts,omitempty"`
	NonRetryableErrors []string             `protobuf:"bytes,4,rep,name=non_retryable_errors,json=nonRetryableErrors,proto3" json:"non_retryable_errors,omitempty"`
}

func (x *Trigger_RetryPolicy) Reset() {
	*x = Trigger_RetryPolicy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_types_v1_process_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Trigger_RetryPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Trigger_RetryPolicy) ProtoMessage() {}

func (x *Trigger_RetryPolicy) ProtoReflect() protoreflect.Message {
	mi := &file_api_types_v1_process_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Trigger_RetryPolicy.ProtoReflect.Descriptor instead.
func (*Trigger_RetryPolicy) Descriptor() ([]byte, []int) {
	return file_api_types_v1_process_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Trigger_RetryPolicy) GetInitialInterval() *durationpb.Duration {
	if x != nil {
		return x.InitialInterval
	}
	return nil
}

func (x *Trigger_RetryPolicy) GetMaximumBackoff() *durationpb.Duration {
	if x != nil {
		return x.MaximumBackoff
	}
	return nil
}

func (x *Trigger_RetryPolicy) GetMaximumAttempts() int32 {
	if x != nil {
		return x.MaximumAttempts
	}
	return 0
}

func (x *Trigger_RetryPolicy) GetNonRetryableErrors() []string {
	if x != nil {
		return x.NonRetryableErrors
	}
	return nil
}

var File_api_types_v1_process_proto protoreflect.FileDescriptor

var file_api_types_v1_process_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x70,
	0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b,
	0x74, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x93, 0x0a, 0x0a, 0x07,
	0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0xb9, 0x02, 0x02, 0x10, 0x01, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x71, 0x75, 0x65, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x51, 0x75, 0x65,
	0x75, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x64, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x64, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12,
	0x28, 0x0a, 0x0f, 0x65, 0x78, 0x61, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x5f, 0x73, 0x75, 0x66, 0x66,
	0x69, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0d, 0x65, 0x78, 0x61, 0x63,
	0x74, 0x49, 0x64, 0x53, 0x75, 0x66, 0x66, 0x69, 0x78, 0x12, 0x53, 0x0a, 0x11, 0x69, 0x64, 0x5f,
	0x73, 0x75, 0x66, 0x66, 0x69, 0x78, 0x5f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x25, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68,
	0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x2e, 0x49, 0x44, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x48, 0x00, 0x52, 0x0f, 0x69,
	0x64, 0x53, 0x75, 0x66, 0x66, 0x69, 0x78, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x12, 0x59,
	0x0a, 0x11, 0x69, 0x66, 0x5f, 0x72, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x54,
	0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x49, 0x66, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67,
	0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x01, 0x52, 0x0f, 0x69, 0x66, 0x52, 0x75, 0x6e, 0x6e,
	0x69, 0x6e, 0x67, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x72, 0x6f,
	0x6e, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x63, 0x72, 0x6f, 0x6e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x46,
	0x0a, 0x11, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x6f, 0x75, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x10, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x54,
	0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x3a, 0x0a, 0x0b, 0x72, 0x75, 0x6e, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x72, 0x75, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x12, 0x3c, 0x0a, 0x0c, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x74, 0x61, 0x73, 0x6b, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74,
	0x12, 0x34, 0x0a, 0x09, 0x61, 0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x09, 0x61, 0x72, 0x67,
	0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x4a, 0x0a, 0x0c, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f,
	0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x50,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x0b, 0x72, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x12, 0x50, 0x0a, 0x0e, 0x6f, 0x6e, 0x5f, 0x64, 0x72, 0x6f, 0x70, 0x5f, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x28, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x2e,
	0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x4f, 0x6e, 0x44, 0x72, 0x6f, 0x70, 0x41, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x48, 0x02, 0x52, 0x0c, 0x6f, 0x6e, 0x44, 0x72, 0x6f, 0x70, 0x41, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x83, 0x02, 0x0a, 0x0b, 0x52, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x12, 0x53, 0x0a, 0x10, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x5f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x0d, 0xba, 0xb9, 0x02, 0x09, 0x10,
	0x01, 0x1a, 0x05, 0x1a, 0x03, 0x33, 0x30, 0x73, 0x52, 0x0f, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61,
	0x6c, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x42, 0x0a, 0x0f, 0x6d, 0x61, 0x78,
	0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0e, 0x6d,
	0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x42, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x12, 0x29, 0x0a,
	0x10, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x61, 0x74, 0x74, 0x65, 0x6d, 0x70, 0x74,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d,
	0x41, 0x74, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x73, 0x12, 0x30, 0x0a, 0x14, 0x6e, 0x6f, 0x6e, 0x5f,
	0x72, 0x65, 0x74, 0x72, 0x79, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x12, 0x6e, 0x6f, 0x6e, 0x52, 0x65, 0x74, 0x72, 0x79,
	0x61, 0x62, 0x6c, 0x65, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x22, 0x55, 0x0a, 0x09, 0x49, 0x44,
	0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x16, 0x49, 0x44, 0x5f, 0x42, 0x55,
	0x49, 0x4c, 0x44, 0x45, 0x52, 0x5f, 0x4f, 0x4e, 0x4c, 0x59, 0x5f, 0x50, 0x52, 0x45, 0x46, 0x49,
	0x58, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x49, 0x44, 0x5f, 0x42, 0x55, 0x49, 0x4c, 0x44, 0x45,
	0x52, 0x5f, 0x52, 0x41, 0x4e, 0x44, 0x4f, 0x4d, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11, 0x49, 0x44,
	0x5f, 0x42, 0x55, 0x49, 0x4c, 0x44, 0x45, 0x52, 0x5f, 0x55, 0x4e, 0x49, 0x51, 0x55, 0x45, 0x10,
	0x02, 0x22, 0x3f, 0x0a, 0x0f, 0x49, 0x66, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x41, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x10, 0x49, 0x46, 0x5f, 0x52, 0x55, 0x4e, 0x4e, 0x49,
	0x4e, 0x47, 0x5f, 0x41, 0x42, 0x4f, 0x52, 0x54, 0x10, 0x00, 0x12, 0x16, 0x0a, 0x12, 0x49, 0x46,
	0x5f, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x5f, 0x4f, 0x56, 0x45, 0x52, 0x4c, 0x41, 0x50,
	0x10, 0x01, 0x22, 0x51, 0x0a, 0x0c, 0x4f, 0x6e, 0x44, 0x72, 0x6f, 0x70, 0x41, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x12, 0x4f, 0x4e, 0x5f, 0x44, 0x52, 0x4f, 0x50, 0x5f, 0x44, 0x4f,
	0x5f, 0x4e, 0x4f, 0x54, 0x48, 0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x12, 0x0a, 0x0e, 0x4f, 0x4e,
	0x5f, 0x44, 0x52, 0x4f, 0x50, 0x5f, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x10, 0x01, 0x12, 0x15,
	0x0a, 0x11, 0x4f, 0x4e, 0x5f, 0x44, 0x52, 0x4f, 0x50, 0x5f, 0x54, 0x45, 0x52, 0x4d, 0x49, 0x4e,
	0x41, 0x54, 0x45, 0x10, 0x02, 0x3a, 0x04, 0xba, 0xb9, 0x02, 0x00, 0x42, 0x0b, 0x0a, 0x09, 0x69,
	0x64, 0x5f, 0x73, 0x75, 0x66, 0x66, 0x69, 0x78, 0x42, 0x0c, 0x0a, 0x0a, 0x69, 0x66, 0x5f, 0x72,
	0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x42, 0x09, 0x0a, 0x07, 0x6f, 0x6e, 0x5f, 0x64, 0x72, 0x6f,
	0x70, 0x42, 0xa3, 0x02, 0x0a, 0x16, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x6d,
	0x65, 0x73, 0x68, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x76, 0x31, 0x42, 0x0c, 0x50, 0x72,
	0x6f, 0x63, 0x65, 0x73, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x39, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65,
	0x73, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x3b,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x50, 0x54, 0x58, 0xaa, 0x02, 0x12,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x56, 0x31, 0xca, 0x02, 0x12, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68, 0x5c, 0x54,
	0x79, 0x70, 0x65, 0x73, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x1e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x6d,
	0x65, 0x73, 0x68, 0x5c, 0x54, 0x79, 0x70, 0x65, 0x73, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x14, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x6d, 0x65, 0x73, 0x68, 0x3a, 0x3a, 0x54, 0x79, 0x70, 0x65, 0x73, 0x3a, 0x3a, 0x56, 0x31, 0xba,
	0xb9, 0x02, 0x54, 0x0a, 0x52, 0x0a, 0x15, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x39, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65,
	0x73, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x3b, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x6d, 0x65, 0x73, 0x68, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_types_v1_process_proto_rawDescOnce sync.Once
	file_api_types_v1_process_proto_rawDescData = file_api_types_v1_process_proto_rawDesc
)

func file_api_types_v1_process_proto_rawDescGZIP() []byte {
	file_api_types_v1_process_proto_rawDescOnce.Do(func() {
		file_api_types_v1_process_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_types_v1_process_proto_rawDescData)
	})
	return file_api_types_v1_process_proto_rawDescData
}

var file_api_types_v1_process_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_api_types_v1_process_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_types_v1_process_proto_goTypes = []interface{}{
	(Trigger_IDBuilder)(0),       // 0: protomesh.types.v1.Trigger.IDBuilder
	(Trigger_IfRunningAction)(0), // 1: protomesh.types.v1.Trigger.IfRunningAction
	(Trigger_OnDropAction)(0),    // 2: protomesh.types.v1.Trigger.OnDropAction
	(*Trigger)(nil),              // 3: protomesh.types.v1.Trigger
	(*Trigger_RetryPolicy)(nil),  // 4: protomesh.types.v1.Trigger.RetryPolicy
	(*durationpb.Duration)(nil),  // 5: google.protobuf.Duration
	(*structpb.Value)(nil),       // 6: google.protobuf.Value
}
var file_api_types_v1_process_proto_depIdxs = []int32{
	0,  // 0: protomesh.types.v1.Trigger.id_suffix_builder:type_name -> protomesh.types.v1.Trigger.IDBuilder
	1,  // 1: protomesh.types.v1.Trigger.if_running_action:type_name -> protomesh.types.v1.Trigger.IfRunningAction
	5,  // 2: protomesh.types.v1.Trigger.execution_timeout:type_name -> google.protobuf.Duration
	5,  // 3: protomesh.types.v1.Trigger.run_timeout:type_name -> google.protobuf.Duration
	5,  // 4: protomesh.types.v1.Trigger.task_timeout:type_name -> google.protobuf.Duration
	6,  // 5: protomesh.types.v1.Trigger.arguments:type_name -> google.protobuf.Value
	4,  // 6: protomesh.types.v1.Trigger.retry_policy:type_name -> protomesh.types.v1.Trigger.RetryPolicy
	2,  // 7: protomesh.types.v1.Trigger.on_drop_action:type_name -> protomesh.types.v1.Trigger.OnDropAction
	5,  // 8: protomesh.types.v1.Trigger.RetryPolicy.initial_interval:type_name -> google.protobuf.Duration
	5,  // 9: protomesh.types.v1.Trigger.RetryPolicy.maximum_backoff:type_name -> google.protobuf.Duration
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_api_types_v1_process_proto_init() }
func file_api_types_v1_process_proto_init() {
	if File_api_types_v1_process_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_types_v1_process_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Trigger); i {
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
		file_api_types_v1_process_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Trigger_RetryPolicy); i {
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
	file_api_types_v1_process_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Trigger_ExactIdSuffix)(nil),
		(*Trigger_IdSuffixBuilder)(nil),
		(*Trigger_IfRunningAction_)(nil),
		(*Trigger_OnDropAction_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_types_v1_process_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_types_v1_process_proto_goTypes,
		DependencyIndexes: file_api_types_v1_process_proto_depIdxs,
		EnumInfos:         file_api_types_v1_process_proto_enumTypes,
		MessageInfos:      file_api_types_v1_process_proto_msgTypes,
	}.Build()
	File_api_types_v1_process_proto = out.File
	file_api_types_v1_process_proto_rawDesc = nil
	file_api_types_v1_process_proto_goTypes = nil
	file_api_types_v1_process_proto_depIdxs = nil
}
