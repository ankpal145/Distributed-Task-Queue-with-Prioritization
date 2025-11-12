package dto

type SubmitTaskDTO struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Priority string                 `json:"priority"`
	Payload  map[string]interface{} `json:"payload"`
	Queue    string                 `json:"queue,omitempty"`
}

type TaskResponseDTO struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Priority  string                 `json:"priority"`
	Queue     string                 `json:"queue"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}
