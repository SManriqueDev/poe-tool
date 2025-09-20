package settings

import (
	"database/sql"

	"github.com/SManriqueDev/poe-tool/backend/db"
)

type Setting struct {
	ID    int
	Key   string
	Value string
}

type Repository struct {
	db *sql.DB
}

// RepositoryOption defines a functional option for Repository configuration
type RepositoryOption func(*Repository)

// WithDatabase sets a custom database connection
func WithDatabase(database *sql.DB) RepositoryOption {
	return func(r *Repository) {
		r.db = database
	}
}

func NewRepository(opts ...RepositoryOption) *Repository {
	r := &Repository{db: db.GetDB()}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Repository) Set(key, value string) error {
	_, err := r.db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}

func (r *Repository) Get(key string) (string, error) {
	var value string
	err := r.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	return value, err
}
