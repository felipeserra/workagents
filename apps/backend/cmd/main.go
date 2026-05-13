package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/health", handleHealth)

	// Companies
	mux.HandleFunc("GET /api/companies", handleListCompanies)
	mux.HandleFunc("POST /api/companies", handleCreateCompany)
	mux.HandleFunc("GET /api/companies/{id}", handleGetCompany)

	// Agents
	mux.HandleFunc("GET /api/agents", handleListAgents)
	mux.HandleFunc("POST /api/agents", handleCreateAgent)
	mux.HandleFunc("GET /api/agents/{id}", handleGetAgent)

	// Tasks
	mux.HandleFunc("GET /api/tasks", handleListTasks)
	mux.HandleFunc("POST /api/tasks", handleCreateTask)

	// Heartbeats
	mux.HandleFunc("POST /api/heartbeats", handleTriggerHeartbeat)
	mux.HandleFunc("GET /api/heartbeats", handleListHeartbeats)

	// CORS middleware
	handler := corsMiddleware(mux)

	log.Printf("WorkAgents API starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ── Health ──

func handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Companies ──

func handleListCompanies(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, []map[string]any{
		{"id": "1", "name": "My AI Company", "goal": "Build the future"},
	})
}

func handleCreateCompany(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Goal string `json:"goal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]string{
		"id":   "new-id",
		"name": body.Name,
		"goal": body.Goal,
	})
}

func handleGetCompany(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	jsonResponse(w, http.StatusOK, map[string]string{"id": id, "name": "Company " + id})
}

// ── Agents ──

func handleListAgents(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")
	_ = companyID
	jsonResponse(w, http.StatusOK, []map[string]any{
		{"id": "1", "name": "CEO Agent", "role": "CEO", "status": "active"},
	})
}

func handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusCreated, map[string]string{"id": "new-agent", "status": "created"})
}

func handleGetAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	jsonResponse(w, http.StatusOK, map[string]string{"id": id, "status": "active"})
}

// ── Tasks ──

func handleListTasks(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, []map[string]any{
		{"id": "1", "title": "Build API", "status": "in_progress"},
	})
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusCreated, map[string]string{"id": "new-task", "status": "created"})
}

// ── Heartbeats ──

func handleTriggerHeartbeat(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusAccepted, map[string]string{"status": "heartbeat_dispatched"})
}

func handleListHeartbeats(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, []map[string]any{
		{"id": "1", "agent_id": "1", "status": "completed"},
	})
}

func init() {
	ctx := context.Background()
	_ = ctx
}
