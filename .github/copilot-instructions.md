# AI Coding Instructions for Poe Tool

## Project Overview
**Poe Tool** is a Wails desktop application with a Go backend and React/TypeScript frontend. It's designed as a tool for Path of Exile game players to manage live searches for trade links. The app uses a modern stack: React 18, TypeScript, Shadcn UI, Tailwind CSS 4, and SQLite for persistence.

## Architecture & Key Components

### Backend Structure (Go)
- **Wails Framework**: Desktop app with embedded frontend (`main.go` embeds `frontend/dist`)
- **Layered Architecture**: Handler → Service → Repository pattern
- **Modules**: `livesearch` and `settings` in `backend/internal/`
- **Database**: SQLite via `backend/db/database.go` with singleton pattern
- **Context Flow**: App startup sets context, handlers receive it via `SetContext()`

### Frontend Structure (React/TypeScript)
- **Build Tool**: Vite with TypeScript
- **UI Framework**: Shadcn UI components on Tailwind CSS 4
- **Layout**: `Layout.tsx` with sidebar navigation (`AppSidebar`)
- **Wails Integration**: Generated bindings in `wailsjs/` for Go method calls
- **Services**: TypeScript wrappers in `src/services/` for backend calls

### Key Patterns

#### Backend Service Pattern
```go
// Services are created in backend/app.go and injected
type Service struct {
    repo *Repository
    settingsSvc *settings.Service  // Cross-service dependencies
}

// Handlers expose methods to frontend
type Handler struct {
    svc *Service
}
```

#### Frontend-Backend Communication
- Go methods bound in `main.go` become callable JS functions
- Generated TypeScript bindings in `wailsjs/go/[module]/Handler.js`
- Service layer wraps Wails calls: `src/services/liveSearchService.ts`
- Real-time updates via Wails events: `EventsOn("linkStatusChanged", callback)`

#### Database Conventions
- SQLite with migrations in `migrations/` (Goose format)
- Repository pattern for data access
- Tables: `trade_links`, `settings`, `live_search_settings`

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
```

### Adding New Features

#### Backend Changes
1. Add methods to Service (`backend/internal/[module]/service.go`)
2. Expose via Handler (`backend/internal/[module]/handler.go`)
3. Bind in `main.go` if new module, or existing handler auto-updates
4. Database changes require migration in `migrations/`

#### Frontend Changes
1. Run `wails dev` to regenerate bindings after backend changes
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
- **Go**: Wails v2, SQLite driver (`github.com/mattn/go-sqlite3`)
- **Frontend**: Wails runtime, Tanstack Table, Radix UI primitives
- **Build**: Biome for code quality, Vite for bundling

## Critical Debugging Points

### Common Issues
- **Binding errors**: Check handler methods are public and in `main.go` Bind slice
- **Context issues**: Ensure `SetContext()` called on handlers after app startup
- **Build failures**: Frontend must build successfully (`npm run build`) for Wails compilation
- **Database**: Ensure `db.Init()` called before service creation in `main.go`

### Development Debugging
- Use `wails dev` with browser at `:34115` for frontend debugging
- Go backend logs appear in terminal running `wails dev`
- Database file `poe_tool.db` in project root for inspection

## Testing Strategy
- No test framework currently configured
- Manual testing via `make dev` and frontend interaction
- Database testing can use temporary SQLite files
