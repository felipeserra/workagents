package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

type Task struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	ParentID    *string `json:"parent_id"`
	AgentID     *string `json:"agent_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    int     `json:"priority"`
	BillingCode string  `json:"billing_code"`
	BudgetSpent float64 `json:"budget_spent"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateTaskRequest struct {
	CompanyID   string  `json:"company_id"`
	ParentID    *string `json:"parent_id"`
	AgentID     *string `json:"agent_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    int     `json:"priority"`
}

func ListTasks(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")
	agentID := r.URL.Query().Get("agent_id")
	status := r.URL.Query().Get("status")

	query := "SELECT id, company_id, parent_id, agent_id, title, description, status, priority, billing_code, budget_spent, created_at, updated_at FROM tasks WHERE 1=1"
	args := []any{}

	if companyID != "" {
		query += " AND company_id = ?"
		args = append(args, companyID)
	}
	if agentID != "" {
		query += " AND agent_id = ?"
		args = append(args, agentID)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY priority DESC, created_at ASC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.CompanyID, &t.ParentID, &t.AgentID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.BillingCode, &t.BudgetSpent, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	jsonResponse(w, http.StatusOK, tasks)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var t Task
	err := db.DB.QueryRow("SELECT id, company_id, parent_id, agent_id, title, description, status, priority, billing_code, budget_spent, created_at, updated_at FROM tasks WHERE id=?", id).
		Scan(&t.ID, &t.CompanyID, &t.ParentID, &t.AgentID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.BillingCode, &t.BudgetSpent, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		jsonError(w, http.StatusNotFound, "task not found")
		return
	}

	// Get comments
	rows, _ := db.DB.Query("SELECT id, task_id, agent_id, content, created_at FROM task_comments WHERE task_id=? ORDER BY created_at ASC", id)
	defer rows.Close()
	type Comment struct {
		ID        string `json:"id"`
		TaskID    string `json:"task_id"`
		AgentID   string `json:"agent_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}
	comments := []Comment{}
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AgentID, &c.Content, &c.CreatedAt); err == nil {
			comments = append(comments, c)
		}
	}
	if comments == nil {
		comments = []Comment{}
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"task":     t,
		"comments": comments,
	})
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	boardID := getUserIDSafe(r)

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Title == "" {
		jsonError(w, http.StatusBadRequest, "title required")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.DB.Exec("INSERT INTO tasks (id, company_id, parent_id, agent_id, title, description, priority, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?)",
		id, req.CompanyID, req.ParentID, req.AgentID, req.Title, req.Description, req.Priority, now, now)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	logActivity(req.CompanyID, boardID, "task.created", "task", id, map[string]string{"title": req.Title})
	jsonResponse(w, http.StatusCreated, map[string]string{"id": id, "status": "created"})
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	db.DB.Exec("UPDATE tasks SET title=?, description=?, priority=?, agent_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", req.Title, req.Description, req.Priority, req.AgentID, id)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func CheckoutTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		jsonError(w, http.StatusBadRequest, "agent_id required")
		return
	}

	// Atomic checkout: only if status is 'backlog' or 'available'
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.DB.Exec("UPDATE tasks SET status='in_progress', agent_id=?, updated_at=? WHERE id=? AND status IN ('backlog','available')", agentID, now, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "checkout failed")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, http.StatusConflict, "task already assigned or completed")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "in_progress", "assigned_to": agentID})
}

func CompleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	now := time.Now().UTC().Format(time.RFC3339); _, err := db.DB.Exec("UPDATE tasks SET status='completed', completed_at=?, updated_at=? WHERE id=?", now, now, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to complete")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "completed"})
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	agentID := getUserIDSafe(r)

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		jsonError(w, http.StatusBadRequest, "content required")
		return
	}

	id := uuid.New().String()
	db.DB.Exec("INSERT INTO task_comments (id, task_id, agent_id, content) VALUES (?,?,?,?)", id, taskID, agentID, body.Content)
	jsonResponse(w, http.StatusCreated, map[string]string{"id": id, "status": "created"})
}
