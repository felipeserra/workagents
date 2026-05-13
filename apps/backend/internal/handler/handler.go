package handler

import (
	"encoding/json"
	"net/http"
)

// Health returns health check response
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListCompanies returns all companies for the authenticated user
func ListCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]any{
		{"id": "1", "name": "My AI Company"},
	})
}

// CreateCompany creates a new company
func CreateCompany(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Goal string `json:"goal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": "new-id", "name": body.Name})
}
