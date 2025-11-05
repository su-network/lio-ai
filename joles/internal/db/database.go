package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"lio-ai/internal/config"
)

// Database represents the database connection
type Database struct {
	conn *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*Database, error) {
	// Ensure DB directory exists if DSN includes a directory
	if dir := filepath.Dir(cfg.Database.DSN); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create database directory '%s': %w", dir, err)
		}
	}

	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc", cfg.Database.DSN)
	
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connection established")

	// Run migrations
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{conn: db}, nil
}

// migrate runs database migrations
func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_documents_title ON documents(title);

	CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id VARCHAR(255) NOT NULL,
		title VARCHAR(255) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_chats_user_id ON chats(user_id);
	CREATE INDEX IF NOT EXISTS idx_chats_updated_at ON chats(updated_at DESC);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		role VARCHAR(50) NOT NULL,
		content TEXT NOT NULL,
		model VARCHAR(100),
		tokens INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

	CREATE TABLE IF NOT EXISTS usage_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id VARCHAR(255) NOT NULL,
		request_type VARCHAR(50) NOT NULL,
		resource_id INTEGER,
		tokens_input INTEGER DEFAULT 0,
		tokens_output INTEGER DEFAULT 0,
		tokens_total INTEGER DEFAULT 0,
		model_used VARCHAR(100),
		cost_usd REAL DEFAULT 0.0,
		duration_ms INTEGER DEFAULT 0,
		endpoint VARCHAR(255),
		success BOOLEAN DEFAULT 1,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_usage_user_id ON usage_metrics(user_id);
	CREATE INDEX IF NOT EXISTS idx_usage_created_at ON usage_metrics(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_usage_request_type ON usage_metrics(request_type);

	CREATE TABLE IF NOT EXISTS user_quotas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id VARCHAR(255) NOT NULL UNIQUE,
		daily_token_limit INTEGER DEFAULT 100000,
		monthly_token_limit INTEGER DEFAULT 3000000,
		daily_tokens_used INTEGER DEFAULT 0,
		monthly_tokens_used INTEGER DEFAULT 0,
		daily_cost_limit_usd REAL DEFAULT 10.0,
		monthly_cost_limit_usd REAL DEFAULT 300.0,
		daily_cost_used_usd REAL DEFAULT 0.0,
		monthly_cost_used_usd REAL DEFAULT 0.0,
		last_reset_daily DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_reset_monthly DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_quotas_user_id ON user_quotas(user_id);

	CREATE TABLE IF NOT EXISTS cost_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model_name VARCHAR(100) NOT NULL UNIQUE,
		cost_per_input_token REAL NOT NULL,
		cost_per_output_token REAL NOT NULL,
		operation_type VARCHAR(50) NOT NULL,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_cost_model_name ON cost_config(model_name);

	-- Insert default cost configurations
	INSERT OR IGNORE INTO cost_config (model_name, cost_per_input_token, cost_per_output_token, operation_type, is_active)
	VALUES 
		('gpt-4', 0.00003, 0.00006, 'chat', 1),
		('gpt-3.5-turbo', 0.0000015, 0.000002, 'chat', 1),
		('claude-3-opus', 0.000015, 0.000075, 'chat', 1),
		('claude-3-sonnet', 0.000003, 0.000015, 'chat', 1),
		('qwen-2.5-coder', 0.000001, 0.000002, 'code_generation', 1),
		('codellama-34b', 0.0000008, 0.0000016, 'code_generation', 1),
		('default', 0.000001, 0.000002, 'chat', 1);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}
	log.Println("✓ Database migrations completed")
	return nil
}

// GetConnection returns the underlying database connection
func (d *Database) GetConnection() *sql.DB {
	return d.conn
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.conn.Close()
}
