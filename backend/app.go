package backend

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
)

type App struct {
	ctx             context.Context
	SettingsHandler *settings.Handler
}

func NewApp() *App {
	settingsService, _ := settings.NewService("PoeTool")

	return &App{
		SettingsHandler: settings.NewHandler(settingsService),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}
