package httpapi

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID        int             `json:"id"`
	TaskID    *string         `json:"task_id"`
	Action    string          `json:"action_string"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateLogRequest struct {
	TaskID  *string         `json:"task_id"`
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
