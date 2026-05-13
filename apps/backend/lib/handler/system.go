package handler

import (
	"net/http"

	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

type ActivityEntry struct {
	ID         string `json:"id"`
	CompanyID  string `json:"company_id"`
	ActorID    string `json:"actor_id"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Metadata   string `json:"metadata"`
	CreatedAt  string `json:"created_at"`
}

// ListActivity retorna feed de atividades
func ListActivity(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")
	limit := parseLimit(r, 50, 500)

	query := `SELECT id, company_id, actor_id, action, target_type, target_id, metadata, created_at
		FROM activity_logs`
	args := []any{}

	if companyID != "" {
		query += " WHERE company_id = ?"
		args = append(args, companyID)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	if rows != nil {
		defer rows.Close()
	} else {
		defer func() {}()
	}

	entries := []ActivityEntry{}
	for rows.Next() {
		var e ActivityEntry
		if err := rows.Scan(&e.ID, &e.CompanyID, &e.ActorID, &e.Action, &e.TargetType, &e.TargetID, &e.Metadata, &e.CreatedAt); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	if entries == nil {
		entries = []ActivityEntry{}
	}

	jsonResponse(w, http.StatusOK, entries)
}

// Health retorna status da API
func Health(w http.ResponseWriter, r *http.Request) {
	err := db.DB.Ping()
	status := "ok"
	if err != nil {
		status = "degraded"
	}

	jsonResponse(w, http.StatusOK, map[string]string{
		"status":  status,
		"service": "workagents-api",
		"version": "0.1.0",
	})
}
