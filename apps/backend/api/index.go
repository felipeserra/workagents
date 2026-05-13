package handler

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/felipeserra/workagents/apps/backend/lib/db"
	internalhandler "github.com/felipeserra/workagents/apps/backend/lib/handler"
)

var router http.Handler
var dbReady bool

func init() {
	// ── Database ──
	if err := db.Connect(); err != nil {
		log.Printf("[vercel] WARNING: db connect failed: %v", err)
		dbReady = false
	} else {
		if err := db.Migrate(); err != nil {
			log.Printf("[vercel] WARNING: db migrate failed: %v", err)
			dbReady = false
		} else {
			dbReady = true
			log.Println("[vercel] Database connected and migrated")
		}
	}

	// ── JWT secret — REQUIRED ──
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("[FATAL] JWT_SECRET environment variable is required")
	}
	internalhandler.SetJWTSecret([]byte(jwtSecret))

	// ── Router ──
	r := chi.NewRouter()

	// --- Middleware stack ---
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(internalhandler.MaxBodySize(1 << 20)) // 1 MB max body

	// Security headers
	r.Use(internalhandler.SecurityHeaders)

	// CORS — explicit whitelist, NOT wildcard
	allowedOrigins := getCORSOrigins()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Api-Key", "X-Request-Id"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiter — per IP, 30 req/s burst 50 for general endpoints
	rateLimiter := internalhandler.NewRateLimiter(30, 50, time.Second)
	rateLimiter.StartCleanupRoutine(10 * time.Minute)

	// Auth rate limiter — stricter for login (5 req/s burst 10)
	authLimiter := internalhandler.NewRateLimiter(5, 10, time.Second)

	// ── Routes ──
	r.Route("/api", func(r chi.Router) {
		// Public — auth endpoints (rate limited)
		r.Group(func(r chi.Router) {
			r.Use(internalhandler.RateLimitMiddleware(authLimiter))
			r.Get("/health", internalhandler.Health)
			r.Post("/auth/register", internalhandler.Register)
			r.Post("/auth/login", internalhandler.Login)
			r.Post("/auth/refresh", internalhandler.RefreshToken)
		})

		// Protected (JWT required)
		r.Group(func(r chi.Router) {
			r.Use(internalhandler.JWTAuth)
			r.Use(internalhandler.RateLimitMiddleware(rateLimiter))

			// Companies
			r.Get("/companies", internalhandler.ListCompanies)
			r.Post("/companies", internalhandler.CreateCompany)
			r.Get("/companies/{id}", internalhandler.GetCompany)
			r.Patch("/companies/{id}", internalhandler.UpdateCompany)
			r.Delete("/companies/{id}", internalhandler.DeleteCompany)

			// Agents
			r.Get("/agents", internalhandler.ListAgents)
			r.Post("/agents", internalhandler.CreateAgent)
			r.Get("/agents/{id}", internalhandler.GetAgent)
			r.Patch("/agents/{id}", internalhandler.UpdateAgent)
			r.Delete("/agents/{id}", internalhandler.DeleteAgent)
			r.Post("/agents/{id}/pause", internalhandler.PauseAgent)
			r.Post("/agents/{id}/resume", internalhandler.ResumeAgent)

			// Tasks
			r.Get("/tasks", internalhandler.ListTasks)
			r.Post("/tasks", internalhandler.CreateTask)
			r.Get("/tasks/{id}", internalhandler.GetTask)
			r.Patch("/tasks/{id}", internalhandler.UpdateTask)
			r.Post("/tasks/{id}/checkout", internalhandler.CheckoutTask)
			r.Post("/tasks/{id}/complete", internalhandler.CompleteTask)
			r.Post("/tasks/{id}/comment", internalhandler.AddComment)

			// Heartbeats
			r.Post("/heartbeats", internalhandler.TriggerHeartbeat)
			r.Get("/heartbeats", internalhandler.ListHeartbeats)
			r.Get("/heartbeats/{id}/logs", internalhandler.GetHeartbeatLogs)

			// Budgets
			r.Post("/budgets", internalhandler.CreateBudget)
			r.Get("/budgets", internalhandler.ListBudgets)
			r.Get("/budgets/usage", internalhandler.GetBudgetUsage)
			r.Patch("/budgets/{id}", internalhandler.UpdateBudget)
			r.Post("/budgets/{id}/override", internalhandler.OverrideBudget)

			// Activity
			r.Get("/activity", internalhandler.ListActivity)

			// Approvals
			r.Post("/approvals", internalhandler.CreateApproval)
			r.Get("/approvals", internalhandler.ListApprovals)
			r.Post("/approvals/{id}/approve", internalhandler.ApproveApproval)
			r.Post("/approvals/{id}/reject", internalhandler.RejectApproval)
		})
	})

	router = r
	log.Println("[vercel] WorkAgents API handler initialized (dbReady:", dbReady, ")")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

// getCORSOrigins returns the allowed origins list.
// Reads from ALLOWED_ORIGINS env var (comma-separated), with sensible defaults.
func getCORSOrigins() []string {
	env := os.Getenv("ALLOWED_ORIGINS")
	if env != "" {
		origins := []string{}
		for _, o := range splitAndTrim(env, ",") {
			if o != "" {
				origins = append(origins, o)
			}
		}
		if len(origins) > 0 {
			return origins
		}
	}

	// Sensible defaults (matching the Vercel deployments)
	return []string{
		"https://workagents.com.br",
		"https://app.workagents.com.br",
		"https://workagents-api.vercel.app",
		"https://workagents-landing.vercel.app",
		"http://localhost:3000",
		"http://localhost:5173",
	}
}

func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, part := range split(s, sep) {
		trimmed := trim(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
