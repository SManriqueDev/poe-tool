package adapters

import (
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// DomainConfig centraliza la configuración para los componentes domain-pure
type DomainConfig struct {
	WebSocket WebSocketConfig `json:"websocket"`
	EventBus  EventBusConfig  `json:"eventbus"`
	APIClient APIClientConfig `json:"api_client"`
	Logger    LoggerConfig    `json:"logger"`
}

// WebSocketConfig configuración para el cliente WebSocket
type WebSocketConfig struct {
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	PingInterval     time.Duration `json:"ping_interval"`
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	MessageBuffer    int           `json:"message_buffer"`
	ReconnectEnabled bool          `json:"reconnect_enabled"`
}

// EventBusConfig configuración para el bus de eventos
type EventBusConfig struct {
	MaxListeners    int  `json:"max_listeners"`
	AsyncEmit       bool `json:"async_emit"`
	EnableMetrics   bool `json:"enable_metrics"`
	EnableDebugLogs bool `json:"enable_debug_logs"`
}

// APIClientConfig configuración para el cliente API
type APIClientConfig struct {
	BaseURL        string        `json:"base_url"`
	UserAgent      string        `json:"user_agent"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	RateLimitDelay time.Duration `json:"rate_limit_delay"`
	EnableLogging  bool          `json:"enable_logging"`
}

// LoggerConfig configuración para el adaptador de logging
type LoggerConfig struct {
	DefaultModule string `json:"default_module"`
	EnableDebug   bool   `json:"enable_debug"`
	EnableInfo    bool   `json:"enable_info"`
	EnableError   bool   `json:"enable_error"`
	EnableWarning bool   `json:"enable_warning"`
}

// DefaultDomainConfig retorna la configuración por defecto para componentes domain-pure
func DefaultDomainConfig() *DomainConfig {
	return &DomainConfig{
		WebSocket: WebSocketConfig{
			MaxRetries:       3,
			RetryDelay:       5 * time.Second,
			PingInterval:     30 * time.Second,
			ReadTimeout:      60 * time.Second,
			WriteTimeout:     10 * time.Second,
			MessageBuffer:    100,
			ReconnectEnabled: true,
		},
		EventBus: EventBusConfig{
			MaxListeners:    10,
			AsyncEmit:       true,
			EnableMetrics:   true,
			EnableDebugLogs: false,
		},
		APIClient: APIClientConfig{
			BaseURL:        "https://www.pathofexile.com/api",
			UserAgent:      "PoeTool/1.0",
			Timeout:        30 * time.Second,
			MaxRetries:     3,
			RetryDelay:     2 * time.Second,
			RateLimitDelay: 100 * time.Millisecond,
			EnableLogging:  true,
		},
		Logger: LoggerConfig{
			DefaultModule: "domain",
			EnableDebug:   true,
			EnableInfo:    true,
			EnableError:   true,
			EnableWarning: true,
		},
	}
}

// DomainComponentsFactory crea instancias configuradas de los componentes domain-pure
type DomainComponentsFactory struct {
	config *DomainConfig
	logger domain.Logger
}

// NewDomainComponentsFactory crea una nueva factory con configuración
func NewDomainComponentsFactory(config *DomainConfig, logger domain.Logger) *DomainComponentsFactory {
	if config == nil {
		config = DefaultDomainConfig()
	}
	return &DomainComponentsFactory{
		config: config,
		logger: logger,
	}
}

// CreateWebSocketClient crea un WebSocket client configurado
func (f *DomainComponentsFactory) CreateWebSocketClient() domain.WebSocketClient {
	client := NewDomainWebSocketClient(f.logger)

	// Aplicar configuración específica
	client.maxRetries = f.config.WebSocket.MaxRetries
	client.retryDelay = f.config.WebSocket.RetryDelay
	client.pingInterval = f.config.WebSocket.PingInterval
	client.readTimeout = f.config.WebSocket.ReadTimeout
	client.writeTimeout = f.config.WebSocket.WriteTimeout

	// Recrear canal con buffer configurado
	if f.config.WebSocket.MessageBuffer != 100 {
		client.messageChannel = make(chan domain.ItemResult, f.config.WebSocket.MessageBuffer)
	}

	f.logger.Info("domain", "WebSocket client created with configuration", map[string]interface{}{
		"max_retries":    f.config.WebSocket.MaxRetries,
		"retry_delay":    f.config.WebSocket.RetryDelay,
		"ping_interval":  f.config.WebSocket.PingInterval,
		"message_buffer": f.config.WebSocket.MessageBuffer,
	})

	return client
}

// CreateEventBus crea un EventBus configurado
func (f *DomainComponentsFactory) CreateEventBus() *DomainEventBus {
	eventBus := NewDomainEventBus(f.logger)

	f.logger.Info("domain", "EventBus created with configuration", map[string]interface{}{
		"max_listeners":     f.config.EventBus.MaxListeners,
		"async_emit":        f.config.EventBus.AsyncEmit,
		"enable_metrics":    f.config.EventBus.EnableMetrics,
		"enable_debug_logs": f.config.EventBus.EnableDebugLogs,
	})

	return eventBus
}

// CreateAPIClient crea un API client configurado
func (f *DomainComponentsFactory) CreateAPIClient() *DomainAPIClient {
	client := NewDomainAPIClient(f.logger)

	// Aplicar configuración específica
	client.baseURL = f.config.APIClient.BaseURL
	client.userAgent = f.config.APIClient.UserAgent
	client.timeout = f.config.APIClient.Timeout
	client.maxRetries = f.config.APIClient.MaxRetries
	client.retryDelay = f.config.APIClient.RetryDelay
	client.rateLimitDelay = f.config.APIClient.RateLimitDelay

	// Actualizar HTTP client con nuevo timeout
	client.SetTimeout(f.config.APIClient.Timeout)

	f.logger.Info("domain", "API client created with configuration", map[string]interface{}{
		"base_url":         f.config.APIClient.BaseURL,
		"timeout":          f.config.APIClient.Timeout,
		"max_retries":      f.config.APIClient.MaxRetries,
		"rate_limit_delay": f.config.APIClient.RateLimitDelay,
	})

	return client
}

// GetConfig retorna la configuración actual
func (f *DomainComponentsFactory) GetConfig() *DomainConfig {
	return f.config
}

// UpdateConfig actualiza la configuración de la factory
func (f *DomainComponentsFactory) UpdateConfig(config *DomainConfig) {
	f.config = config
	f.logger.Info("domain", "Factory configuration updated", map[string]interface{}{
		"websocket_retries": config.WebSocket.MaxRetries,
		"api_timeout":       config.APIClient.Timeout,
		"eventbus_metrics":  config.EventBus.EnableMetrics,
	})
}

// Validate valida la configuración
func (c *DomainConfig) Validate() error {
	// Validaciones básicas
	if c.WebSocket.MaxRetries < 0 {
		return domain.ErrInvalidSettingValue
	}
	if c.WebSocket.RetryDelay <= 0 {
		return domain.ErrInvalidSettingValue
	}
	if c.APIClient.Timeout <= 0 {
		return domain.ErrInvalidSettingValue
	}
	if c.EventBus.MaxListeners <= 0 {
		return domain.ErrInvalidSettingValue
	}

	return nil
}
