package resthandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/ankurpal/distributed-task-queue/internal/service"
	"github.com/ankurpal/distributed-task-queue/internal/worker"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	taskService   *service.TaskService
	workerService *worker.WorkerService
}

func NewHandler(taskService *service.TaskService, workerService *worker.WorkerService) *Handler {
	return &Handler{
		taskService:   taskService,
		workerService: workerService,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/health", h.Health)

		tasks := api.Group("/tasks")
		{
			tasks.POST("", h.SubmitTask)
			tasks.GET("/:id", h.GetTask)
			tasks.DELETE("/:id", h.CancelTask)
			tasks.GET("", h.ListTasks)
		}

		dlq := api.Group("/dlq")
		{
			dlq.GET("", h.GetDeadLetterQueue)
			dlq.POST("/:id/retry", h.RetryDeadTask)
		}

		workers := api.Group("/workers")
		{
			workers.POST("/register", h.RegisterWorker)
			workers.GET("/:id", h.GetWorker)
			workers.DELETE("/:id", h.DeregisterWorker)
			workers.GET("", h.ListWorkers)
			workers.GET("/:id/stats", h.GetWorkerStats)
		}

		queues := api.Group("/queues")
		{
			queues.GET("/stats", h.GetQueueStats)
		}

		system := api.Group("/system")
		{
			system.GET("/stats", h.GetSystemStats)
		}
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "task-queue",
	})
}

func (h *Handler) SubmitTask(c *gin.Context) {
	var req service.SubmitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.SubmitTask(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")

	err := h.taskService.CancelTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task cancelled"})
}

func (h *Handler) ListTasks(c *gin.Context) {
	filter := &entity.TaskFilter{
		Type:   c.Query("type"),
		Limit:  100,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = []entity.TaskStatus{entity.TaskStatus(statusStr)}
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			filter.Priority = []entity.Priority{entity.Priority(priority)}
		}
	}

	tasks, err := h.taskService.ListTasks(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) GetDeadLetterQueue(c *gin.Context) {
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	tasks, err := h.taskService.GetDeadLetterQueue(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) RetryDeadTask(c *gin.Context) {
	taskID := c.Param("id")

	err := h.taskService.RetryDeadTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task requeued"})
}

func (h *Handler) RegisterWorker(c *gin.Context) {
	var req worker.RegisterWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	worker, err := h.workerService.RegisterWorker(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, worker)
}

func (h *Handler) GetWorker(c *gin.Context) {
	workerID := c.Param("id")

	worker, err := h.workerService.GetWorker(c.Request.Context(), workerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
		return
	}

	c.JSON(http.StatusOK, worker)
}

func (h *Handler) DeregisterWorker(c *gin.Context) {
	workerID := c.Param("id")

	err := h.workerService.DeregisterWorker(c.Request.Context(), workerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "worker deregistered"})
}

func (h *Handler) ListWorkers(c *gin.Context) {
	workers, err := h.workerService.ListWorkers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workers)
}

func (h *Handler) GetWorkerStats(c *gin.Context) {
	workerID := c.Param("id")

	stats, err := h.workerService.GetWorkerStats(c.Request.Context(), workerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetQueueStats(c *gin.Context) {
	stats, err := h.taskService.GetQueueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetSystemStats(c *gin.Context) {
	queueStats, err := h.taskService.GetQueueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	workerStats, err := h.workerService.GetAllWorkerStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var totalDepth int64
	for _, queue := range queueStats {
		totalDepth += queue.Depth
	}

	var totalTasksProcessed int64
	var totalTasksFailed int64
	for _, worker := range workerStats {
		totalTasksProcessed += worker.TasksProcessed
		totalTasksFailed += worker.TasksFailed
	}

	c.JSON(http.StatusOK, gin.H{
		"queues": queueStats,
		"workers": gin.H{
			"total": len(workerStats),
			"stats": workerStats,
		},
		"tasks": gin.H{
			"total_in_queue":  totalDepth,
			"total_processed": totalTasksProcessed,
			"total_failed":    totalTasksFailed,
		},
	})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ParseJSONPayload(data string) json.RawMessage {
	return json.RawMessage(data)
}
