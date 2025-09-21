package logging

import (
	"database/sql"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/db"
)

type Repository struct {
	db *sql.DB
}

func NewRepository() *Repository {
	return &Repository{
		db: db.GetDB(),
	}
}

// CreateLogEntry creates a new log entry
func (r *Repository) CreateLogEntry(entry LogEntry) error {
	query := `
		INSERT INTO logs (timestamp, module, level, message, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(query, entry.Timestamp, entry.Module, entry.Level, entry.Message, entry.Metadata, entry.CreatedAt)
	return err
}

// CreateLogEntryAndReturn creates a new log entry and returns it with the generated ID
func (r *Repository) CreateLogEntryAndReturn(entry LogEntry) (*LogEntry, error) {
	query := `
		INSERT INTO logs (timestamp, module, level, message, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	result, err := r.db.Exec(query, entry.Timestamp, entry.Module, entry.Level, entry.Message, entry.Metadata, entry.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Get the ID of the created entry
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Set the ID and return the entry
	entry.ID = id
	return &entry, nil
}

// GetLogEntries retrieves log entries based on filter
func (r *Repository) GetLogEntries(filter LogFilter) ([]LogEntry, error) {
	query := `
		SELECT id, timestamp, module, level, message, metadata, created_at
		FROM logs
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply filters
	if filter.Module != nil {
		query += " AND module = ?"
		args = append(args, *filter.Module)
	}

	if filter.Level != nil {
		query += " AND level = ?"
		args = append(args, *filter.Level)
	}

	if filter.StartTime != nil {
		query += " AND timestamp >= ?"
		args = append(args, *filter.StartTime)
	}

	if filter.EndTime != nil {
		query += " AND timestamp <= ?"
		args = append(args, *filter.EndTime)
	}

	if filter.Search != "" {
		query += " AND (message LIKE ? OR metadata LIKE ?)"
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Order by timestamp desc (newest first)
	query += " ORDER BY timestamp DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.Module, &entry.Level, &entry.Message, &entry.Metadata, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetLogEntriesCount returns the total count of log entries matching the filter
func (r *Repository) GetLogEntriesCount(filter LogFilter) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM logs
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply same filters as GetLogEntries (excluding limit/offset)
	if filter.Module != nil {
		query += " AND module = ?"
		args = append(args, *filter.Module)
	}

	if filter.Level != nil {
		query += " AND level = ?"
		args = append(args, *filter.Level)
	}

	if filter.StartTime != nil {
		query += " AND timestamp >= ?"
		args = append(args, *filter.StartTime)
	}

	if filter.EndTime != nil {
		query += " AND timestamp <= ?"
		args = append(args, *filter.EndTime)
	}

	if filter.Search != "" {
		query += " AND (message LIKE ? OR metadata LIKE ?)"
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	var count int64
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// DeleteOldLogEntries deletes log entries older than the specified number of days
func (r *Repository) DeleteOldLogEntries(retentionDays int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result, err := r.db.Exec("DELETE FROM logs WHERE timestamp < ?", cutoffTime)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// DeleteExcessLogEntries deletes the oldest log entries when count exceeds maxEntries
func (r *Repository) DeleteExcessLogEntries(maxEntries int) (int64, error) {
	// First, get the current count
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
	if err != nil {
		return 0, err
	}

	if count <= int64(maxEntries) {
		return 0, nil // No need to delete
	}

	// Delete the oldest entries
	entriesToDelete := count - int64(maxEntries)
	result, err := r.db.Exec(`
		DELETE FROM logs
		WHERE id IN (
			SELECT id FROM logs
			ORDER BY timestamp ASC
			LIMIT ?
		)
	`, entriesToDelete)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetLogModules returns all distinct modules that have log entries
func (r *Repository) GetLogModules() ([]LogModule, error) {
	rows, err := r.db.Query("SELECT DISTINCT module FROM logs ORDER BY module")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []LogModule
	for rows.Next() {
		var module LogModule
		if err := rows.Scan(&module); err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}

	return modules, nil
}

// GetLogLevels returns all distinct levels that have log entries
func (r *Repository) GetLogLevels() ([]LogLevel, error) {
	rows, err := r.db.Query("SELECT DISTINCT level FROM logs ORDER BY level")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []LogLevel
	for rows.Next() {
		var level LogLevel
		if err := rows.Scan(&level); err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}

	return levels, nil
}
