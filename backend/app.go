package backend

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/adapters"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	lsapplication "github.com/SManriqueDev/poe-tool/backend/internal/livesearch/application"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/SManriqueDev/poe-tool/backend/internal/logging"
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	SettingsHandler   *settings.Handler
	LiveSearchHandler *livesearch.Handler
	LoggingHandler    *logging.Handler

	settingsService *settings.Service
	loggingService  *logging.Service
	lsService       *livesearch.Service

	windowManager     domain.WindowManager
	hideoutAutomation domain.HideoutAutomation
	systemAPIClient   domain.SystemAPIClient
}

func NewApp() *App {
	settingsService, _ := settings.NewService("PoeTool")
	loggingService := logging.NewService(settingsService)
	lsService := livesearch.NewService(settingsService, loggingService)

	lsService.SetupEventEmitter(loggingService)

	domainTradeLinkRepo := adapters.NewDomainTradeLinkRepository()
	domainLiveSearchRepo := adapters.NewDomainLiveSearchRepository()

	loggerAdapter := adapters.NewLoggerAdapter(loggingService)
	domainConfig := adapters.DefaultDomainConfig()
	domainFactory := adapters.NewDomainComponentsFactory(domainConfig, loggerAdapter)

	domainWebSocketClient := domainFactory.CreateWebSocketClient()
	domainEventBus := domainFactory.CreateEventBus()

	domainSystemAPIClient := domainFactory.CreateSystemAPIClient()
	domainWindowManager := domainFactory.CreateWindowManager()
	domainHideoutAutomation := domainFactory.CreateHideoutAutomation(domainSystemAPIClient, domainLiveSearchRepo)

	tradeLinkAppSvc := lsapplication.NewTradeLinkApplicationService(domainTradeLinkRepo, loggerAdapter)
	hideoutAppSvc := lsapplication.NewHideoutApplicationService(domainLiveSearchRepo, domainHideoutAutomation, loggerAdapter)

	// FASE 5: Crear LiveSearchApplicationService completo
	liveSearchAppSvc := lsapplication.NewLiveSearchApplicationService(
		domainTradeLinkRepo,
		domainLiveSearchRepo,
		domainWebSocketClient,
		domainEventBus,
		loggerAdapter,
	)

	return &App{
		SettingsHandler:   settings.NewHandler(settingsService),
		LoggingHandler:    logging.NewHandler(loggingService),
		LiveSearchHandler: livesearch.NewHandler(tradeLinkAppSvc, hideoutAppSvc, liveSearchAppSvc, loggerAdapter, domainWindowManager),

		settingsService: settingsService,
		loggingService:  loggingService,
		lsService:       lsService,

		windowManager:     domainWindowManager,
		hideoutAutomation: domainHideoutAutomation,
		systemAPIClient:   domainSystemAPIClient,
	}
}

func (a *App) Startup() {
}

func (a *App) SetAppInstance(app *application.App) {
	livesearch.GetAppInstance = func() *application.App {
		return app
	}

	if domainWindowManager, ok := a.windowManager.(*adapters.DomainWindowManager); ok {
		domainWindowManager.SetApplication(app)
	}
}

func (a *App) SetupContexts(ctx context.Context) {
	a.loggingService.SetContext(ctx)
	a.lsService.SetContext(ctx)
}
