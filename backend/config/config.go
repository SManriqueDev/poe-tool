package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/models"
)

var (
	cfg        *models.Config
	configPath = getConfigPath()
	once       sync.Once
)

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	var appDir string
	switch runtime.GOOS {
	case "darwin":
		appDir = filepath.Join(home, "Library", "Application Support", "PoeTool")
	case "linux":
		appDir = filepath.Join(home, ".config", "poetool")
	default:
		appDir = filepath.Join(home, ".poetool")
	}
	os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "config.json")
}

func GetConfig() *models.Config {
	once.Do(func() {
		cfg = &models.Config{}
		file, err := os.ReadFile(configPath)
		if err == nil {
			_ = json.Unmarshal(file, cfg)
		}
	})
	return cfg
}

func SaveConfig(c *models.Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, data, 0644)
	if err == nil {
		cfg = c
	}
	return err
}
