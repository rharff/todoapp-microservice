package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB             *pgxpool.Pool
	AuditServiceURL string
	HTTPClient     *http.Client
}

func NewHandler(db *pgxpool.Pool, auditServiceURL string, client *http.Client) *Handler {
	return &Handler{DB: db, AuditServiceURL: strings.TrimRight(auditServiceURL, "/"), HTTPClient: client}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/tasks", h.GetTasks)
	r.Post("/tasks", h.CreateTask)
	r.Put("/tasks/{id}", h.UpdateTask)
	return r
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(r.Context(), `SELECT id, title, stage, position, created_at FROM tasks ORDER BY stage, position, created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Stage, &task.Position, &task.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan task")
			return
		}
		tasks = append(tasks, task)
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	stage := strings.TrimSpace(req.Stage)
	if stage == "" {
		stage = "todo"
	}
	if !isValidStage(stage) {
		writeError(w, http.StatusBadRequest, "invalid stage")
		return
	}

	var nextPosition int
	err := h.DB.QueryRow(r.Context(), `SELECT COALESCE(MAX(position), 0) + 1 FROM tasks WHERE stage = $1`, stage).Scan(&nextPosition)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute position")
		return
	}

	var task Task
	err = h.DB.QueryRow(r.Context(), `
		INSERT INTO tasks (title, stage, position)
		VALUES ($1, $2, $3)
		RETURNING id, title, stage, position, created_at
	`, req.Title, stage, nextPosition).Scan(&task.ID, &task.Title, &task.Stage, &task.Position, &task.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, task)
	go h.sendAuditEvent(context.Background(), AuditRequest{
		TaskID: task.ID,
		Action: "task_created",
		Payload: map[string]interface{}{
			"title": task.Title,
			"stage": task.Stage,
			"position": task.Position,
		},
	})
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "task id is required")
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if strings.TrimSpace(req.Stage) != "" && !isValidStage(req.Stage) {
		writeError(w, http.StatusBadRequest, "invalid stage")
		return
	}
	if strings.TrimSpace(req.Stage) == "" && strings.TrimSpace(req.Title) == "" && req.Position == nil {
		writeError(w, http.StatusBadRequest, "title, stage, or position is required")
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	var current Task
	err = tx.QueryRow(r.Context(), `SELECT id, title, stage, position, created_at FROM tasks WHERE id = $1`, id).
		Scan(&current.ID, &current.Title, &current.Stage, &current.Position, &current.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load task")
		return
	}

	updatedStage := current.Stage
	if strings.TrimSpace(req.Stage) != "" {
		updatedStage = strings.TrimSpace(req.Stage)
	}

	updatedPosition := current.Position
	if req.Position != nil {
		updatedPosition = *req.Position
	} else if updatedStage != current.Stage {
		if err := tx.QueryRow(r.Context(), `SELECT COALESCE(MAX(position), 0) + 1 FROM tasks WHERE stage = $1`, updatedStage).
			Scan(&updatedPosition); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to compute position")
			return
		}
	}

	updatedTitle := current.Title
	if strings.TrimSpace(req.Title) != "" {
		updatedTitle = strings.TrimSpace(req.Title)
	}

	var task Task
	err = tx.QueryRow(r.Context(), `
		UPDATE tasks
		SET title = $1,
			stage = $2,
			position = $3
		WHERE id = $4
		RETURNING id, title, stage, position, created_at
	`, updatedTitle, updatedStage, updatedPosition, id).Scan(&task.ID, &task.Title, &task.Stage, &task.Position, &task.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit task")
		return
	}

	writeJSON(w, http.StatusOK, task)
	go h.sendAuditEvent(context.Background(), AuditRequest{
		TaskID: task.ID,
		Action: "task_updated",
		Payload: map[string]interface{}{
			"title": task.Title,
			"stage": task.Stage,
			"position": task.Position,
		},
	})
}

func isValidStage(stage string) bool {
	switch stage {
	case "todo", "in_progress", "review", "done":
		return true
	default:
		return false
	}
}

func (h *Handler) sendAuditEvent(ctx context.Context, event AuditRequest) {
	if h.AuditServiceURL == "" || h.HTTPClient == nil {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.AuditServiceURL+"/logs", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
