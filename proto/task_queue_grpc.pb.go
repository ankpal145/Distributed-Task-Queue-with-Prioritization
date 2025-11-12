package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	TaskQueueService_SubmitTask_FullMethodName      = "/taskqueue.TaskQueueService/SubmitTask"
	TaskQueueService_GetTaskStatus_FullMethodName   = "/taskqueue.TaskQueueService/GetTaskStatus"
	TaskQueueService_CancelTask_FullMethodName      = "/taskqueue.TaskQueueService/CancelTask"
	TaskQueueService_StreamTasks_FullMethodName     = "/taskqueue.TaskQueueService/StreamTasks"
	TaskQueueService_RegisterWorker_FullMethodName  = "/taskqueue.TaskQueueService/RegisterWorker"
	TaskQueueService_UpdateHeartbeat_FullMethodName = "/taskqueue.TaskQueueService/UpdateHeartbeat"
)

type TaskQueueServiceClient interface {
	SubmitTask(ctx context.Context, in *SubmitTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	GetTaskStatus(ctx context.Context, in *GetTaskStatusRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	CancelTask(ctx context.Context, in *CancelTaskRequest, opts ...grpc.CallOption) (*CancelTaskResponse, error)
	StreamTasks(ctx context.Context, in *StreamTasksRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[TaskUpdate], error)
	RegisterWorker(ctx context.Context, in *RegisterWorkerRequest, opts ...grpc.CallOption) (*WorkerResponse, error)
	UpdateHeartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
}

type taskQueueServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskQueueServiceClient(cc grpc.ClientConnInterface) TaskQueueServiceClient {
	return &taskQueueServiceClient{cc}
}

func (c *taskQueueServiceClient) SubmitTask(ctx context.Context, in *SubmitTaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, TaskQueueService_SubmitTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskQueueServiceClient) GetTaskStatus(ctx context.Context, in *GetTaskStatusRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, TaskQueueService_GetTaskStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskQueueServiceClient) CancelTask(ctx context.Context, in *CancelTaskRequest, opts ...grpc.CallOption) (*CancelTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CancelTaskResponse)
	err := c.cc.Invoke(ctx, TaskQueueService_CancelTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskQueueServiceClient) StreamTasks(ctx context.Context, in *StreamTasksRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[TaskUpdate], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &TaskQueueService_ServiceDesc.Streams[0], TaskQueueService_StreamTasks_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamTasksRequest, TaskUpdate]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type TaskQueueService_StreamTasksClient = grpc.ServerStreamingClient[TaskUpdate]

func (c *taskQueueServiceClient) RegisterWorker(ctx context.Context, in *RegisterWorkerRequest, opts ...grpc.CallOption) (*WorkerResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WorkerResponse)
	err := c.cc.Invoke(ctx, TaskQueueService_RegisterWorker_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskQueueServiceClient) UpdateHeartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, TaskQueueService_UpdateHeartbeat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type TaskQueueServiceServer interface {
	SubmitTask(context.Context, *SubmitTaskRequest) (*TaskResponse, error)
	GetTaskStatus(context.Context, *GetTaskStatusRequest) (*TaskResponse, error)
	CancelTask(context.Context, *CancelTaskRequest) (*CancelTaskResponse, error)
	StreamTasks(*StreamTasksRequest, grpc.ServerStreamingServer[TaskUpdate]) error
	RegisterWorker(context.Context, *RegisterWorkerRequest) (*WorkerResponse, error)
	UpdateHeartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	mustEmbedUnimplementedTaskQueueServiceServer()
}

type UnimplementedTaskQueueServiceServer struct{}

func (UnimplementedTaskQueueServiceServer) SubmitTask(context.Context, *SubmitTaskRequest) (*TaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTask not implemented")
}
func (UnimplementedTaskQueueServiceServer) GetTaskStatus(context.Context, *GetTaskStatusRequest) (*TaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTaskStatus not implemented")
}
func (UnimplementedTaskQueueServiceServer) CancelTask(context.Context, *CancelTaskRequest) (*CancelTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelTask not implemented")
}
func (UnimplementedTaskQueueServiceServer) StreamTasks(*StreamTasksRequest, grpc.ServerStreamingServer[TaskUpdate]) error {
	return status.Errorf(codes.Unimplemented, "method StreamTasks not implemented")
}
func (UnimplementedTaskQueueServiceServer) RegisterWorker(context.Context, *RegisterWorkerRequest) (*WorkerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterWorker not implemented")
}
func (UnimplementedTaskQueueServiceServer) UpdateHeartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateHeartbeat not implemented")
}
func (UnimplementedTaskQueueServiceServer) mustEmbedUnimplementedTaskQueueServiceServer() {}
func (UnimplementedTaskQueueServiceServer) testEmbeddedByValue()                          {}

type UnsafeTaskQueueServiceServer interface {
	mustEmbedUnimplementedTaskQueueServiceServer()
}

func RegisterTaskQueueServiceServer(s grpc.ServiceRegistrar, srv TaskQueueServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TaskQueueService_ServiceDesc, srv)
}

func _TaskQueueService_SubmitTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskQueueServiceServer).SubmitTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaskQueueService_SubmitTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskQueueServiceServer).SubmitTask(ctx, req.(*SubmitTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskQueueService_GetTaskStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTaskStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskQueueServiceServer).GetTaskStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaskQueueService_GetTaskStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskQueueServiceServer).GetTaskStatus(ctx, req.(*GetTaskStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskQueueService_CancelTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskQueueServiceServer).CancelTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaskQueueService_CancelTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskQueueServiceServer).CancelTask(ctx, req.(*CancelTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskQueueService_StreamTasks_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamTasksRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TaskQueueServiceServer).StreamTasks(m, &grpc.GenericServerStream[StreamTasksRequest, TaskUpdate]{ServerStream: stream})
}

type TaskQueueService_StreamTasksServer = grpc.ServerStreamingServer[TaskUpdate]

func _TaskQueueService_RegisterWorker_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterWorkerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskQueueServiceServer).RegisterWorker(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaskQueueService_RegisterWorker_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskQueueServiceServer).RegisterWorker(ctx, req.(*RegisterWorkerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskQueueService_UpdateHeartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskQueueServiceServer).UpdateHeartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaskQueueService_UpdateHeartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskQueueServiceServer).UpdateHeartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var TaskQueueService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "taskqueue.TaskQueueService",
	HandlerType: (*TaskQueueServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitTask",
			Handler:    _TaskQueueService_SubmitTask_Handler,
		},
		{
			MethodName: "GetTaskStatus",
			Handler:    _TaskQueueService_GetTaskStatus_Handler,
		},
		{
			MethodName: "CancelTask",
			Handler:    _TaskQueueService_CancelTask_Handler,
		},
		{
			MethodName: "RegisterWorker",
			Handler:    _TaskQueueService_RegisterWorker_Handler,
		},
		{
			MethodName: "UpdateHeartbeat",
			Handler:    _TaskQueueService_UpdateHeartbeat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamTasks",
			Handler:       _TaskQueueService_StreamTasks_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/task_queue.proto",
}
