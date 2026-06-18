package logging

import (
	"context"
	"time"
)

type Handler struct {
	svc           *Service
	windowManager WindowManager
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) SetWindowManager(wm WindowManager) {
	h.windowManager = wm
}

func (h *Handler) SetContext(ctx context.Context) {
	h.svc.SetContext(ctx)
}

// OpenLogsWindow opens the dedicated logs window via the window manager
func (h *Handler) OpenLogsWindow() error {
	if h.windowManager != nil {
		return h.windowManager.OpenLogsWindow(context.Background())
	}
	return nil
}

// GetLogEntries retrieves log entries with filtering and pagination
func (h *Handler) GetLogEntries(filter LogFilter) ([]LogEntry, error) {
	return h.svc.GetLogEntries(filter)
}

// GetLogEntriesCount returns the total count of log entries
func (h *Handler) GetLogEntriesCount(filter LogFilter) (int64, error) {
	return h.svc.GetLogEntriesCount(filter)
}

// GetLogModules returns available log modules
func (h *Handler) GetLogModules() ([]string, error) {
	modules, err := h.svc.GetLogModules()
	if err != nil {
		return nil, err
	}

	// Convert to strings for frontend
	result := make([]string, len(modules))
	for i, module := range modules {
		result[i] = string(module)
	}

	return result, nil
}

// GetLogLevels returns available log levels
func (h *Handler) GetLogLevels() ([]string, error) {
	levels, err := h.svc.GetLogLevels()
	if err != nil {
		return nil, err
	}

	// Convert to strings for frontend
	result := make([]string, len(levels))
	for i, level := range levels {
		result[i] = string(level)
	}

	return result, nil
}

// GetRecentLogs returns the most recent log entries (last 100)
func (h *Handler) GetRecentLogs() ([]LogEntry, error) {
	filter := LogFilter{
		Limit: 100,
	}
	return h.svc.GetLogEntries(filter)
}

// GetLogsByModule returns log entries for a specific module
func (h *Handler) GetLogsByModule(module string, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	logModule := LogModule(module)
	filter := LogFilter{
		Module: &logModule,
		Limit:  limit,
	}

	return h.svc.GetLogEntries(filter)
}

// GetLogsByLevel returns log entries for a specific level
func (h *Handler) GetLogsByLevel(level string, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	logLevel := LogLevel(level)
	filter := LogFilter{
		Level: &logLevel,
		Limit: limit,
	}

	return h.svc.GetLogEntries(filter)
}

// GetLogsByDateRange returns log entries within a date range
func (h *Handler) GetLogsByDateRange(startTime, endTime time.Time, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	filter := LogFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     limit,
	}

	return h.svc.GetLogEntries(filter)
}

// SearchLogs searches log entries by message content
func (h *Handler) SearchLogs(searchText string, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	filter := LogFilter{
		Search: searchText,
		Limit:  limit,
	}

	return h.svc.GetLogEntries(filter)
}

// ClearLogs removes all log entries (use with caution)
func (h *Handler) ClearLogs() error {
	// For now, we'll delete logs older than 0 days (all logs)
	_, err := h.svc.repo.DeleteOldLogEntries(0)
	if err != nil {
		return err
	}

	h.svc.Info(LogModuleSystem, "All logs cleared by user", nil)
	return nil
}

// CleanupOldLogs removes old log entries based on retention policy
func (h *Handler) CleanupOldLogs() error {
	return h.svc.CleanupOldLogs()
}

// GetLogConfig returns the current logging configuration
func (h *Handler) GetLogConfig() LogConfig {
	return h.svc.GetConfig()
}

// UpdateLogConfig updates the logging configuration
func (h *Handler) UpdateLogConfig(config LogConfig) error {
	return h.svc.UpdateConfig(config)
}

// GetLogStats returns statistics about log entries
func (h *Handler) GetLogStats() (map[string]interface{}, error) {
	// Get total count
	totalCount, err := h.svc.GetLogEntriesCount(LogFilter{})
	if err != nil {
		return nil, err
	}

	// Get count by module
	modules, err := h.svc.GetLogModules()
	if err != nil {
		return nil, err
	}

	moduleStats := make(map[string]int64)
	for _, module := range modules {
		count, err := h.svc.GetLogEntriesCount(LogFilter{Module: &module})
		if err != nil {
			continue
		}
		moduleStats[string(module)] = count
	}

	// Get count by level
	levels, err := h.svc.GetLogLevels()
	if err != nil {
		return nil, err
	}

	levelStats := make(map[string]int64)
	for _, level := range levels {
		count, err := h.svc.GetLogEntriesCount(LogFilter{Level: &level})
		if err != nil {
			continue
		}
		levelStats[string(level)] = count
	}

	// Get today's count
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	todayCount, err := h.svc.GetLogEntriesCount(LogFilter{
		StartTime: &today,
		EndTime:   &tomorrow,
	})
	if err != nil {
		todayCount = 0
	}

	return map[string]interface{}{
		"total_entries": totalCount,
		"today_entries": todayCount,
		"by_module":     moduleStats,
		"by_level":      levelStats,
		"config":        h.svc.GetConfig(),
	}, nil
}
