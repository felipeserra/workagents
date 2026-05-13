package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

// ── Context Keys ──

type contextKey string

const (
	ContextUserID     contextKey = "user_id"
	ContextUserEmail  contextKey = "user_email"
	ContextUserName   contextKey = "user_name"
	ContextUserRole   contextKey = "user_role"
	ContextAgentID    contextKey = "agent_id"
	ContextCompanyIDs contextKey = "company_ids"
)

// ── Safe Context Accessors ──

// getUserIDSafe extracts user_id from context safely.
// Returns empty string if not found or invalid type — no panic.
func getUserIDSafe(r *http.Request) string {
	v, ok := r.Context().Value(ContextUserID).(string)
	if !ok {
		return ""
	}
	return v
}

// getUserEmailSafe extracts user_email from context safely.
func getUserEmailSafe(r *http.Request) string {
	v, ok := r.Context().Value(ContextUserEmail).(string)
	if !ok {
		return ""
	}
	return v
}

// getUserNameSafe extracts user_name from context safely.
func getUserNameSafe(r *http.Request) string {
	v, ok := r.Context().Value(ContextUserName).(string)
	if !ok {
		return ""
	}
	return v
}

// ── Safe Limit Parsing ──

// parseLimit parses the "limit" query parameter safely.
// Returns parsed value, or defaultVal if missing/invalid.
// Caps at maxVal to prevent abuse.
func parseLimit(r *http.Request, defaultVal, maxVal int) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return defaultVal
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return defaultVal
	}
	if limit > maxVal {
		return maxVal
	}
	return limit
}

// ── JWT Auth Middleware ──

// JWTAuth middleware validates JWT token and injects user info into context.
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if jwtSecret == nil {
			jsonError(w, http.StatusInternalServerError, "server not configured for auth")
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			jsonError(w, http.StatusUnauthorized, "authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			jsonError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			jsonError(w, http.StatusUnauthorized, "empty token")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			jsonError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			jsonError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}

		// Extract claims with type-safe defaults
		ctx := r.Context()
		if sub, ok := claims["sub"].(string); ok {
			ctx = context.WithValue(ctx, ContextUserID, sub)
		}
		if email, ok := claims["email"].(string); ok {
			ctx = context.WithValue(ctx, ContextUserEmail, email)
		}
		if name, ok := claims["name"].(string); ok {
			ctx = context.WithValue(ctx, ContextUserName, name)
		}
		if role, ok := claims["role"].(string); ok {
			ctx = context.WithValue(ctx, ContextUserRole, role)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ── Agent API Key Auth Middleware ──

// AgentAuth middleware validates X-Api-Key against the agent_api_keys table.
func AgentAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Api-Key")
		if apiKey == "" {
			jsonError(w, http.StatusUnauthorized, "X-Api-Key header required")
			return
		}

		// Hash the key for lookup (we store key_hash, not the raw key)
		keyHash := sha256Hex(apiKey)

		var agentID string
		var lastUsedAt *string
		err := db.QueryRow(
			"SELECT agent_id FROM agent_api_keys WHERE key_hash = ?",
			keyHash,
		).Scan(&agentID)
		if err != nil {
			jsonError(w, http.StatusUnauthorized, "invalid API key")
			return
		}

		// Update last_used_at asynchronously (non-critical, ignore errors)
		now := time.Now().UTC().Format(time.RFC3339)
		go func() {
			_, _ = db.Exec("UPDATE agent_api_keys SET last_used_at = ? WHERE key_hash = ?", now, keyHash)
		}()

		// Verify agent exists and is active
		var status string
		err = db.QueryRow("SELECT status FROM agents WHERE id = ?", agentID).Scan(&status)
		if err != nil || status != "active" {
			_ = lastUsedAt // suppress unused
			jsonError(w, http.StatusForbidden, "agent not active or not found")
			return
		}

		ctx := context.WithValue(r.Context(), ContextAgentID, agentID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// ── Company Scoping Middleware ──

// CompanyScoped validates that the authenticated user has access to the given company_id.
// The company_id is extracted from URL query param, URL param, or request body.
func CompanyScoped(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := getUserIDSafe(r)
		if userID == "" {
			jsonError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		// Extract company_id from various locations
		companyID := r.URL.Query().Get("company_id")
		if companyID == "" {
			companyID = r.URL.Query().Get("companyId")
		}

		if companyID == "" {
			// Try URL param (from chi)
			if r.Context().Value("chi_company_id") != nil {
				if v, ok := r.Context().Value("chi_company_id").(string); ok {
					companyID = v
				}
			}
		}

		if companyID != "" {
			var count int
			err := db.QueryRow(
				"SELECT COUNT(*) FROM company_members WHERE user_id = ? AND company_id = ?",
				userID, companyID,
			).Scan(&count)
			if err != nil || count == 0 {
				jsonError(w, http.StatusForbidden, "access denied to this company")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
