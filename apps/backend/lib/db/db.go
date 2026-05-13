package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var driverName string

// rebind converts ? placeholders to $N for PostgreSQL
func rebind(query string) string {
	if driverName != "postgres" {
		return query
	}
	n := 0
	return strings.ReplaceAll(query, "?", func() string {
		n++
		return fmt.Sprintf("$%d", n)
	}())
}

// Exec wraps sql.DB.Exec with automatic placeholder rebinding
func Exec(query string, args ...any) (sql.Result, error) {
	return DB.Exec(rebind(query), args...)
}

// Query wraps sql.DB.Query with automatic placeholder rebinding
func Query(query string, args ...any) (*sql.Rows, error) {
	return DB.Query(rebind(query), args...)
}

// QueryRow wraps sql.DB.QueryRow with automatic placeholder rebinding
func QueryRow(query string, args ...any) *sql.Row {
	return DB.QueryRow(rebind(query), args...)
}

func Connect() error {
	driverName = os.Getenv("DB_DRIVER")
	if driverName == "" {
		driverName = "sqlite3"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		if driverName == "sqlite3" {
			home, _ := os.UserHomeDir()
			dbPath := filepath.Join(home, ".workagents", "data.db")
			os.MkdirAll(filepath.Dir(dbPath), 0755)
			dsn = dbPath
		} else {
			dsn = "postgres://localhost:5432/workagents?sslmode=disable"
		}
	}

	var err error
	DB, err = sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("db connect: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	if driverName == "postgres" {
		DB.SetMaxOpenConns(10)
		DB.SetMaxIdleConns(5)
	}

	log.Printf("[db] connected: %s", driverName)
	return nil
}

func Migrate() error {
	if DB == nil {
		return fmt.Errorf("db not connected")
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS companies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			goal TEXT NOT NULL DEFAULT '',
			budget_limit REAL DEFAULT 0,
			active INTEGER DEFAULT 1,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			company_id TEXT NOT NULL REFERENCES companies(id),
			name TEXT NOT NULL,
			role TEXT NOT NULL,
			adapter_type TEXT NOT NULL DEFAULT 'process',
			adapter_config TEXT DEFAULT '{}',
			reports_to TEXT REFERENCES agents(id),
			capabilities TEXT DEFAULT '',
			status TEXT DEFAULT 'active',
			budget_limit REAL DEFAULT 0,
			heartbeat_schedule TEXT DEFAULT '',
			context_mode TEXT DEFAULT 'thin',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			company_id TEXT NOT NULL REFERENCES companies(id),
			parent_id TEXT REFERENCES tasks(id),
			agent_id TEXT REFERENCES agents(id),
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			status TEXT DEFAULT 'backlog',
			priority INTEGER DEFAULT 0,
			billing_code TEXT DEFAULT '',
			budget_spent REAL DEFAULT 0,
			completed_at TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS task_comments (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL REFERENCES tasks(id),
			agent_id TEXT REFERENCES agents(id),
			content TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS heartbeats (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL REFERENCES agents(id),
			status TEXT DEFAULT 'pending',
			mode TEXT DEFAULT 'command',
			context_sent TEXT DEFAULT '{}',
			result TEXT DEFAULT '{}',
			started_at TEXT DEFAULT CURRENT_TIMESTAMP,
			completed_at TEXT,
			logs TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS budgets (
			id TEXT PRIMARY KEY,
			company_id TEXT NOT NULL REFERENCES companies(id),
			agent_id TEXT REFERENCES agents(id),
			period_start TEXT NOT NULL,
			period_end TEXT NOT NULL,
			limit_tokens INTEGER DEFAULT 0,
			limit_dollars REAL DEFAULT 0,
			spent_tokens INTEGER DEFAULT 0,
			spent_dollars REAL DEFAULT 0,
			alert_at REAL DEFAULT 0.80,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id TEXT PRIMARY KEY,
			company_id TEXT NOT NULL REFERENCES companies(id),
			actor_id TEXT NOT NULL,
			action TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			metadata TEXT DEFAULT '{}',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS agent_api_keys (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL REFERENCES agents(id),
			key_hash TEXT NOT NULL,
			name TEXT DEFAULT '',
			last_used_at TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS approval_requests (
			id TEXT PRIMARY KEY,
			company_id TEXT NOT NULL REFERENCES companies(id),
			request_type TEXT NOT NULL,
			requested_by TEXT REFERENCES agents(id),
			status TEXT DEFAULT 'pending',
			target_data TEXT DEFAULT '{}',
			reviewed_by TEXT DEFAULT '',
			reviewed_at TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS board_users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		// ── Security tables ──
		`CREATE TABLE IF NOT EXISTS company_members (
			user_id TEXT NOT NULL REFERENCES board_users(id),
			company_id TEXT NOT NULL REFERENCES companies(id),
			role TEXT NOT NULL DEFAULT 'member',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, company_id)
		)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES board_users(id),
			revoked INTEGER DEFAULT 0,
			expires_at TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		// ── Indexes for performance ──
		`CREATE INDEX IF NOT EXISTS idx_agents_company ON agents(company_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_company ON tasks(company_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_company ON activity_logs(company_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id)`,
	}

	for i, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}

	log.Println("[db] migrations complete")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
