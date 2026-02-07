package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	
	r.Post("/logs", h.CreateLog)
	r.Get("/logs", h.GetLogs)
	return r
}

func (h *Handler) CreateLog(w http.ResponseWriter, r *http.Request) {
	var req CreateLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if strings.TrimSpace(req.Action) == "" {
		writeError(w, http.StatusBadRequest, "action is required")
		return
	}

	var log AuditLog
	err := h.DB.QueryRow(r.Context(), `
		INSERT INTO audit_logs (task_id, action_string, payload)
		VALUES ($1, $2, $3)
		RETURNING id, task_id, action_string, payload, created_at
	`, req.TaskID, req.Action, req.Payload).Scan(&log.ID, &log.TaskID, &log.Action, &log.Payload, &log.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create audit log")
		return
	}

	writeJSON(w, http.StatusCreated, log)
}

func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(r.Context(), `
		SELECT id, task_id, action_string, payload, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch logs")
		return
	}
	defer rows.Close()

	logs := make([]AuditLog, 0)
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.TaskID, &log.Action, &log.Payload, &log.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan log")
			return
		}
		logs = append(logs, log)
	}

	writeJSON(w, http.StatusOK, logs)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
