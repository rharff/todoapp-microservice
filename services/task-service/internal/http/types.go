package httpapi

import "time"

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Stage     string    `json:"stage"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
	Stage string `json:"stage"`
}

type UpdateTaskRequest struct {
	Title    string `json:"title"`
	Stage    string `json:"stage"`
	Position *int   `json:"position"`
}

type AuditRequest struct {
	TaskID  string      `json:"task_id"`
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
