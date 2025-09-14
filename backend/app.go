package backend

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
)

type App struct {
	ctx               context.Context
	SettingsHandler   *settings.Handler
	LiveSearchHandler *livesearch.Handler
}

func NewApp() *App {
	settingsService, _ := settings.NewService("PoeTool")
	lsService := livesearch.NewService(settingsService)

	return &App{
		SettingsHandler:   settings.NewHandler(settingsService),
		LiveSearchHandler: livesearch.NewHandler(lsService),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.LiveSearchHandler.SetContext(ctx)
}
