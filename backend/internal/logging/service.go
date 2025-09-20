package logging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
)

type Service struct {
	repo        *Repository
	settingsSvc *settings.Service
	ctx         context.Context
	config      LogConfig
}

func NewService(settingsSvc *settings.Service) *Service {
	s := &Service{
		repo:        NewRepository(),
		settingsSvc: settingsSvc,
		config:      DefaultLogConfig(),
	}

	// Load logging configuration from settings
	s.loadConfig()

	return s
}

func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// loadConfig loads logging configuration from settings
func (s *Service) loadConfig() {
	// For now, use default config - we'll add settings integration later
	s.config = DefaultLogConfig()

	// You can extend this to load from settings when expanded
}

// Log creates a new log entry if logging is enabled for the module and level
func (s *Service) Log(module LogModule, level LogLevel, message string, metadata interface{}) error {
	if !s.shouldLog(module, level) {
		return nil
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Module:    module,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now(),
	}

	if metadata != nil {
		if err := entry.SetMetadata(metadata); err != nil {
			log.Printf("Failed to set metadata for log entry: %v", err)
		}
	}

	if err := s.repo.CreateLogEntry(entry); err != nil {
		log.Printf("Failed to create log entry: %v", err)
		return err
	}

	// Emit real-time update if enabled
	if s.config.RealTimeUpdates && s.ctx != nil {
		// We'll add event emission later
	}

	return nil
}

// LogItemFound logs when a new item is found during live search
func (s *Service) LogItemFound(searchID, itemID, itemName, league, searchURL string, price *Price, itemDetails string) error {
	metadata := ItemFoundMetadata{
		SearchID:    searchID,
		ItemID:      itemID,
		ItemName:    itemName,
		Price:       price,
		League:      league,
		SearchURL:   searchURL,
		ItemDetails: itemDetails,
	}

	message := fmt.Sprintf("New item found: %s", itemName)
	if price != nil {
		message += fmt.Sprintf(" - %.2f %s", price.Amount, price.Currency)
	}

	return s.Log(LogModuleLiveSearch, LogLevelSuccess, message, metadata)
}

// LogAPICall logs API requests and responses
func (s *Service) LogAPICall(url, method string, statusCode int, responseTime time.Duration, errorMessage string) error {
	metadata := APICallMetadata{
		URL:          url,
		Method:       method,
		StatusCode:   statusCode,
		ResponseTime: responseTime.Milliseconds(),
		ErrorMessage: errorMessage,
	}

	level := LogLevelInfo
	message := fmt.Sprintf("API %s %s - %d", method, url, statusCode)

	if statusCode >= 400 {
		level = LogLevelError
		message += fmt.Sprintf(" - %s", errorMessage)
	} else if responseTime > 5*time.Second {
		level = LogLevelWarning
		message += " - Slow response"
	}

	return s.Log(LogModuleAPI, level, message, metadata)
}

// LogWebSocketEvent logs WebSocket events
func (s *Service) LogWebSocketEvent(searchID, eventType string, messageCount int, errorMessage string) error {
	metadata := WebSocketMetadata{
		SearchID:     searchID,
		EventType:    eventType,
		MessageCount: messageCount,
		ErrorMessage: errorMessage,
	}

	level := LogLevelInfo
	message := fmt.Sprintf("WebSocket %s for %s", eventType, searchID)

	if errorMessage != "" {
		level = LogLevelError
		message += fmt.Sprintf(" - %s", errorMessage)
	} else if messageCount > 0 {
		message += fmt.Sprintf(" - %d messages", messageCount)
	}

	return s.Log(LogModuleWebSocket, level, message, metadata)
}

// Debug logs a debug message
func (s *Service) Debug(module LogModule, message string, metadata interface{}) error {
	return s.Log(module, LogLevelDebug, message, metadata)
}

// Info logs an info message
func (s *Service) Info(module LogModule, message string, metadata interface{}) error {
	return s.Log(module, LogLevelInfo, message, metadata)
}

// Warning logs a warning message
func (s *Service) Warning(module LogModule, message string, metadata interface{}) error {
	return s.Log(module, LogLevelWarning, message, metadata)
}

// Error logs an error message
func (s *Service) Error(module LogModule, message string, metadata interface{}) error {
	return s.Log(module, LogLevelError, message, metadata)
}

// Success logs a success message
func (s *Service) Success(module LogModule, message string, metadata interface{}) error {
	return s.Log(module, LogLevelSuccess, message, metadata)
}

// GetLogEntries retrieves log entries with filtering
func (s *Service) GetLogEntries(filter LogFilter) ([]LogEntry, error) {
	// Set default limit if not specified
	if filter.Limit <= 0 {
		filter.Limit = 100
	}

	return s.repo.GetLogEntries(filter)
}

// GetLogEntriesCount returns the total count of log entries
func (s *Service) GetLogEntriesCount(filter LogFilter) (int64, error) {
	return s.repo.GetLogEntriesCount(filter)
}

// GetLogModules returns available log modules
func (s *Service) GetLogModules() ([]LogModule, error) {
	return s.repo.GetLogModules()
}

// GetLogLevels returns available log levels
func (s *Service) GetLogLevels() ([]LogLevel, error) {
	return s.repo.GetLogLevels()
}

// CleanupOldLogs removes old log entries based on configuration
func (s *Service) CleanupOldLogs() error {
	if s.config.RetentionDays > 0 {
		deleted, err := s.repo.DeleteOldLogEntries(s.config.RetentionDays)
		if err != nil {
			return err
		}
		if deleted > 0 {
			s.Info(LogModuleSystem, fmt.Sprintf("Cleaned up %d old log entries", deleted), nil)
		}
	}

	if s.config.MaxEntries > 0 {
		deleted, err := s.repo.DeleteExcessLogEntries(s.config.MaxEntries)
		if err != nil {
			return err
		}
		if deleted > 0 {
			s.Info(LogModuleSystem, fmt.Sprintf("Cleaned up %d excess log entries", deleted), nil)
		}
	}

	return nil
}

// GetConfig returns the current logging configuration
func (s *Service) GetConfig() LogConfig {
	return s.config
}

// UpdateConfig updates the logging configuration
func (s *Service) UpdateConfig(config LogConfig) error {
	s.config = config
	// TODO: Save to settings when settings support logging config
	return nil
}

// shouldLog determines if a log entry should be created based on configuration
func (s *Service) shouldLog(module LogModule, level LogLevel) bool {
	if !s.config.Enabled {
		return false
	}

	// Check if module is enabled
	moduleEnabled := false
	for _, enabledModule := range s.config.LogModules {
		if enabledModule == module {
			moduleEnabled = true
			break
		}
	}
	if !moduleEnabled {
		return false
	}

	// Check log level
	levelPriority := map[LogLevel]int{
		LogLevelDebug:   1,
		LogLevelInfo:    2,
		LogLevelWarning: 3,
		LogLevelError:   4,
		LogLevelSuccess: 2,
	}

	currentPriority := levelPriority[s.config.LogLevel]
	logPriority := levelPriority[level]

	return logPriority >= currentPriority
}
