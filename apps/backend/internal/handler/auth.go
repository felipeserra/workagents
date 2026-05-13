package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"workagents/apps/backend/internal/db"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "workagents-dev-secret-change-in-production"
	}
	jwtSecret = []byte(secret)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func hashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}

// Login autentica board user e retorna JWT
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.Email == "" || req.Password == "" {
		jsonError(w, http.StatusBadRequest, "email and password required")
		return
	}

	var user struct {
		ID           string
		Name         string
		Email        string
		PasswordHash string
	}
	err := db.DB.QueryRow(
		"SELECT id, name, email, password_hash FROM board_users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if user.PasswordHash != hashPassword(req.Password) {
		jsonError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  "board",
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Error signing JWT: %v", err)
		jsonError(w, http.StatusInternalServerError, "auth error")
		return
	}

	resp := LoginResponse{
		Token: tokenString,
	}
	resp.User.ID = user.ID
	resp.User.Name = user.Name
	resp.User.Email = user.Email

	jsonResponse(w, http.StatusOK, resp)
}

// Register cria novo board user
func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		jsonError(w, http.StatusBadRequest, "name, email and password required")
		return
	}

	id := uuid.New().String()
	hash := hashPassword(req.Password)

	_, err := db.DB.Exec(
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
