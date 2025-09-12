package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Service struct {
	filePath string
	data     *Config
}

func NewService(appName string) (*Service, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user config directory: %w", err)
	}

	appDir := filepath.Join(configDir, appName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return nil, fmt.Errorf("could not create config directory: %w", err)
		}
	}

	filePath := filepath.Join(appDir, "config.json")
	s := &Service{filePath: filePath}

	if err := s.load(); err != nil {
		// initialize with default values
		s.data = &Config{
			PoeSessid:         "",
			AccountName:       "",
			League:            "Standard",
			AutomationEnabled: false,
			Delay:             1000,
			DefaultTradeLinks: []DefaultTradeLink{},
		}
		_ = s.Save()
	}

	return s, nil
}

func (s *Service) load() error {
	bytes, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return err
	}
	s.data = &cfg
	return nil
}

func (s *Service) Save() error {
	bytes, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, bytes, 0644)
}

func (s *Service) Get() *Config {
	return s.data
}

func (s *Service) Update(newData Config) error {
	s.data = &newData
	return s.Save()
}
