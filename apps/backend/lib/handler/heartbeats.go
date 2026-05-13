package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

type Heartbeat struct {
	ID          string  `json:"id"`
	AgentID     string  `json:"agent_id"`
	Status      string  `json:"status"`
	Mode        string  `json:"mode"`
	ContextSent string  `json:"context_sent"`
	Result      string  `json:"result"`
	StartedAt   string  `json:"started_at"`
	CompletedAt *string `json:"completed_at"`
	Logs        string  `json:"logs"`
}

func TriggerHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentID string `json:"agent_id"`
		Mode    string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.AgentID == "" {
		jsonError(w, http.StatusBadRequest, "agent_id required")
		return
	}
	if req.Mode == "" {
		req.Mode = "command"
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.DB.Exec("INSERT INTO heartbeats (id, agent_id, status, mode, started_at) VALUES (?,?,?,?,?)",
		id, req.AgentID, "running", req.Mode, now)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to trigger heartbeat")
		return
	}

	jsonResponse(w, http.StatusAccepted, map[string]string{
		"id":     id,
		"status": "running",
	})
}

func ListHeartbeats(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	limit := parseLimit(r, 20, 500)

	query := "SELECT id, agent_id, status, mode, context_sent, result, started_at, completed_at, logs FROM heartbeats"
	args := []any{}

	if agentID != "" {
		query += " WHERE agent_id = ?"
		args = append(args, agentID)
	}

	query += " ORDER BY started_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	hbs := []Heartbeat{}
	for rows.Next() {
		var h Heartbeat
		if err := rows.Scan(&h.ID, &h.AgentID, &h.Status, &h.Mode, &h.ContextSent, &h.Result, &h.StartedAt, &h.CompletedAt, &h.Logs); err != nil {
			continue
		}
		hbs = append(hbs, h)
	}
	if hbs == nil {
		hbs = []Heartbeat{}
	}
	jsonResponse(w, http.StatusOK, hbs)
}

func GetHeartbeatLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var logs string
	err := db.DB.QueryRow("SELECT logs FROM heartbeats WHERE id=?", id).Scan(&logs)
	if err != nil {
		jsonError(w, http.StatusNotFound, "heartbeat not found")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"logs": logs})
}
