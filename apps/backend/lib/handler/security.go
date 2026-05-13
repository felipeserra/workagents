package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ──────────────────────────────────────────────────────────────────────────────
// Password Hashing (bcrypt — replaces SHA-256)
// ──────────────────────────────────────────────────────────────────────────────

// hashPasswordBcrypt generates a bcrypt hash for the given password.
// Uses cost 12 (balance of security & performance).
func hashPasswordBcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(bytes), nil
}

// checkPasswordBcrypt compares a password against a bcrypt hash.
func checkPasswordBcrypt(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// isBcryptHash returns true if the hash starts with the bcrypt prefix "$2".
func isBcryptHash(hash string) bool {
	return len(hash) >= 4 && (hash[:4] == "$2a$" || hash[:4] == "$2b$" || hash[:4] == "$2y$")
}

// hashPasswordLegacy is the old SHA-256 method — kept for migration only.
func hashPasswordLegacy(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}

// ──────────────────────────────────────────────────────────────────────────────
// Input Sanitization
// ──────────────────────────────────────────────────────────────────────────────

// sanitizeText removes control characters (except \n) and trims whitespace.
func sanitizeText(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
}

// ──────────────────────────────────────────────────────────────────────────────
// Rate Limiter (in-memory token bucket per IP)
// ──────────────────────────────────────────────────────────────────────────────

// RateLimiter implements a simple per-IP token bucket rate limiter.
// WARNING: In-memory only — for multi-instance HA, replace with Redis.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientBucket
	rate     int           // tokens per interval
	burst    int           // max burst
	interval time.Duration // refill interval
}

type clientBucket struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a new rate limiter.
// Example: NewRateLimiter(10, 20, time.Second) — 10 req/s, burst 20.
func NewRateLimiter(rate, burst int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		clients:  make(map[string]*clientBucket),
		rate:     rate,
		burst:    burst,
		interval: interval,
	}
}

// Allow checks if a request from the given key (IP + route) is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.clients[key]
	if !ok {
		rl.clients[key] = &clientBucket{tokens: rl.burst - 1, lastCheck: time.Now()}
		return true
	}

	// Refill tokens
	elapsed := time.Since(b.lastCheck)
	b.lastCheck = time.Now()
	b.tokens += int(elapsed / rl.interval) * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// RateLimitMiddleware returns an HTTP middleware that rate-limits per client IP.
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			// Extract real IP from headers if behind proxy
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = strings.Split(xff, ",")[0]
			} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
				ip = xri
			}
			key := ip + ":" + r.URL.Path

			if !rl.Allow(key) {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.interval.Seconds()))
				jsonError(w, http.StatusTooManyRequests, "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Security Headers Middleware
// ──────────────────────────────────────────────────────────────────────────────

// SecurityHeaders adds security-related HTTP headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Request Body Size Limit Middleware
// ──────────────────────────────────────────────────────────────────────────────

// MaxBodySize limits the request body to n bytes.
func MaxBodySize(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Cleanup Goroutine for Rate Limiter
// ──────────────────────────────────────────────────────────────────────────────

// StartCleanupRoutine periodically purges stale rate limiter entries.
func (rl *RateLimiter) StartCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for key, b := range rl.clients {
				if now.Sub(b.lastCheck) > interval*2 {
					delete(rl.clients, key)
				}
			}
			rl.mu.Unlock()
		}
	}()
	log.Printf("[security] rate limiter cleanup started (interval: %v)", interval)
}

// ──────────────────────────────────────────────────────────────────────────────
// JSON Validation
// ──────────────────────────────────────────────────────────────────────────────

// isValidJSON returns true if the string is valid JSON.
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
