package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

type Agent struct {
	ID                string  `json:"id"`
	CompanyID         string  `json:"company_id"`
	Name              string  `json:"name"`
	Role              string  `json:"role"`
	AdapterType       string  `json:"adapter_type"`
	AdapterConfig     string  `json:"adapter_config"`
	ReportsTo         *string `json:"reports_to"`
	Capabilities      string  `json:"capabilities"`
	Status            string  `json:"status"`
	BudgetLimit       float64 `json:"budget_limit"`
	HeartbeatSchedule string  `json:"heartbeat_schedule"`
	ContextMode       string  `json:"context_mode"`
	TaskCount         int     `json:"task_count,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type CreateAgentRequest struct {
	CompanyID     string  `json:"company_id"`
	Name          string  `json:"name"`
	Role          string  `json:"role"`
	AdapterType   string  `json:"adapter_type"`
	AdapterConfig string  `json:"adapter_config"`
	ReportsTo     *string `json:"reports_to"`
	Capabilities  string  `json:"capabilities"`
	ContextMode   string  `json:"context_mode"`
}

func ListAgents(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")

	query := `SELECT id, company_id, name, role, adapter_type, adapter_config, reports_to,
		capabilities, status, budget_limit, heartbeat_schedule, context_mode, created_at, updated_at
		FROM agents WHERE status != 'terminated'`
	args := []any{}

	if companyID != "" {
		query += " AND company_id = ?"
		args = append(args, companyID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	agents := []Agent{}
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.CompanyID, &a.Name, &a.Role, &a.AdapterType,
			&a.AdapterConfig, &a.ReportsTo, &a.Capabilities, &a.Status, &a.BudgetLimit,
			&a.HeartbeatSchedule, &a.ContextMode, &a.CreatedAt, &a.UpdatedAt); err != nil {
			continue
		}
		agents = append(agents, a)
	}
	if agents == nil {
		agents = []Agent{}
	}

	jsonResponse(w, http.StatusOK, agents)
}

func GetAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var a Agent
	err := db.DB.QueryRow(`SELECT id, company_id, name, role, adapter_type, adapter_config,
		reports_to, capabilities, status, budget_limit, heartbeat_schedule, context_mode,
		created_at, updated_at FROM agents WHERE id = ?`, id).
		Scan(&a.ID, &a.CompanyID, &a.Name, &a.Role, &a.AdapterType, &a.AdapterConfig,
			&a.ReportsTo, &a.Capabilities, &a.Status, &a.BudgetLimit, &a.HeartbeatSchedule,
			&a.ContextMode, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		jsonError(w, http.StatusNotFound, "agent not found")
		return
	}
	jsonResponse(w, http.StatusOK, a)
}

func CreateAgent(w http.ResponseWriter, r *http.Request) {
	boardID := getUserIDSafe(r)

	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" || req.Role == "" {
		jsonError(w, http.StatusBadRequest, "name and role required")
		return
	}

	// Input sanitization
	req.Name = sanitizeText(req.Name)
	req.Role = sanitizeText(req.Role)

	if req.AdapterType == "" {
		req.AdapterType = "process"
	}

	// Validate adapter_config is valid JSON
	if req.AdapterConfig == "" {
		req.AdapterConfig = "{}"
	} else if req.AdapterConfig != "{}" {
		if !isValidJSON(req.AdapterConfig) {
			jsonError(w, http.StatusBadRequest, "adapter_config must be valid JSON")
			return
		}
	}
	if req.ContextMode == "" {
		req.ContextMode = "thin"
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.DB.Exec(`INSERT INTO agents
		(id, company_id, name, role, adapter_type, adapter_config, reports_to, capabilities, context_mode, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, req.CompanyID, req.Name, req.Role, req.AdapterType, req.AdapterConfig,
		req.ReportsTo, req.Capabilities, req.ContextMode, now, now)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create agent")
		return
	}

	logActivity(req.CompanyID, boardID, "agent.created", "agent", id, map[string]string{"name": req.Name})
	jsonResponse(w, http.StatusCreated, map[string]string{"id": id, "status": "created"})
}

func UpdateAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := getUserIDSafe(r)

	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	_, err := db.DB.Exec(`UPDATE agents SET name=?, role=?, adapter_type=?, adapter_config=?,
		reports_to=?, capabilities=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		req.Name, req.Role, req.AdapterType, req.AdapterConfig, req.ReportsTo, req.Capabilities, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to update")
		return
	}

	logActivity("", boardID, "agent.updated", "agent", id, map[string]string{"name": req.Name})
	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func DeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := getUserIDSafe(r)

	db.DB.Exec("UPDATE agents SET status='terminated', updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	logActivity("", boardID, "agent.deleted", "agent", id, nil)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func PauseAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := getUserIDSafe(r)

	db.DB.Exec("UPDATE agents SET status='paused', updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	logActivity("", boardID, "agent.paused", "agent", id, nil)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "paused"})
}

func ResumeAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := getUserIDSafe(r)

	db.DB.Exec("UPDATE agents SET status='active', updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	logActivity("", boardID, "agent.resumed", "agent", id, nil)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "active"})
}
