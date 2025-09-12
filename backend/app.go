package backend

import (
	"context"
	"fmt"
	"github.com/SManriqueDev/poe-tool/backend/models"
	"github.com/SManriqueDev/poe-tool/backend/services"
)

type App struct {
	ctx             context.Context
	settingsService *services.SettingsService
}

func NewApp() *App {
	return &App{
		settingsService: services.NewSettingsService(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) LoadConfig() *models.Config {
	return a.settingsService.Load()
}

func (a *App) SaveConfig(cfg *models.Config) error {
	return a.settingsService.Save(cfg)
}

func (b *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s!", name)
}
