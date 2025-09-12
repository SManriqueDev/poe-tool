package services

import (
	"github.com/SManriqueDev/poe-tool/backend/config"
	"github.com/SManriqueDev/poe-tool/backend/models"
)

type SettingsService struct{}

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

func (s *SettingsService) Load() *models.Config {
	return config.GetConfig()
}

func (s *SettingsService) Save(cfg *models.Config) error {
	return config.SaveConfig(cfg)
}
