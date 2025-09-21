package backend

import (
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	"github.com/SManriqueDev/poe-tool/backend/internal/logging"
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
)

type App struct {
	SettingsHandler   *settings.Handler
	LiveSearchHandler *livesearch.Handler
	LoggingHandler    *logging.Handler
}

func NewApp() *App {
	settingsService, _ := settings.NewService("PoeTool")
	loggingService := logging.NewService(settingsService)
	lsService := livesearch.NewService(settingsService, loggingService)

	return &App{
		SettingsHandler:   settings.NewHandler(settingsService),
		LoggingHandler:    logging.NewHandler(loggingService),
		LiveSearchHandler: livesearch.NewHandler(lsService),
	}
}

func (a *App) Startup() {
	// In v3, services can handle their own initialization
	// Context will be provided by the Wails v3 runtime when needed
}
