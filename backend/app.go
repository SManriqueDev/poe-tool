package backend

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	"github.com/SManriqueDev/poe-tool/backend/internal/logging"
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	SettingsHandler   *settings.Handler
	LiveSearchHandler *livesearch.Handler
	LoggingHandler    *logging.Handler

	// Services for context management
	settingsService *settings.Service
	loggingService  *logging.Service
	lsService       *livesearch.Service
}

func NewApp() *App {
	settingsService, _ := settings.NewService("PoeTool")
	loggingService := logging.NewService(settingsService)
	lsService := livesearch.NewService(settingsService, loggingService)

	// Configure event emitter for real-time log updates
	lsService.SetupEventEmitter(loggingService)

	return &App{
		SettingsHandler:   settings.NewHandler(settingsService),
		LoggingHandler:    logging.NewHandler(loggingService),
		LiveSearchHandler: livesearch.NewHandler(lsService),

		// Store service references for context management
		settingsService: settingsService,
		loggingService:  loggingService,
		lsService:       lsService,
	}
}

func (a *App) Startup() {
	// In v3, services can handle their own initialization
	// Context will be provided by the Wails v3 runtime when needed
}

func (a *App) SetAppInstance(app *application.App) {
	// Configurar la función para que livesearch pueda acceder a la aplicación
	livesearch.GetAppInstance = func() *application.App {
		return app
	}
}

// SetupContexts configura los contextos de los servicios
func (a *App) SetupContexts(ctx context.Context) {
	a.loggingService.SetContext(ctx)
	a.lsService.SetContext(ctx)
}
