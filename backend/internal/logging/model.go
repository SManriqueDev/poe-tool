package logging

import (
	"encoding/json"
	"time"
)

// LogLevel represents the severity of the log entry
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelSuccess LogLevel = "success"
)

// LogModule represents which part of the application generated the log
type LogModule string

const (
	LogModuleLiveSearch LogModule = "livesearch"
	LogModuleSettings   LogModule = "settings"
	LogModuleWebSocket  LogModule = "websocket"
	LogModuleAPI        LogModule = "api"
	LogModuleSystem     LogModule = "system"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID        int64     `json:"id" db:"id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Module    LogModule `json:"module" db:"module"`
	Level     LogLevel  `json:"level" db:"level"`
	Message   string    `json:"message" db:"message"`
	Metadata  string    `json:"metadata" db:"metadata"` // JSON string for flexible data
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ItemFoundMetadata represents metadata for when new items are found
type ItemFoundMetadata struct {
	SearchID    string `json:"search_id"`
	ItemID      string `json:"item_id"`
	ItemName    string `json:"item_name"`
	Price       *Price `json:"price,omitempty"`
	League      string `json:"league"`
	SearchURL   string `json:"search_url"`
	ItemDetails string `json:"item_details,omitempty"`
}

// Price represents item pricing information
type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Type     string  `json:"type"` // "exact", "negotiable", etc.
}

// APICallMetadata represents metadata for API calls
type APICallMetadata struct {
	URL          string `json:"url"`
	Method       string `json:"method"`
	StatusCode   int    `json:"status_code"`
	ResponseTime int64  `json:"response_time_ms"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// WebSocketMetadata represents metadata for WebSocket events
type WebSocketMetadata struct {
	SearchID     string `json:"search_id"`
	EventType    string `json:"event_type"`
	MessageCount int    `json:"message_count,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// SetMetadata converts a metadata struct to JSON and sets it
func (l *LogEntry) SetMetadata(metadata interface{}) error {
	if metadata == nil {
		l.Metadata = ""
		return nil
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	l.Metadata = string(jsonData)
	return nil
}

// GetMetadata unmarshals the metadata JSON into the provided interface
func (l *LogEntry) GetMetadata(v interface{}) error {
	if l.Metadata == "" {
		return nil
	}
	return json.Unmarshal([]byte(l.Metadata), v)
}

// LogFilter represents filters for querying logs
type LogFilter struct {
	Module    *LogModule `json:"module,omitempty"`
	Level     *LogLevel  `json:"level,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Search    string     `json:"search,omitempty"` // Search in message
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Enabled         bool        `json:"enabled"`
	LogLevel        LogLevel    `json:"log_level"`
	LogModules      []LogModule `json:"log_modules"`
	LogNewItems     bool        `json:"log_new_items"`
	LogAPIRequests  bool        `json:"log_api_requests"`
	LogWebSocket    bool        `json:"log_websocket"`
	RetentionDays   int         `json:"retention_days"`
	MaxEntries      int         `json:"max_entries"`
	RealTimeUpdates bool        `json:"real_time_updates"`
}

// DefaultLogConfig returns the default logging configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Enabled:         true,
		LogLevel:        LogLevelInfo,
		LogModules:      []LogModule{LogModuleLiveSearch, LogModuleSettings, LogModuleWebSocket, LogModuleAPI},
		LogNewItems:     true,
		LogAPIRequests:  true,
		LogWebSocket:    true,
		RetentionDays:   30,
		MaxEntries:      10000,
		RealTimeUpdates: true,
	}
}
