# AI Coding Instructions for Poe Tool

## Project Overview
**Poe Tool** is a Wails v3 desktop application with a Go backend and React/TypeScript frontend for Path of Exile trade automation. Features real-time live search monitoring and automatic hideout teleportation.

## Architecture & Key Components

### Backend Structure (Go)
- **Wails v3 Framework**: Desktop app with embedded frontend (`main.go` embeds `frontend/dist`)
- **Clean Architecture**: Domain → Application → Infrastructure layers with dependency injection
- **Modules**: `livesearch`, `settings`, and `logging` in `backend/internal/`
- **Application Services**: Business logic in `backend/internal/livesearch/application/` (e.g., `TradeLinkApplicationService`, `LiveSearchApplicationService`)
- **Domain Layer**: Interfaces and models in `backend/internal/livesearch/domain/` defining contracts
- **Adapters**: Infrastructure implementations in `backend/internal/adapters/` bridging domain interfaces
- **Database**: SQLite via `backend/db/database.go` with Goose migrations in `migrations/`

### Frontend Structure (React/TypeScript)
- **Build Tool**: Vite with TypeScript
- **UI Framework**: Shadcn UI components on Tailwind CSS 4
- **Layout**: `Layout.tsx` with sidebar navigation (`AppSidebar`)
- **Wails Integration**: Generated bindings in `wailsjs/go/` for Go method calls (auto-regenerated with `wails3 dev`)
- **Services**: TypeScript wrappers in `src/services/` for backend calls with type safety
- **Real-time Updates**: Wails events for backend → frontend communication (e.g., `EventsOn("linkStatusChanged")`)

## Key Patterns

### Backend Service Creation & Dependency Injection
```go
// Services created in backend/app.go with explicit dependency injection
func NewApp() *App {
    settingsService, _ := settings.NewService("PoeTool")
    loggingService := logging.NewService(settingsService)      // depends on settings
    lsService := livesearch.NewService(settingsService, loggingService)  // depends on both

    // Domain factory creates infrastructure components
    domainFactory := adapters.NewDomainComponentsFactory(domainConfig, loggerAdapter)
    domainWebSocketClient := domainFactory.CreateWebSocketClient()
    domainEventBus := domainFactory.CreateEventBus()

    // Application services use domain interfaces
    tradeLinkAppSvc := lsapplication.NewTradeLinkApplicationService(domainTradeLinkRepo, loggerAdapter)
    liveSearchAppSvc := lsapplication.NewLiveSearchApplicationService(
        domainTradeLinkRepo, domainLiveSearchRepo, domainWebSocketClient, domainEventBus, loggerAdapter)

    return &App{
        LiveSearchHandler: livesearch.NewHandler(lsService, tradeLinkAppSvc, hideoutAppSvc, liveSearchAppSvc),
    }
}
```

### Application Service Pattern
```go
// Application services implement use cases using domain interfaces
type TradeLinkApplicationService struct {
    repo   domain.TradeLinkRepository
    logger domain.Logger
}

func (s *TradeLinkApplicationService) AddTradeLink(ctx context.Context, url, description string) error {
    tradeLink := &domain.TradeLink{URL: url, Description: description, Selected: true}
    return s.repo.Create(ctx, tradeLink)
}
```

### Wails v3 Service Binding Pattern
```go
// In main.go - services auto-bound to frontend
wailsApp := application.New(application.Options{
    Services: []application.Service{
        application.NewService(app.SettingsHandler),
        application.NewService(app.LiveSearchHandler),
        application.NewService(app.LoggingHandler),
    },
})
```

### Frontend-Backend Communication
- Go methods bound in `main.go` become callable JS functions
- Generated TypeScript bindings in `wailsjs/go/[module]/Handler.js`
- Service layer wraps Wails calls: `src/services/liveSearchService.ts`
- Real-time updates via Wails events: `EventsOn("linkStatusChanged", callback)`

### Database Conventions
- SQLite with migrations in `migrations/` (Goose format: `00001_create_tables.sql`)
- Repository pattern with standard CRUD methods
- Settings stored as key-value pairs in dedicated table
- Use transactions for multi-table operations

## Development Workflow

### Essential Commands
```bash
# Development (hot reload frontend + backend)
make dev

# Frontend-only changes (browser debugging)
cd frontend && npm run dev  # Opens http://localhost:34115

# Code quality
make lint    # Biome linting
make format  # Biome formatting
make check   # Biome auto-fix

# Building
make build-mac     # macOS universal
make build-windows # Windows x64
make build-linux   # Linux x64

# Alternative task-based builds (Wails v3)
wails3 task build        # Current platform
wails3 task windows:build # Cross-compile Windows
```

### Adding New Features

#### Backend Changes
1. Add domain interfaces in `backend/internal/livesearch/domain/interfaces.go`
2. Implement domain logic in `backend/internal/livesearch/domain/usecases.go`
3. Create application service in `backend/internal/livesearch/application/`
4. Add adapter implementation in `backend/internal/adapters/`
5. Expose via Handler (`backend/internal/[module]/handler.go`)
6. Bind in `main.go` if new module, or existing handler auto-updates
7. Database changes require migration in `migrations/`

#### Frontend Changes
1. Run `wails3 dev` to regenerate bindings after backend changes
2. Create service wrapper in `src/services/` for type safety
3. Use Shadcn UI components: `npx shadcn-ui@latest add [component]`
4. Follow the data table pattern (`src/live-search/`) for complex UIs

## Project-Specific Conventions

### File Organization
- Backend modules: `backend/internal/[feature]/` with handler, service, repository, model
- Application layer: `backend/internal/livesearch/application/` for use case implementations
- Domain layer: `backend/internal/livesearch/domain/` for interfaces and core business logic
- Adapters: `backend/internal/adapters/` for infrastructure implementations
- Frontend features: Organize by domain in `src/pages/` and `src/[feature]/`
- Shared UI: `src/components/ui/` (Shadcn) and `src/components/` (custom)

### Clean Architecture Status
- **Completed**: Domain interfaces, application services, and adapter implementations
- **Current State**: Full clean architecture with domain components integrated into application services
- **Migration Pattern**: Domain components created via factory, injected into application services

### Styling Approach
- **Tailwind CSS 4**: Use utility classes with new v4 syntax
- **Shadcn UI**: Component library for consistent design system
- **Theme Support**: Dark/light mode via `next-themes` and `ModeToggle` component

### State Management
- Local state with React hooks for UI concerns
- Wails events for real-time backend → frontend communication
- No global state library - backend acts as state source

### Database Patterns
- Repository pattern with methods like `GetByID`, `Create`, `Update`, `Delete`
- Settings stored as key-value pairs in dedicated table
- Use transactions for multi-table operations

## Integration Points

### WebSocket Integration
- Custom WebSocket client in `backend/internal/livesearch/websocket_client.go`
- Event bus pattern for decoupled communication
- Real-time updates broadcast via Wails events to frontend

### External Dependencies
- **Go**: Wails v3, SQLite driver (`github.com/mattn/go-sqlite3`)
- **Frontend**: Wails runtime, Tanstack Table, Radix UI primitives
- **Build**: Biome for code quality, Vite for bundling

## Critical Debugging Points

### Common Issues
- **Binding errors**: Check handler methods are public and in `main.go` Bind slice
- **Context issues**: Ensure `SetContext()` called on handlers after app startup
- **Build failures**: Frontend must build successfully (`npm run build`) for Wails compilation
- **Database**: Ensure `db.Init()` called before service creation in `main.go`

### Development Debugging
- Use `wails3 dev` with browser at `:34115` for frontend debugging
- Go backend logs appear in terminal running `wails3 dev`
- Database file `poe_tool.db` in project root for inspection

## Testing Strategy
- Application service tests in `backend/internal/livesearch/application/`
- Repository tests for data access layer
- Manual testing via `make dev` and frontend interaction
- Database testing can use temporary SQLite files
