package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TaskClient struct {
	baseURL string
	client  *http.Client
}

func NewTaskClient(baseURL string) *TaskClient {
	return &TaskClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SubmitTaskRequest struct {
	Type       string                 `json:"type"`
	Priority   int                    `json:"priority"`
	Payload    map[string]interface{} `json:"payload"`
	MaxRetries int                    `json:"max_retries,omitempty"`
	Timeout    int64                  `json:"timeout,omitempty"`
}

type TaskResponse struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    int                    `json:"priority"`
	Status      string                 `json:"status"`
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt time.Time              `json:"scheduled_at"`
}

func (c *TaskClient) SubmitTask(req *SubmitTaskRequest) (*TaskResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(
		c.baseURL+"/api/v1/tasks",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var task TaskResponse
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &task, nil
}

func (c *TaskClient) GetTask(taskID string) (*TaskResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/api/v1/tasks/" + taskID)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var task TaskResponse
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &task, nil
}

func main() {
	client := NewTaskClient("http://localhost:8080")

	task, err := client.SubmitTask(&SubmitTaskRequest{
		Type:     "email",
		Priority: 30,
		Payload: map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Welcome!",
			"body":    "Thank you for signing up.",
		},
		MaxRetries: 3,
		Timeout:    300000000000,
	})

	if err != nil {
		fmt.Printf("Error submitting task: %v\n", err)
		return
	}

	fmt.Printf("Task submitted successfully!\n")
	fmt.Printf("Task ID: %s\n", task.ID)
	fmt.Printf("Status: %s\n", task.Status)

	fmt.Println("\nWaiting for task to complete...")
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)

		task, err := client.GetTask(task.ID)
		if err != nil {
			fmt.Printf("Error getting task: %v\n", err)
			continue
		}

		fmt.Printf("Status: %s\n", task.Status)

		if task.Status == "completed" {
			fmt.Printf("\nTask completed successfully!\n")
			if task.Result != nil {
				fmt.Printf("Result: %+v\n", task.Result)
			}
			break
		} else if task.Status == "failed" || task.Status == "dead" {
			fmt.Printf("\nTask failed: %s\n", task.Error)
			break
		}
	}
}
