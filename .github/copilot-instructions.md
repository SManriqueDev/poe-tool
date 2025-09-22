# AI Coding Instructions for Poe Tool

## Project Overview
**Poe Tool** is a Wails v3 desktop application with a Go backend and React/TypeScript frontend. It's designed as a Path of Exile trade automation tool featuring real-time live search monitoring and automatic hideout teleportation. The app uses a modern stack: React 18, TypeScript, Shadcn UI, Tailwind CSS 4, and SQLite for persistence.

## Architecture & Key Components

### Backend Structure (Go)
- **Wails v3 Framework**: Desktop app with embedded frontend (`main.go` embeds `frontend/dist`)
- **Layered Architecture**: Handler → Service → Repository pattern with dependency injection
- **Modules**: `livesearch`, `settings`, and `logging` in `backend/internal/`
- **Database**: SQLite via `backend/db/database.go` with Goose migrations in `migrations/`
- **Service Dependencies**: Cross-service injection (e.g., `livesearch` depends on `settings` and `logging`)
- **Context Flow**: App startup creates services with dependencies, then sets contexts via `SetupContexts()`

### Frontend Structure (React/TypeScript)
- **Build Tool**: Vite with TypeScript
- **UI Framework**: Shadcn UI components on Tailwind CSS 4
- **Layout**: `Layout.tsx` with sidebar navigation (`AppSidebar`)
- **Wails Integration**: Generated bindings in `wailsjs/` for Go method calls (auto-regenerated with `wails3 dev`)
- **Services**: TypeScript wrappers in `src/services/` for backend calls with type safety
- **Real-time Updates**: Wails events for backend → frontend communication (e.g., `EventsOn("linkStatusChanged")`)

### Key Patterns

#### Backend Service Creation & Dependency Injection
```go
// Services created in backend/app.go with explicit dependency injection
func NewApp() *App {
    settingsService, _ := settings.NewService("PoeTool")
    loggingService := logging.NewService(settingsService)      // depends on settings
    lsService := livesearch.NewService(settingsService, loggingService)  // depends on both

    // Configure cross-service communication
    lsService.SetupEventEmitter(loggingService)

    return &App{
        SettingsHandler:   settings.NewHandler(settingsService),
        LoggingHandler:    logging.NewHandler(loggingService),
        LiveSearchHandler: livesearch.NewHandler(lsService),
    }
}
```

#### Wails v3 Service Binding Pattern
```go
// In main.go - services auto-bound to frontend
wailsApp := application.New(application.Options{
    Services: []application.Service{
        application.NewService(app.SettingsHandler),
        application.NewService(app.LiveSearchHandler),
        application.NewService(app.LoggingHandler),  // New logging module
    },
})
```

#### Frontend-Backend Communication
- Go methods bound in `main.go` become callable JS functions
- Generated TypeScript bindings in `wailsjs/go/[module]/Handler.js`
- Service layer wraps Wails calls: `src/services/liveSearchService.ts`
- Real-time updates via Wails events: `EventsOn("linkStatusChanged", callback)`

#### Database Conventions
- SQLite with migrations in `migrations/` (Goose format: `00001_create_tables.sql`)
- Repository pattern for data access with standard CRUD methods
- Tables: `trade_links`, `settings`, `live_search_settings`, `logs` (with performance indexes)
- Settings stored as key-value pairs; logging with structured metadata

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
1. Add methods to Service (`backend/internal/[module]/service.go`)
2. Expose via Handler (`backend/internal/[module]/handler.go`)
3. Bind in `main.go` if new module, or existing handler auto-updates
4. Database changes require migration in `migrations/`

#### Frontend Changes
1. Run `wails3 dev` to regenerate bindings after backend changes
2. Create service wrapper in `src/services/` for type safety
3. Use Shadcn UI components: `npx shadcn-ui@latest add [component]`
4. Follow the data table pattern (`src/live-search/`) for complex UIs

## Project-Specific Conventions

### File Organization
- Backend modules: `backend/internal/[feature]/` with handler, service, repository, model
- Frontend features: Organize by domain in `src/pages/` and `src/[feature]/`
- Shared UI: `src/components/ui/` (Shadcn) and `src/components/` (custom)

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
- No test framework currently configured
- Manual testing via `make dev` and frontend interaction
- Database testing can use temporary SQLite files
