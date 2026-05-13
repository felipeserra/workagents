package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

type Budget struct {
	ID           string  `json:"id"`
	CompanyID    string  `json:"company_id"`
	AgentID      *string `json:"agent_id"`
	PeriodStart  string  `json:"period_start"`
	PeriodEnd    string  `json:"period_end"`
	LimitTokens  int64   `json:"limit_tokens"`
	LimitDollars float64 `json:"limit_dollars"`
	SpentTokens  int64   `json:"spent_tokens"`
	SpentDollars float64 `json:"spent_dollars"`
	AlertAt      float64 `json:"alert_at"`
	CreatedAt    string  `json:"created_at"`
}

func CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req Budget
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.CompanyID == "" {
		jsonError(w, http.StatusBadRequest, "company_id required")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	if req.AlertAt == 0 {
		req.AlertAt = 0.80
	}

	_, err := db.DB.Exec(`INSERT INTO budgets
		(id, company_id, agent_id, period_start, period_end, limit_tokens, limit_dollars, alert_at, created_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		id, req.CompanyID, req.AgentID, req.PeriodStart, req.PeriodEnd, req.LimitTokens, req.LimitDollars, req.AlertAt, now)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create budget")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]string{"id": id, "status": "created"})
}

func ListBudgets(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")

	query := "SELECT id, company_id, agent_id, period_start, period_end, limit_tokens, limit_dollars, spent_tokens, spent_dollars, alert_at, created_at FROM budgets WHERE 1=1"
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

	budgets := []Budget{}
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.ID, &b.CompanyID, &b.AgentID, &b.PeriodStart, &b.PeriodEnd, &b.LimitTokens, &b.LimitDollars, &b.SpentTokens, &b.SpentDollars, &b.AlertAt, &b.CreatedAt); err != nil {
			continue
		}
		budgets = append(budgets, b)
	}
	if budgets == nil {
		budgets = []Budget{}
	}
	jsonResponse(w, http.StatusOK, budgets)
}

func GetBudgetUsage(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")

	var totalTokens, totalDollars float64
	query := "SELECT COALESCE(SUM(spent_tokens),0), COALESCE(SUM(spent_dollars),0) FROM budgets"
	args := []any{}

	if companyID != "" {
		query += " WHERE company_id = ?"
		args = append(args, companyID)
	}

	db.DB.QueryRow(query, args...).Scan(&totalTokens, &totalDollars)

	jsonResponse(w, http.StatusOK, map[string]float64{
		"total_tokens": totalTokens,
		"total_dollars": totalDollars,
	})
}

func UpdateBudget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		LimitTokens  int64   `json:"limit_tokens"`
		LimitDollars float64 `json:"limit_dollars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	db.DB.Exec("UPDATE budgets SET limit_tokens=?, limit_dollars=?, updated_at=? WHERE id=?",
		req.LimitTokens, req.LimitDollars, now, id)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func OverrideBudget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		LimitTokens  int64   `json:"limit_tokens"`
		LimitDollars float64 `json:"limit_dollars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	db.DB.Exec("UPDATE budgets SET limit_tokens=?, limit_dollars=?, updated_at=? WHERE id=?",
		req.LimitTokens, req.LimitDollars, now, id)

	// Log board override
	boardID := getUserIDSafe(r)
	logActivity("", boardID, "budget.override", "budget", id, nil)

	jsonResponse(w, http.StatusOK, map[string]string{"status": "overridden"})
}
