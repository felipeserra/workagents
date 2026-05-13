package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey for request context
type contextKey string

const (
	ContextUserID    contextKey = "user_id"
	ContextUserEmail contextKey = "user_email"
	ContextUserName  contextKey = "user_name"
	ContextUserRole  contextKey = "user_role"
)

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// JWTAuth middleware validates JWT token and injects user info into context
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			jsonError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			jsonError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserID, claims["sub"])
		ctx = context.WithValue(ctx, ContextUserEmail, claims["email"])
		ctx = context.WithValue(ctx, ContextUserName, claims["name"])
		ctx = context.WithValue(ctx, ContextUserRole, claims["role"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AgentAuth middleware validates agent API key
func AgentAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Api-Key")
		if apiKey == "" {
			jsonError(w, http.StatusUnauthorized, "X-Api-Key header required")
			return
		}

		// TODO: validate against agent_api_keys table
		_ = apiKey

		next.ServeHTTP(w, r)
	})
}
