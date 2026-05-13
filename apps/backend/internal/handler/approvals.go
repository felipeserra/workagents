package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"workagents/apps/backend/internal/db"
)

type ApprovalRequest struct {
	ID           string `json:"id"`
	CompanyID    string `json:"company_id"`
	RequestType  string `json:"request_type"`
	RequestedBy  string `json:"requested_by"`
	Status       string `json:"status"`
	TargetData   string `json:"target_data"`
	ReviewedBy   string `json:"reviewed_by,omitempty"`
	ReviewedAt   string `json:"reviewed_at,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func CreateApproval(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CompanyID   string `json:"company_id"`
		RequestType string `json:"request_type"`
		RequestedBy string `json:"requested_by"`
		TargetData  string `json:"target_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.DB.Exec("INSERT INTO approval_requests (id, company_id, request_type, requested_by, target_data, created_at) VALUES (?,?,?,?,?,?)",
		id, req.CompanyID, req.RequestType, req.RequestedBy, req.TargetData, now)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create approval request")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]string{"id": id, "status": "pending"})
}

func ListApprovals(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	query := "SELECT id, company_id, request_type, requested_by, status, target_data, COALESCE(reviewed_by,''), COALESCE(reviewed_at,''), created_at FROM approval_requests"
	args := []any{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	approvals := []ApprovalRequest{}
	for rows.Next() {
		var a ApprovalRequest
		if err := rows.Scan(&a.ID, &a.CompanyID, &a.RequestType, &a.RequestedBy, &a.Status, &a.TargetData, &a.ReviewedBy, &a.ReviewedAt, &a.CreatedAt); err != nil {
			continue
		}
		approvals = append(approvals, a)
	}
	if approvals == nil {
		approvals = []ApprovalRequest{}
	}
	jsonResponse(w, http.StatusOK, approvals)
}

func ApproveApproval(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := r.Context().Value("user_id").(string)
	now := time.Now().UTC().Format(time.RFC3339)

	db.DB.Exec("UPDATE approval_requests SET status='approved', reviewed_by=?, reviewed_at=? WHERE id=? AND status='pending'", boardID, now, id)

	logActivity("", boardID, "approval.approved", "approval", id, nil)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "approved"})
}

func RejectApproval(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	boardID := r.Context().Value("user_id").(string)
	now := time.Now().UTC().Format(time.RFC3339)

	db.DB.Exec("UPDATE approval_requests SET status='rejected', reviewed_by=?, reviewed_at=? WHERE id=? AND status='pending'", boardID, now, id)

	logActivity("", boardID, "approval.rejected", "approval", id, nil)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "rejected"})
}
