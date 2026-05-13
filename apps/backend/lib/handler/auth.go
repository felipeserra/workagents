package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/felipeserra/workagents/apps/backend/lib/db"
)

var jwtSecret []byte
var jwtSecretLoaded bool

// SetJWTSecret configures the JWT signing key.
// Must be called during init before any request is handled.
// Enforces minimum 32-byte key length for HS256.
func SetJWTSecret(secret []byte) {
	if len(secret) < 32 {
		log.Fatalf("[FATAL] JWT secret must be at least 32 bytes, got %d", len(secret))
	}
	jwtSecret = secret
	jwtSecretLoaded = true
	log.Println("[auth] JWT secret configured")
}

// initJWT loads the JWT secret from the environment.
// CRITICAL: Does NOT fallback to a hardcoded dev secret.
// If JWT_SECRET is not set, the service will refuse to start.
func initJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required — set it to a 256+ bit random secret")
	}
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters (256 bits) for HS256")
	}
	jwtSecret = []byte(secret)
	jwtSecretLoaded = true
	return nil
}

// ── Types ──

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token         string `json:"token"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	ExpiresIn     int    `json:"expires_in"`
	User          UserInfo `json:"user"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ── Handlers ──

// Login autentica board user e retorna JWT (1h) + refresh token (7d)
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	// Sanitize and validate
	req.Email = sanitizeText(req.Email)
	if req.Email == "" || req.Password == "" {
		jsonError(w, http.StatusBadRequest, "email and password required")
		return
	}

	// Lookup user
	var user struct {
		ID           string
		Name         string
		Email        string
		PasswordHash string
	}
	err := db.QueryRow(
		"SELECT id, name, email, password_hash FROM board_users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check password — supports both bcrypt (new) and SHA-256 (legacy for migration)
	valid := checkPasswordBcrypt(req.Password, user.PasswordHash)
	if !valid && isBcryptHash(user.PasswordHash) {
		jsonError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Legacy SHA-256 migration: re-hash to bcrypt on successful login
	if !isBcryptHash(user.PasswordHash) {
		if hashPasswordLegacy(req.Password) != user.PasswordHash {
			jsonError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		// Upgrade to bcrypt
		newHash, err := hashPasswordBcrypt(req.Password)
		if err == nil {
			db.Exec("UPDATE board_users SET password_hash = ? WHERE id = ?", newHash, user.ID)
		}
	}

	// Generate access token (1h)
	now := time.Now()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  "board",
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
		"jti":   uuid.New().String(),
	})

	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		log.Printf("[auth] error signing JWT: %v", err)
		jsonError(w, http.StatusInternalServerError, "auth error")
		return
	}

	// Generate refresh token (7d)
	refreshToken := uuid.New().String()
	refreshExpiresAt := now.Add(7 * 24 * time.Hour)

	_, err = db.Exec(
		"INSERT INTO refresh_tokens (id, user_id, expires_at) VALUES (?, ?, ?)",
		refreshToken, user.ID, refreshExpiresAt.Format(time.RFC3339),
	)
	if err != nil {
		log.Printf("[auth] error storing refresh token: %v", err)
		// Non-fatal — token still works, but refresh won't
	}

	jsonResponse(w, http.StatusOK, LoginResponse{
		Token:        accessTokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

// RefreshToken gera um novo access token a partir de um refresh token válido
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.RefreshToken == "" {
		jsonError(w, http.StatusBadRequest, "refresh_token required")
		return
	}

	// Validate refresh token
	var userID, expiresAt string
	err := db.QueryRow(
		"SELECT user_id, expires_at FROM refresh_tokens WHERE id = ? AND revoked = 0",
		req.RefreshToken,
	).Scan(&userID, &expiresAt)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	// Check expiry
	expTime, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil || time.Now().After(expTime) {
		jsonError(w, http.StatusUnauthorized, "refresh token expired")
		return
	}

	// Fetch user
	var user UserInfo
	err = db.QueryRow(
		"SELECT id, name, email FROM board_users WHERE id = ?", userID,
	).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "user not found")
		return
	}

	// Revoke old refresh token (one-time use)
	db.Exec("UPDATE refresh_tokens SET revoked = 1 WHERE id = ?", req.RefreshToken)

	// Issue new access token (1h)
	now := time.Now()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  "board",
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
		"jti":   uuid.New().String(),
	})

	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		log.Printf("[auth] error signing JWT: %v", err)
		jsonError(w, http.StatusInternalServerError, "auth error")
		return
	}

	// Issue new refresh token (7d, rotation)
	newRefreshToken := uuid.New().String()
	refreshExpiresAt := now.Add(7 * 24 * time.Hour)
	db.Exec(
		"INSERT INTO refresh_tokens (id, user_id, expires_at) VALUES (?, ?, ?)",
		newRefreshToken, userID, refreshExpiresAt.Format(time.RFC3339),
	)

	jsonResponse(w, http.StatusOK, LoginResponse{
		Token:        accessTokenString,
		RefreshToken: newRefreshToken,
		ExpiresIn:    3600,
		User:         user,
	})
}

// Register cria novo board user com bcrypt hash
func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	req.Name = sanitizeText(req.Name)
	req.Email = sanitizeText(req.Email)

	if req.Name == "" || req.Email == "" || req.Password == "" {
		jsonError(w, http.StatusBadRequest, "name, email and password required")
		return
	}

	// Password strength validation
	if len(req.Password) < 8 {
		jsonError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	id := uuid.New().String()
	hash, err := hashPasswordBcrypt(req.Password)
	if err != nil {
		log.Printf("[auth] bcrypt error: %v", err)
		jsonError(w, http.StatusInternalServerError, "registration error")
		return
	}

	_, err = db.Exec(
		"INSERT INTO board_users (id, email, password_hash, name) VALUES (?, ?, ?, ?)",
		id, req.Email, hash, req.Name,
	)
	if err != nil {
		jsonError(w, http.StatusConflict, "email already registered")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]string{
		"id":    id,
		"name":  req.Name,
		"email": req.Email,
	})
}
