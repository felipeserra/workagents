package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

// ── Types ──

type Company struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Goal        string  `json:"goal"`
	BudgetLimit float64 `json:"budget_limit"`
	Active      bool    `json:"active"`
	AgentCount  int     `json:"agent_count,omitempty"`
	TaskCount   int     `json:"task_count,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateCompanyRequest struct {
	Name string `json:"name"`
	Goal string `json:"goal"`
}

// ── Handlers ──

// ListCompanies retorna todas as empresas ativas
func ListCompanies(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.name, c.goal, c.budget_limit, c.active, c.created_at, c.updated_at,
			(SELECT COUNT(*) FROM agents a WHERE a.company_id = c.id AND a.status = 'active') as agent_count,
			(SELECT COUNT(*) FROM tasks t WHERE t.company_id = c.id AND t.status NOT IN ('cancelled')) as task_count
		FROM companies c
		WHERE c.active = 1
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	companies := []Company{}
	for rows.Next() {
		var c Company
		var active int
		if err := rows.Scan(&c.ID, &c.Name, &c.Goal, &c.BudgetLimit, &active, &c.CreatedAt, &c.UpdatedAt, &c.AgentCount, &c.TaskCount); err != nil {
			continue
		}
		c.Active = active == 1
		companies = append(companies, c)
	}

	if companies == nil {
		companies = []Company{}
	}

	jsonResponse(w, http.StatusOK, companies)
}

// GetCompany retorna detalhes de uma empresa
func GetCompany(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var c Company
	var active int
	err := db.DB.QueryRow(`
		SELECT id, name, goal, budget_limit, active, created_at, updated_at
		FROM companies WHERE id = ? AND active = 1
	`, id).Scan(&c.ID, &c.Name, &c.Goal, &c.BudgetLimit, &active, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		jsonError(w, http.StatusNotFound, "company not found")
		return
	}
	c.Active = active == 1

	jsonResponse(w, http.StatusOK, c)
}

// CreateCompany cria uma nova empresa
func CreateCompany(w http.ResponseWriter, r *http.Request) {
	boardID := getUserIDSafe(r)

	var req CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	req.Name = sanitizeText(req.Name)
	req.Goal = sanitizeText(req.Goal)

	if req.Name == "" {
		jsonError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Validate name length
	if len(req.Name) > 255 {
		jsonError(w, http.StatusBadRequest, "name too long (max 255)")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.DB.Exec(
		"INSERT INTO companies (id, name, goal, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		id, req.Name, req.Goal, now, now,
	)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create company")
		return
	}

	// Add creator as company member (owner)
	_, _ = db.DB.Exec(
		"INSERT INTO company_members (user_id, company_id, role) VALUES (?, ?, ?)",
		boardID, id, "owner",
	)

	// Activity log
	logActivity(id, boardID, "company.created", "company", id, map[string]string{"name": req.Name})

	jsonResponse(w, http.StatusCreated, Company{
		ID:        id,
		Name:      req.Name,
		Goal:      req.Goal,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// UpdateCompany atualiza dados da empresa
func UpdateCompany(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := getUserIDSafe(r)

	var req CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	req.Name = sanitizeText(req.Name)
	req.Goal = sanitizeText(req.Goal)

	now := time.Now().UTC().Format(time.RFC3339)

	result, err := db.DB.Exec(
		"UPDATE companies SET name = ?, goal = ?, updated_at = ? WHERE id = ? AND active = 1",
		req.Name, req.Goal, now, id,
	)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to update")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, http.StatusNotFound, "company not found")
		return
	}

	logActivity(id, boardID, "company.updated", "company", id, map[string]string{"name": req.Name})

	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteCompany soft-delete uma empresa
func DeleteCompany(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := getUserIDSafe(r)

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.DB.Exec("UPDATE companies SET active = 0, updated_at = ? WHERE id = ?", now, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to delete")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, http.StatusNotFound, "company not found")
		return
	}

	logActivity(id, boardID, "company.deleted", "company", id, nil)

	jsonResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}
