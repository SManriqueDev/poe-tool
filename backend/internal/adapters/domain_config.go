package adapters

import (
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// DomainConfig centraliza la configuración para los componentes domain-pure
type DomainConfig struct {
	WebSocket         WebSocketConfig         `json:"websocket"`
	EventBus          EventBusConfig          `json:"eventbus"`
	APIClient         APIClientConfig         `json:"api_client"`
	Logger            LoggerConfig            `json:"logger"`
	WindowManager     WindowManagerConfig     `json:"window_manager"`
	HideoutAutomation HideoutAutomationConfig `json:"hideout_automation"`
	SystemAPIClient   SystemAPIClientConfig   `json:"system_api_client"`
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

// WindowManagerConfig configuración para el window manager
type WindowManagerConfig struct {
	DefaultWidth  int  `json:"default_width"`
	DefaultHeight int  `json:"default_height"`
	EnableLogs    bool `json:"enable_logs"`
}

// HideoutAutomationConfig configuración para automatización de hideout
type HideoutAutomationConfig struct {
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
	ProcessDelay time.Duration `json:"process_delay"`
	MaxQueueSize int           `json:"max_queue_size"`
	EnableQueue  bool          `json:"enable_queue"`
	AutoStart    bool          `json:"auto_start"`
}

// SystemAPIClientConfig configuración para el cliente de APIs del sistema
type SystemAPIClientConfig struct {
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	RateLimitDelay time.Duration `json:"rate_limit_delay"`
	UserAgent      string        `json:"user_agent"`
	EnableLogging  bool          `json:"enable_logging"`
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
		WindowManager: WindowManagerConfig{
			DefaultWidth:  800,
			DefaultHeight: 600,
			EnableLogs:    true,
		},
		HideoutAutomation: HideoutAutomationConfig{
			MaxRetries:   3,
			RetryDelay:   2 * time.Second,
			ProcessDelay: 8 * time.Second, // Realistic trading time between visits
			MaxQueueSize: 50,
			EnableQueue:  true,
			AutoStart:    true,
		},
		SystemAPIClient: SystemAPIClientConfig{
			Timeout:        30 * time.Second,
			MaxRetries:     3,
			RetryDelay:     2 * time.Second,
			RateLimitDelay: 500 * time.Millisecond, // Conservative for PoE API
			UserAgent:      "PoeTool/1.0",
			EnableLogging:  true,
		},
	}
}

// DomainComponentsFactory crea componentes domain-pure
type DomainComponentsFactory struct {
	config *DomainConfig
	logger domain.Logger
}

// NewDomainComponentsFactory crea una nueva factory
func NewDomainComponentsFactory(config *DomainConfig, logger domain.Logger) *DomainComponentsFactory {
	return &DomainComponentsFactory{
		config: config,
		logger: logger,
	}
}

// CreateEventBus crea el bus de eventos del dominio
func (f *DomainComponentsFactory) CreateEventBus() domain.EventBus {
	eventBus := NewDomainEventBus(f.logger)

	// Registrar el listener que conecta a Wails
	wailsListener := NewWailsEventListener()
	eventBus.RegisterNewItemListener(wailsListener)
	eventBus.RegisterLinkStatusListener(wailsListener)
	eventBus.RegisterLiveSearchStatusListener(wailsListener)

	f.logger.Info("domain", "Event bus created with Wails integration", nil)

	return eventBus
}

// CreateWebSocketClient crea un cliente WebSocket configurado
func (f *DomainComponentsFactory) CreateWebSocketClient() domain.WebSocketClient {
	client := NewDomainWebSocketClient(f.logger)

	// Aplicar configuración específica
	client.maxRetries = f.config.WebSocket.MaxRetries
	client.retryDelay = f.config.WebSocket.RetryDelay
	client.pingInterval = f.config.WebSocket.PingInterval
	client.readTimeout = f.config.WebSocket.ReadTimeout
	client.writeTimeout = f.config.WebSocket.WriteTimeout

	f.logger.Info("domain", "WebSocket client created with configuration", map[string]interface{}{
		"max_retries":   f.config.WebSocket.MaxRetries,
		"retry_delay":   f.config.WebSocket.RetryDelay,
		"ping_interval": f.config.WebSocket.PingInterval,
	})

	return client
}

// CreateSystemAPIClient crea un System API Client configurado
func (f *DomainComponentsFactory) CreateSystemAPIClient() domain.SystemAPIClient {
	client := NewDomainSystemAPIClient(f.logger)

	// Aplicar configuración específica
	client.SetTimeout(f.config.SystemAPIClient.Timeout)
	client.SetRateLimit(f.config.SystemAPIClient.RateLimitDelay)
	client.maxRetries = f.config.SystemAPIClient.MaxRetries
	client.retryDelay = f.config.SystemAPIClient.RetryDelay
	client.userAgent = f.config.SystemAPIClient.UserAgent

	f.logger.Info("domain", "System API client created with configuration", map[string]interface{}{
		"timeout":          f.config.SystemAPIClient.Timeout,
		"max_retries":      f.config.SystemAPIClient.MaxRetries,
		"retry_delay":      f.config.SystemAPIClient.RetryDelay,
		"rate_limit_delay": f.config.SystemAPIClient.RateLimitDelay,
		"user_agent":       f.config.SystemAPIClient.UserAgent,
		"enable_logging":   f.config.SystemAPIClient.EnableLogging,
	})

	return client
}

// CreateWindowManager crea un Window Manager configurado
func (f *DomainComponentsFactory) CreateWindowManager() domain.WindowManager {
	manager := NewDomainWindowManager(f.logger)

	// Aplicar configuración específica
	manager.SetDefaultSize(f.config.WindowManager.DefaultWidth, f.config.WindowManager.DefaultHeight)

	f.logger.Info("domain", "Window manager created with configuration", map[string]interface{}{
		"default_width":  f.config.WindowManager.DefaultWidth,
		"default_height": f.config.WindowManager.DefaultHeight,
		"enable_logs":    f.config.WindowManager.EnableLogs,
	})

	return manager
}

// CreateHideoutAutomation crea un Hideout Automation configurado
func (f *DomainComponentsFactory) CreateHideoutAutomation(systemAPIClient domain.SystemAPIClient, settingsRepo domain.LiveSearchRepository) domain.HideoutAutomation {
	automation := NewDomainHideoutAutomation(f.logger, systemAPIClient, settingsRepo)

	// Aplicar configuración específica
	automation.SetConfiguration(
		f.config.HideoutAutomation.MaxRetries,
		f.config.HideoutAutomation.RetryDelay,
		f.config.HideoutAutomation.ProcessDelay,
		f.config.HideoutAutomation.MaxQueueSize,
	)

	f.logger.Info("domain", "Hideout automation created with configuration", map[string]interface{}{
		"max_retries":    f.config.HideoutAutomation.MaxRetries,
		"retry_delay":    f.config.HideoutAutomation.RetryDelay,
		"process_delay":  f.config.HideoutAutomation.ProcessDelay,
		"max_queue_size": f.config.HideoutAutomation.MaxQueueSize,
		"enable_queue":   f.config.HideoutAutomation.EnableQueue,
		"auto_start":     f.config.HideoutAutomation.AutoStart,
	})

	return automation
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
