package grpcserver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/ankurpal/distributed-task-queue/internal/service"
	"github.com/ankurpal/distributed-task-queue/internal/worker"
	pb "github.com/ankurpal/distributed-task-queue/proto"
)

type Server struct {
	pb.UnimplementedTaskQueueServiceServer
	taskService   *service.TaskService
	workerService *worker.WorkerService
}

func NewServer(taskService *service.TaskService, workerService *worker.WorkerService) *Server {
	return &Server{
		taskService:   taskService,
		workerService: workerService,
	}
}

func (s *Server) SubmitTask(ctx context.Context, req *pb.SubmitTaskRequest) (*pb.TaskResponse, error) {
	priority := convertPriorityFromProto(req.Priority)

	scheduledAt := time.Now()
	if req.ScheduledAt > 0 {
		scheduledAt = time.Unix(req.ScheduledAt, 0)
	}

	timeout := 5 * time.Minute
	if req.TimeoutMs > 0 {
		timeout = time.Duration(req.TimeoutMs) * time.Millisecond
	}

	submitReq := &service.SubmitTaskRequest{
		Type:        req.Type,
		Priority:    priority,
		Payload:     json.RawMessage(req.Payload),
		MaxRetries:  int(req.MaxRetries),
		Timeout:     timeout,
		ScheduledAt: &scheduledAt,
		Metadata:    req.Metadata,
	}

	task, err := s.taskService.SubmitTask(ctx, submitReq)
	if err != nil {
		return nil, err
	}

	return convertTaskToProto(task), nil
}

func (s *Server) GetTaskStatus(ctx context.Context, req *pb.GetTaskStatusRequest) (*pb.TaskResponse, error) {
	task, err := s.taskService.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	return convertTaskToProto(task), nil
}

func (s *Server) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	err := s.taskService.CancelTask(ctx, req.TaskId)
	if err != nil {
		return &pb.CancelTaskResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.CancelTaskResponse{
		Success: true,
		Message: "task cancelled successfully",
	}, nil
}

func (s *Server) StreamTasks(req *pb.StreamTasksRequest, stream pb.TaskQueueService_StreamTasksServer) error {

	ctx := stream.Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for _, taskID := range req.TaskIds {
				task, err := s.taskService.GetTask(ctx, taskID)
				if err != nil {
					continue
				}

				if len(req.Statuses) > 0 {
					match := false
					taskStatus := convertTaskStatusToProto(task.Status)
					for _, status := range req.Statuses {
						if status == taskStatus {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}

				update := &pb.TaskUpdate{
					TaskId:    task.ID,
					Status:    convertTaskStatusToProto(task.Status),
					Timestamp: time.Now().Unix(),
				}

				if err := stream.Send(update); err != nil {
					return err
				}
			}
		}
	}
}

func (s *Server) RegisterWorker(ctx context.Context, req *pb.RegisterWorkerRequest) (*pb.WorkerResponse, error) {
	registerReq := &worker.RegisterWorkerRequest{
		Name:        req.Name,
		Queues:      req.Queues,
		Concurrency: int(req.Concurrency),
		Metadata:    req.Metadata,
	}

	worker, err := s.workerService.RegisterWorker(ctx, registerReq)
	if err != nil {
		return nil, err
	}

	return &pb.WorkerResponse{
		Id:           worker.ID,
		Name:         worker.Name,
		Queues:       worker.Queues,
		Concurrency:  int32(worker.Concurrency),
		RegisteredAt: worker.StartedAt.Unix(),
	}, nil
}

func (s *Server) UpdateHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	_, err := s.workerService.GetWorker(ctx, req.WorkerId)
	if err != nil {
		return &pb.HeartbeatResponse{
			Success:       false,
			NextHeartbeat: 0,
		}, err
	}

	return &pb.HeartbeatResponse{
		Success:       true,
		NextHeartbeat: time.Now().Add(5 * time.Second).Unix(),
	}, nil
}

func convertPriorityFromProto(p pb.Priority) entity.Priority {
	switch p {
	case pb.Priority_CRITICAL:
		return entity.PriorityCritical
	case pb.Priority_HIGH:
		return entity.PriorityHigh
	case pb.Priority_NORMAL:
		return entity.PriorityNormal
	case pb.Priority_LOW:
		return entity.PriorityLow
	default:
		return entity.PriorityNormal
	}
}

func convertPriorityToProto(p entity.Priority) pb.Priority {
	switch p {
	case entity.PriorityCritical:
		return pb.Priority_CRITICAL
	case entity.PriorityHigh:
		return pb.Priority_HIGH
	case entity.PriorityNormal:
		return pb.Priority_NORMAL
	case entity.PriorityLow:
		return pb.Priority_LOW
	default:
		return pb.Priority_NORMAL
	}
}

func convertTaskStatusToProto(s entity.TaskStatus) pb.TaskStatus {
	switch s {
	case entity.TaskStatusPending:
		return pb.TaskStatus_PENDING
	case entity.TaskStatusProcessing:
		return pb.TaskStatus_PROCESSING
	case entity.TaskStatusCompleted:
		return pb.TaskStatus_COMPLETED
	case entity.TaskStatusFailed:
		return pb.TaskStatus_FAILED
	case entity.TaskStatusDead:
		return pb.TaskStatus_DEAD
	case entity.TaskStatusCancelled:
		return pb.TaskStatus_CANCELLED
	default:
		return pb.TaskStatus_PENDING
	}
}

func convertTaskToProto(task *entity.Task) *pb.TaskResponse {
	resp := &pb.TaskResponse{
		Id:          task.ID,
		Type:        task.Type,
		Priority:    convertPriorityToProto(task.Priority),
		Payload:     []byte(task.Payload),
		Status:      convertTaskStatusToProto(task.Status),
		Result:      []byte(task.Result),
		Error:       task.Error,
		RetryCount:  int32(task.RetryCount),
		MaxRetries:  int32(task.MaxRetries),
		CreatedAt:   task.CreatedAt.Unix(),
		ScheduledAt: task.ScheduledAt.Unix(),
		TimeoutMs:   task.Timeout.Milliseconds(),
		WorkerId:    task.WorkerID,
		Metadata:    task.Metadata,
	}

	if task.StartedAt != nil {
		resp.StartedAt = task.StartedAt.Unix()
	}

	if task.CompletedAt != nil {
		resp.CompletedAt = task.CompletedAt.Unix()
	}

	return resp
}
