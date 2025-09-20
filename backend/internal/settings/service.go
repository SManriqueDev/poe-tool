package settings

import (
	"strconv"
)

type Service struct {
	repo          *Repository
	data          *Config
	defaultConfig *Config
}

// ServiceOption defines a functional option for Service configuration
type ServiceOption func(*ServiceOptions)

// ServiceOptions holds configuration options for the Service
type ServiceOptions struct {
	repository    *Repository
	defaultConfig *Config
	skipLoad      bool
}

// WithRepository sets a custom repository
func WithRepository(repo *Repository) ServiceOption {
	return func(opts *ServiceOptions) {
		opts.repository = repo
	}
}

// WithDefaultConfig sets custom default configuration values
func WithDefaultConfig(config *Config) ServiceOption {
	return func(opts *ServiceOptions) {
		opts.defaultConfig = config
	}
}

// WithSkipLoad skips the initial load from database (useful for testing)
func WithSkipLoad(skip bool) ServiceOption {
	return func(opts *ServiceOptions) {
		opts.skipLoad = skip
	}
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	return &Config{
		PoeSessid:         "",
		AccountName:       "",
		League:            "Standard",
		AutomationEnabled: false,
		Delay:             1000,
	}
}

// NewService creates a new Settings service with optional configuration.
//
// Usage examples:
//   - Basic: NewService("MyApp")
//   - With custom defaults: NewService("MyApp", WithDefaultConfig(customConfig))
//   - For testing: NewService("MyApp", WithSkipLoad(true), WithRepository(mockRepo))
func NewService(appName string, opts ...ServiceOption) (*Service, error) {
	// Apply default options
	options := &ServiceOptions{
		repository:    NewRepository(),
		defaultConfig: getDefaultConfig(),
		skipLoad:      false,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	s := &Service{
		repo:          options.repository,
		defaultConfig: options.defaultConfig,
	}

	if !options.skipLoad {
		if err := s.load(); err != nil {
			// initialize with default values
			s.data = options.defaultConfig
			_ = s.Save()
		}
	} else {
		// Use default config when skipping load
		s.data = options.defaultConfig
	}

	return s, nil
}

func (s *Service) load() error {
	config := &Config{}

	// Load each setting from database
	if poeSessid, err := s.repo.Get("poesessid"); err == nil {
		config.PoeSessid = poeSessid
	}

	if accountName, err := s.repo.Get("accountName"); err == nil {
		config.AccountName = accountName
	}

	if league, err := s.repo.Get("league"); err == nil {
		config.League = league
	} else {
		config.League = "Standard"
	}

	if automationEnabledStr, err := s.repo.Get("automationEnabled"); err == nil {
		config.AutomationEnabled = automationEnabledStr == "true"
	}

	if delayStr, err := s.repo.Get("delay"); err == nil {
		if delay, parseErr := strconv.Atoi(delayStr); parseErr == nil {
			config.Delay = delay
		} else {
			config.Delay = 1000
		}
	} else {
		config.Delay = 1000
	}

	s.data = config
	return nil
}

func (s *Service) Save() error {
	// Save each setting to database
	if err := s.repo.Set("poesessid", s.data.PoeSessid); err != nil {
		return err
	}

	if err := s.repo.Set("accountName", s.data.AccountName); err != nil {
		return err
	}

	if err := s.repo.Set("league", s.data.League); err != nil {
		return err
	}

	automationEnabledStr := "false"
	if s.data.AutomationEnabled {
		automationEnabledStr = "true"
	}
	if err := s.repo.Set("automationEnabled", automationEnabledStr); err != nil {
		return err
	}

	if err := s.repo.Set("delay", strconv.Itoa(s.data.Delay)); err != nil {
		return err
	}

	return nil
}

func (s *Service) Get() *Config {
	return s.data
}

func (s *Service) Update(newData Config) error {
	s.data = &newData
	return s.Save()
}

// Reset resets the configuration to default values
func (s *Service) Reset() error {
	s.data = s.defaultConfig
	return s.Save()
}

// UpdateField updates a specific field in the configuration
func (s *Service) UpdateField(field string, value interface{}) error {
	switch field {
	case "poesessid":
		if v, ok := value.(string); ok {
			s.data.PoeSessid = v
		}
	case "accountName":
		if v, ok := value.(string); ok {
			s.data.AccountName = v
		}
	case "league":
		if v, ok := value.(string); ok {
			s.data.League = v
		}
	case "automationEnabled":
		if v, ok := value.(bool); ok {
			s.data.AutomationEnabled = v
		}
	case "delay":
		if v, ok := value.(int); ok {
			s.data.Delay = v
		}
	default:
		return nil // silently ignore unknown fields
	}
	return s.Save()
}

// Reload reloads the configuration from the database
func (s *Service) Reload() error {
	return s.load()
}

// NewTestService creates a service configured for testing (skips database loading)
func NewTestService(customDefaults ...*Config) *Service {
	var defaultConfig *Config
	if len(customDefaults) > 0 {
		defaultConfig = customDefaults[0]
	} else {
		defaultConfig = getDefaultConfig()
	}

	service, _ := NewService("test", WithSkipLoad(true), WithDefaultConfig(defaultConfig))
	return service
}
