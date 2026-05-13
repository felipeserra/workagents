package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"workagents/apps/backend/internal/db"
	"workagents/apps/backend/internal/handler"
	"workagents/apps/backend/internal/middleware"
)

func main() {
	// Database
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Api-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ── Public routes ──
	r.Route("/api", func(r chi.Router) {
		// Health
		r.Get("/health", handler.Health)

		// Auth (public)
		r.Post("/auth/register", handler.Register)
		r.Post("/auth/login", handler.Login)

		// ── Protected routes (Board JWT) ──
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)

			// Companies
			r.Get("/companies", handler.ListCompanies)
			r.Post("/companies", handler.CreateCompany)
			r.Get("/companies/{id}", handler.GetCompany)
			r.Patch("/companies/{id}", handler.UpdateCompany)
			r.Delete("/companies/{id}", handler.DeleteCompany)

			// Agents
			r.Get("/agents", handler.ListAgents)
			r.Post("/agents", handler.CreateAgent)
			r.Get("/agents/{id}", handler.GetAgent)
			r.Patch("/agents/{id}", handler.UpdateAgent)
			r.Delete("/agents/{id}", handler.DeleteAgent)
			r.Post("/agents/{id}/pause", handler.PauseAgent)
			r.Post("/agents/{id}/resume", handler.ResumeAgent)

			// Tasks
			r.Get("/tasks", handler.ListTasks)
			r.Post("/tasks", handler.CreateTask)
			r.Get("/tasks/{id}", handler.GetTask)
			r.Patch("/tasks/{id}", handler.UpdateTask)
			r.Post("/tasks/{id}/checkout", handler.CheckoutTask)
			r.Post("/tasks/{id}/complete", handler.CompleteTask)
			r.Post("/tasks/{id}/comment", handler.AddComment)

			// Heartbeats
			r.Post("/heartbeats", handler.TriggerHeartbeat)
			r.Get("/heartbeats", handler.ListHeartbeats)
			r.Get("/heartbeats/{id}/logs", handler.GetHeartbeatLogs)

			// Budgets
			r.Post("/budgets", handler.CreateBudget)
			r.Get("/budgets", handler.ListBudgets)
			r.Get("/budgets/usage", handler.GetBudgetUsage)
			r.Patch("/budgets/{id}", handler.UpdateBudget)
			r.Post("/budgets/{id}/override", handler.OverrideBudget)

			// Activity
			r.Get("/activity", handler.ListActivity)

			// Approvals
			r.Post("/approvals", handler.CreateApproval)
			r.Get("/approvals", handler.ListApprovals)
			r.Post("/approvals/{id}/approve", handler.ApproveApproval)
			r.Post("/approvals/{id}/reject", handler.RejectApproval)
		})
	})

	// Start
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("[server] WorkAgents API starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
