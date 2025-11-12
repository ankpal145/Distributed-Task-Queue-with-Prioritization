package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Priority int32

const (
	Priority_LOW      Priority = 0
	Priority_NORMAL   Priority = 1
	Priority_HIGH     Priority = 2
	Priority_CRITICAL Priority = 3
)

var (
	Priority_name = map[int32]string{
		0: "LOW",
		1: "NORMAL",
		2: "HIGH",
		3: "CRITICAL",
	}
	Priority_value = map[string]int32{
		"LOW":      0,
		"NORMAL":   1,
		"HIGH":     2,
		"CRITICAL": 3,
	}
)

func (x Priority) Enum() *Priority {
	p := new(Priority)
	*p = x
	return p
}

func (x Priority) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Priority) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_task_queue_proto_enumTypes[0].Descriptor()
}

func (Priority) Type() protoreflect.EnumType {
	return &file_proto_task_queue_proto_enumTypes[0]
}

func (x Priority) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

func (Priority) EnumDescriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{0}
}

type TaskStatus int32

const (
	TaskStatus_PENDING    TaskStatus = 0
	TaskStatus_PROCESSING TaskStatus = 1
	TaskStatus_COMPLETED  TaskStatus = 2
	TaskStatus_FAILED     TaskStatus = 3
	TaskStatus_DEAD       TaskStatus = 4
	TaskStatus_CANCELLED  TaskStatus = 5
)

var (
	TaskStatus_name = map[int32]string{
		0: "PENDING",
		1: "PROCESSING",
		2: "COMPLETED",
		3: "FAILED",
		4: "DEAD",
		5: "CANCELLED",
	}
	TaskStatus_value = map[string]int32{
		"PENDING":    0,
		"PROCESSING": 1,
		"COMPLETED":  2,
		"FAILED":     3,
		"DEAD":       4,
		"CANCELLED":  5,
	}
)

func (x TaskStatus) Enum() *TaskStatus {
	p := new(TaskStatus)
	*p = x
	return p
}

func (x TaskStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_task_queue_proto_enumTypes[1].Descriptor()
}

func (TaskStatus) Type() protoreflect.EnumType {
	return &file_proto_task_queue_proto_enumTypes[1]
}

func (x TaskStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

func (TaskStatus) EnumDescriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{1}
}

type SubmitTaskRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          string                 `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Priority      Priority               `protobuf:"varint,2,opt,name=priority,proto3,enum=taskqueue.Priority" json:"priority,omitempty"`
	Payload       []byte                 `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
	MaxRetries    int32                  `protobuf:"varint,4,opt,name=max_retries,json=maxRetries,proto3" json:"max_retries,omitempty"`
	TimeoutMs     int64                  `protobuf:"varint,5,opt,name=timeout_ms,json=timeoutMs,proto3" json:"timeout_ms,omitempty"`
	ScheduledAt   int64                  `protobuf:"varint,6,opt,name=scheduled_at,json=scheduledAt,proto3" json:"scheduled_at,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,7,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitTaskRequest) Reset() {
	*x = SubmitTaskRequest{}
	mi := &file_proto_task_queue_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitTaskRequest) ProtoMessage() {}

func (x *SubmitTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SubmitTaskRequest) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{0}
}

func (x *SubmitTaskRequest) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *SubmitTaskRequest) GetPriority() Priority {
	if x != nil {
		return x.Priority
	}
	return Priority_LOW
}

func (x *SubmitTaskRequest) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *SubmitTaskRequest) GetMaxRetries() int32 {
	if x != nil {
		return x.MaxRetries
	}
	return 0
}

func (x *SubmitTaskRequest) GetTimeoutMs() int64 {
	if x != nil {
		return x.TimeoutMs
	}
	return 0
}

func (x *SubmitTaskRequest) GetScheduledAt() int64 {
	if x != nil {
		return x.ScheduledAt
	}
	return 0
}

func (x *SubmitTaskRequest) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type TaskResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Priority      Priority               `protobuf:"varint,3,opt,name=priority,proto3,enum=taskqueue.Priority" json:"priority,omitempty"`
	Payload       []byte                 `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`
	Status        TaskStatus             `protobuf:"varint,5,opt,name=status,proto3,enum=taskqueue.TaskStatus" json:"status,omitempty"`
	Result        []byte                 `protobuf:"bytes,6,opt,name=result,proto3" json:"result,omitempty"`
	Error         string                 `protobuf:"bytes,7,opt,name=error,proto3" json:"error,omitempty"`
	RetryCount    int32                  `protobuf:"varint,8,opt,name=retry_count,json=retryCount,proto3" json:"retry_count,omitempty"`
	MaxRetries    int32                  `protobuf:"varint,9,opt,name=max_retries,json=maxRetries,proto3" json:"max_retries,omitempty"`
	CreatedAt     int64                  `protobuf:"varint,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	ScheduledAt   int64                  `protobuf:"varint,11,opt,name=scheduled_at,json=scheduledAt,proto3" json:"scheduled_at,omitempty"`
	StartedAt     int64                  `protobuf:"varint,12,opt,name=started_at,json=startedAt,proto3" json:"started_at,omitempty"`
	CompletedAt   int64                  `protobuf:"varint,13,opt,name=completed_at,json=completedAt,proto3" json:"completed_at,omitempty"`
	TimeoutMs     int64                  `protobuf:"varint,14,opt,name=timeout_ms,json=timeoutMs,proto3" json:"timeout_ms,omitempty"`
	WorkerId      string                 `protobuf:"bytes,15,opt,name=worker_id,json=workerId,proto3" json:"worker_id,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,16,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskResponse) Reset() {
	*x = TaskResponse{}
	mi := &file_proto_task_queue_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResponse) ProtoMessage() {}

func (x *TaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TaskResponse) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{1}
}

func (x *TaskResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TaskResponse) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *TaskResponse) GetPriority() Priority {
	if x != nil {
		return x.Priority
	}
	return Priority_LOW
}

func (x *TaskResponse) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *TaskResponse) GetStatus() TaskStatus {
	if x != nil {
		return x.Status
	}
	return TaskStatus_PENDING
}

func (x *TaskResponse) GetResult() []byte {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *TaskResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *TaskResponse) GetRetryCount() int32 {
	if x != nil {
		return x.RetryCount
	}
	return 0
}

func (x *TaskResponse) GetMaxRetries() int32 {
	if x != nil {
		return x.MaxRetries
	}
	return 0
}

func (x *TaskResponse) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *TaskResponse) GetScheduledAt() int64 {
	if x != nil {
		return x.ScheduledAt
	}
	return 0
}

func (x *TaskResponse) GetStartedAt() int64 {
	if x != nil {
		return x.StartedAt
	}
	return 0
}

func (x *TaskResponse) GetCompletedAt() int64 {
	if x != nil {
		return x.CompletedAt
	}
	return 0
}

func (x *TaskResponse) GetTimeoutMs() int64 {
	if x != nil {
		return x.TimeoutMs
	}
	return 0
}

func (x *TaskResponse) GetWorkerId() string {
	if x != nil {
		return x.WorkerId
	}
	return ""
}

func (x *TaskResponse) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type GetTaskStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTaskStatusRequest) Reset() {
	*x = GetTaskStatusRequest{}
	mi := &file_proto_task_queue_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTaskStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTaskStatusRequest) ProtoMessage() {}

func (x *GetTaskStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetTaskStatusRequest) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{2}
}

func (x *GetTaskStatusRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type CancelTaskRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CancelTaskRequest) Reset() {
	*x = CancelTaskRequest{}
	mi := &file_proto_task_queue_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelTaskRequest) ProtoMessage() {}

func (x *CancelTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CancelTaskRequest) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{3}
}

func (x *CancelTaskRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type CancelTaskResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CancelTaskResponse) Reset() {
	*x = CancelTaskResponse{}
	mi := &file_proto_task_queue_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelTaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelTaskResponse) ProtoMessage() {}

func (x *CancelTaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CancelTaskResponse) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{4}
}

func (x *CancelTaskResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *CancelTaskResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type StreamTasksRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskIds       []string               `protobuf:"bytes,1,rep,name=task_ids,json=taskIds,proto3" json:"task_ids,omitempty"`
	Statuses      []TaskStatus           `protobuf:"varint,2,rep,packed,name=statuses,proto3,enum=taskqueue.TaskStatus" json:"statuses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StreamTasksRequest) Reset() {
	*x = StreamTasksRequest{}
	mi := &file_proto_task_queue_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamTasksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamTasksRequest) ProtoMessage() {}

func (x *StreamTasksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*StreamTasksRequest) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{5}
}

func (x *StreamTasksRequest) GetTaskIds() []string {
	if x != nil {
		return x.TaskIds
	}
	return nil
}

func (x *StreamTasksRequest) GetStatuses() []TaskStatus {
	if x != nil {
		return x.Statuses
	}
	return nil
}

type TaskUpdate struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TaskId        string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	Status        TaskStatus             `protobuf:"varint,2,opt,name=status,proto3,enum=taskqueue.TaskStatus" json:"status,omitempty"`
	Timestamp     int64                  `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskUpdate) Reset() {
	*x = TaskUpdate{}
	mi := &file_proto_task_queue_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskUpdate) ProtoMessage() {}

func (x *TaskUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TaskUpdate) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{6}
}

func (x *TaskUpdate) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *TaskUpdate) GetStatus() TaskStatus {
	if x != nil {
		return x.Status
	}
	return TaskStatus_PENDING
}

func (x *TaskUpdate) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type RegisterWorkerRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Queues        []string               `protobuf:"bytes,2,rep,name=queues,proto3" json:"queues,omitempty"`
	Concurrency   int32                  `protobuf:"varint,3,opt,name=concurrency,proto3" json:"concurrency,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,4,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegisterWorkerRequest) Reset() {
	*x = RegisterWorkerRequest{}
	mi := &file_proto_task_queue_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterWorkerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterWorkerRequest) ProtoMessage() {}

func (x *RegisterWorkerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RegisterWorkerRequest) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{7}
}

func (x *RegisterWorkerRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *RegisterWorkerRequest) GetQueues() []string {
	if x != nil {
		return x.Queues
	}
	return nil
}

func (x *RegisterWorkerRequest) GetConcurrency() int32 {
	if x != nil {
		return x.Concurrency
	}
	return 0
}

func (x *RegisterWorkerRequest) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type WorkerResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Queues        []string               `protobuf:"bytes,3,rep,name=queues,proto3" json:"queues,omitempty"`
	Concurrency   int32                  `protobuf:"varint,4,opt,name=concurrency,proto3" json:"concurrency,omitempty"`
	RegisteredAt  int64                  `protobuf:"varint,5,opt,name=registered_at,json=registeredAt,proto3" json:"registered_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *WorkerResponse) Reset() {
	*x = WorkerResponse{}
	mi := &file_proto_task_queue_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WorkerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkerResponse) ProtoMessage() {}

func (x *WorkerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*WorkerResponse) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{8}
}

func (x *WorkerResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *WorkerResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *WorkerResponse) GetQueues() []string {
	if x != nil {
		return x.Queues
	}
	return nil
}

func (x *WorkerResponse) GetConcurrency() int32 {
	if x != nil {
		return x.Concurrency
	}
	return 0
}

func (x *WorkerResponse) GetRegisteredAt() int64 {
	if x != nil {
		return x.RegisteredAt
	}
	return 0
}

type HeartbeatRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	WorkerId      string                 `protobuf:"bytes,1,opt,name=worker_id,json=workerId,proto3" json:"worker_id,omitempty"`
	Timestamp     int64                  `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeartbeatRequest) Reset() {
	*x = HeartbeatRequest{}
	mi := &file_proto_task_queue_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatRequest) ProtoMessage() {}

func (x *HeartbeatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HeartbeatRequest) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{9}
}

func (x *HeartbeatRequest) GetWorkerId() string {
	if x != nil {
		return x.WorkerId
	}
	return ""
}

func (x *HeartbeatRequest) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type HeartbeatResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	NextHeartbeat int64                  `protobuf:"varint,2,opt,name=next_heartbeat,json=nextHeartbeat,proto3" json:"next_heartbeat,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeartbeatResponse) Reset() {
	*x = HeartbeatResponse{}
	mi := &file_proto_task_queue_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatResponse) ProtoMessage() {}

func (x *HeartbeatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_task_queue_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HeartbeatResponse) Descriptor() ([]byte, []int) {
	return file_proto_task_queue_proto_rawDescGZIP(), []int{10}
}

func (x *HeartbeatResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *HeartbeatResponse) GetNextHeartbeat() int64 {
	if x != nil {
		return x.NextHeartbeat
	}
	return 0
}

var File_proto_task_queue_proto protoreflect.FileDescriptor

const file_proto_task_queue_proto_rawDesc = "" +
	"\n" +
	"\x16proto/task_queue.proto\x12\ttaskqueue\"\xda\x02\n" +
	"\x11SubmitTaskRequest\x12\x12\n" +
	"\x04type\x18\x01 \x01(\tR\x04type\x12/\n" +
	"\bpriority\x18\x02 \x01(\x0e2\x13.taskqueue.PriorityR\bpriority\x12\x18\n" +
	"\apayload\x18\x03 \x01(\fR\apayload\x12\x1f\n" +
	"\vmax_retries\x18\x04 \x01(\x05R\n" +
	"maxRetries\x12\x1d\n" +
	"\n" +
	"timeout_ms\x18\x05 \x01(\x03R\ttimeoutMs\x12!\n" +
	"\fscheduled_at\x18\x06 \x01(\x03R\vscheduledAt\x12F\n" +
	"\bmetadata\x18\a \x03(\v2*.taskqueue.SubmitTaskRequest.MetadataEntryR\bmetadata\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\xdc\x04\n" +
	"\fTaskResponse\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04type\x18\x02 \x01(\tR\x04type\x12/\n" +
	"\bpriority\x18\x03 \x01(\x0e2\x13.taskqueue.PriorityR\bpriority\x12\x18\n" +
	"\apayload\x18\x04 \x01(\fR\apayload\x12-\n" +
	"\x06status\x18\x05 \x01(\x0e2\x15.taskqueue.TaskStatusR\x06status\x12\x16\n" +
	"\x06result\x18\x06 \x01(\fR\x06result\x12\x14\n" +
	"\x05error\x18\a \x01(\tR\x05error\x12\x1f\n" +
	"\vretry_count\x18\b \x01(\x05R\n" +
	"retryCount\x12\x1f\n" +
	"\vmax_retries\x18\t \x01(\x05R\n" +
	"maxRetries\x12\x1d\n" +
	"\n" +
	"created_at\x18\n" +
	" \x01(\x03R\tcreatedAt\x12!\n" +
	"\fscheduled_at\x18\v \x01(\x03R\vscheduledAt\x12\x1d\n" +
	"\n" +
	"started_at\x18\f \x01(\x03R\tstartedAt\x12!\n" +
	"\fcompleted_at\x18\r \x01(\x03R\vcompletedAt\x12\x1d\n" +
	"\n" +
	"timeout_ms\x18\x0e \x01(\x03R\ttimeoutMs\x12\x1b\n" +
	"\tworker_id\x18\x0f \x01(\tR\bworkerId\x12A\n" +
	"\bmetadata\x18\x10 \x03(\v2%.taskqueue.TaskResponse.MetadataEntryR\bmetadata\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"/\n" +
	"\x14GetTaskStatusRequest\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\",\n" +
	"\x11CancelTaskRequest\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\"H\n" +
	"\x12CancelTaskResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x18\n" +
	"\amessage\x18\x02 \x01(\tR\amessage\"b\n" +
	"\x12StreamTasksRequest\x12\x19\n" +
	"\btask_ids\x18\x01 \x03(\tR\ataskIds\x121\n" +
	"\bstatuses\x18\x02 \x03(\x0e2\x15.taskqueue.TaskStatusR\bstatuses\"r\n" +
	"\n" +
	"TaskUpdate\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\x12-\n" +
	"\x06status\x18\x02 \x01(\x0e2\x15.taskqueue.TaskStatusR\x06status\x12\x1c\n" +
	"\ttimestamp\x18\x03 \x01(\x03R\ttimestamp\"\xee\x01\n" +
	"\x15RegisterWorkerRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x16\n" +
	"\x06queues\x18\x02 \x03(\tR\x06queues\x12 \n" +
	"\vconcurrency\x18\x03 \x01(\x05R\vconcurrency\x12J\n" +
	"\bmetadata\x18\x04 \x03(\v2..taskqueue.RegisterWorkerRequest.MetadataEntryR\bmetadata\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\x93\x01\n" +
	"\x0eWorkerResponse\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x12\x16\n" +
	"\x06queues\x18\x03 \x03(\tR\x06queues\x12 \n" +
	"\vconcurrency\x18\x04 \x01(\x05R\vconcurrency\x12#\n" +
	"\rregistered_at\x18\x05 \x01(\x03R\fregisteredAt\"M\n" +
	"\x10HeartbeatRequest\x12\x1b\n" +
	"\tworker_id\x18\x01 \x01(\tR\bworkerId\x12\x1c\n" +
	"\ttimestamp\x18\x02 \x01(\x03R\ttimestamp\"T\n" +
	"\x11HeartbeatResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12%\n" +
	"\x0enext_heartbeat\x18\x02 \x01(\x03R\rnextHeartbeat*7\n" +
	"\bPriority\x12\a\n" +
	"\x03LOW\x10\x00\x12\n" +
	"\n" +
	"\x06NORMAL\x10\x01\x12\b\n" +
	"\x04HIGH\x10\x02\x12\f\n" +
	"\bCRITICAL\x10\x03*]\n" +
	"\n" +
	"TaskStatus\x12\v\n" +
	"\aPENDING\x10\x00\x12\x0e\n" +
	"\n" +
	"PROCESSING\x10\x01\x12\r\n" +
	"\tCOMPLETED\x10\x02\x12\n" +
	"\n" +
	"\x06FAILED\x10\x03\x12\b\n" +
	"\x04DEAD\x10\x04\x12\r\n" +
	"\tCANCELLED\x10\x052\xd1\x03\n" +
	"\x10TaskQueueService\x12C\n" +
	"\n" +
	"SubmitTask\x12\x1c.taskqueue.SubmitTaskRequest\x1a\x17.taskqueue.TaskResponse\x12I\n" +
	"\rGetTaskStatus\x12\x1f.taskqueue.GetTaskStatusRequest\x1a\x17.taskqueue.TaskResponse\x12I\n" +
	"\n" +
	"CancelTask\x12\x1c.taskqueue.CancelTaskRequest\x1a\x1d.taskqueue.CancelTaskResponse\x12E\n" +
	"\vStreamTasks\x12\x1d.taskqueue.StreamTasksRequest\x1a\x15.taskqueue.TaskUpdate0\x01\x12M\n" +
	"\x0eRegisterWorker\x12 .taskqueue.RegisterWorkerRequest\x1a\x19.taskqueue.WorkerResponse\x12L\n" +
	"\x0fUpdateHeartbeat\x12\x1b.taskqueue.HeartbeatRequest\x1a\x1c.taskqueue.HeartbeatResponseB8Z6github.com/ankurpal/distributed-task-queue/proto;protob\x06proto3"

var (
	file_proto_task_queue_proto_rawDescOnce sync.Once
	file_proto_task_queue_proto_rawDescData []byte
)

func file_proto_task_queue_proto_rawDescGZIP() []byte {
	file_proto_task_queue_proto_rawDescOnce.Do(func() {
		file_proto_task_queue_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_task_queue_proto_rawDesc), len(file_proto_task_queue_proto_rawDesc)))
	})
	return file_proto_task_queue_proto_rawDescData
}

var file_proto_task_queue_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_proto_task_queue_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_proto_task_queue_proto_goTypes = []any{
	(Priority)(0),
	(TaskStatus)(0),
	(*SubmitTaskRequest)(nil),
	(*TaskResponse)(nil),
	(*GetTaskStatusRequest)(nil),
	(*CancelTaskRequest)(nil),
	(*CancelTaskResponse)(nil),
	(*StreamTasksRequest)(nil),
	(*TaskUpdate)(nil),
	(*RegisterWorkerRequest)(nil),
	(*WorkerResponse)(nil),
	(*HeartbeatRequest)(nil),
	(*HeartbeatResponse)(nil),
	nil,
	nil,
	nil,
}
var file_proto_task_queue_proto_depIdxs = []int32{
	0,
	13,
	0,
	1,
	14,
	1,
	1,
	15,
	2,
	4,
	5,
	7,
	9,
	11,
	3,
	3,
	6,
	8,
	10,
	12,
	14,
	8,
	8,
	8,
	0,
}

func init() { file_proto_task_queue_proto_init() }
func file_proto_task_queue_proto_init() {
	if File_proto_task_queue_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_task_queue_proto_rawDesc), len(file_proto_task_queue_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_task_queue_proto_goTypes,
		DependencyIndexes: file_proto_task_queue_proto_depIdxs,
		EnumInfos:         file_proto_task_queue_proto_enumTypes,
		MessageInfos:      file_proto_task_queue_proto_msgTypes,
	}.Build()
	File_proto_task_queue_proto = out.File
	file_proto_task_queue_proto_goTypes = nil
	file_proto_task_queue_proto_depIdxs = nil
}
