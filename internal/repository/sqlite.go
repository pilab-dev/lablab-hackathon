package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// PromptRepository persists LLM prompt/response pairs for auditing and replay.
type PromptRepository interface {
	SaveLLMExchange(ctx context.Context, userPrompt, assistantJSON string) error
	Close() error
}

// SQLiteRepository stores prompt logs in a local SQLite file.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository opens (or creates) a SQLite database at path.
// An empty path returns an error so callers can fall back to nil.
func NewSQLiteRepository(path string) (PromptRepository, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("sqlite path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}

	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS llm_exchanges (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at TEXT NOT NULL,
	user_prompt TEXT NOT NULL,
	assistant_json TEXT NOT NULL
);`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite schema: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// SaveLLMExchange appends one user/assistant exchange.
func (r *SQLiteRepository) SaveLLMExchange(ctx context.Context, userPrompt, assistantJSON string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO llm_exchanges (created_at, user_prompt, assistant_json) VALUES (?, ?, ?)`,
		time.Now().UTC().Format(time.RFC3339Nano),
		userPrompt,
		assistantJSON,
	)
	if err != nil {
		return fmt.Errorf("insert llm exchange: %w", err)
	}
	return nil
}

// Close releases the database handle.
func (r *SQLiteRepository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}
