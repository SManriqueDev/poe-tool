# Poe Tool - AGENTS.md

## Project
Wails v3 desktop app (Go 1.24 + React 18 + TS 5 + Tailwind CSS 4) for Path of Exile trade automation.
Module: `github.com/SManriqueDev/poe-tool`

## Quick start
```bash
make deps          # npm install in frontend/
make dev           # wails3 dev (hot-reload Go + TS, Vite on :9245)
make lint          # biome lint ./src (runs from frontend/)
make format        # biome format --write ./src
make check         # biome check --write --unsafe ./src (applies unsafe fixes)
```

No `make test` target. Run tests with:
```bash
go test ./...                                               # all tests
go test ./backend/internal/livesearch/...                   # livesearch only
go test ./backend/internal/livesearch/application -run TestTradeLink  # single test
```

## Architecture
- **Clean Architecture**: `domain/` (pure Go interfaces/models) → `application/` (use cases) → `adapters/` (infra bridge)
- **Backend modules** under `backend/internal/`: `livesearch/` (core + domain + application sub-pkgs), `settings/`, `logging/`, `adapters/`
- **Frontend** under `frontend/src/`: `pages/`, `services/`, `components/ui/` (shadcn-style), `components/` (custom), `live-search/`
- **Entry point**: `main.go` → `db.Init("poe_tool.db", migrationsFS)` → `backend.NewApp()` (DI) → register `[]application.Service` → `wailsApp.Run()`

## Build
```bash
make build-windows             # Wails3 + CGO (build/windows/Taskfile.yml)
make build-windows-exe         # standalone Go .exe (CGO_ENABLED=0, smaller)
make build-windows-production  # Wails3 + CGO + -H windowsgui (no console)
make clean                     # rm -rf build/bin frontend/dist
```
Binaries go to `bin/`. Frontend must build (`tsc && vite build` → `frontend/dist/`) before Go compilation.

## Go patterns
- **Wails handler exposure**: exported methods on handler structs registered in `main.go` as `application.Service`. Parameters/returns must be JSON-serializable. Handler is auto-bound to frontend.
- **Real-time push**: `app.Event.Emit("eventName", data)` from Go → `EventsOn("eventName", cb)` in frontend (via `@wailsio/runtime`).
- **Database**: SQLite via `mattn/go-sqlite3`. Singleton via `sync.Once` in `backend/db/database.go`. Migrations in `migrations/` (Goose format). Settings stored as key-value pairs.
- **Test mocks**: two patterns coexist: (a) `testify/mock` structs in `application/` tests, (b) hand-written mock structs in package-level tests (`trade_link_manager_test.go`). Settings has `NewTestService(WithSkipLoad(true))` to avoid DB dependency.
- **`NewService` returns error**: in `backend/app.go`, the error from `settings.NewService` is intentionally ignored (`_, _`). Don't change this without reviewing.
- **Dev config**: `wails3 dev` loads from `build/config.yml`, not `wails.json`. The config runs `go mod tidy` + `wails3 task build` + `wails3 task run` on changes. Ignores `.git`, `node_modules`, `frontend`, `bin`.
- **`wails.json` is stale**: actual app name/version are in `build/config.yml`.

## Frontend patterns
- **Path aliases**: `@/` → `src/`, `~wails/` → `wailsjs/`
- **Binding imports**: `import { Handler } from "../../bindings/github.com/SManriqueDev/poe-tool/backend/internal/<pkg>/index.js"`
- **Routing**: `HashRouter` in `main.tsx`. Routes: `/` + `/settings` → Settings, `/search` → LiveSearch, `/logs` → Logs, `/livesearch-logs` → LiveSearchLogsWindow (separate window, no layout)
- **Theme**: `next-themes` dark/light via `ModeToggle` component
- **Style**: Tailwind CSS v4 via `@tailwindcss/vite` plugin (no PostCSS config). biome uses tabs for indent, double quotes.
- **Vite config**: ignores `frontend/bindings/` and `.bindings-tmp-*/` in watch mode. HMR via ws://localhost.
- **Backend changes require** `make dev` (or `wails3 dev`) restart to regenerate bindings under `frontend/bindings/`.

## Key constraints
- `go:embed` directives in `main.go` for `frontend/dist` and `migrations/` — don't break the embed path
- `agent` mode (Wails-generated CLI arg) not supported — handle gracefully
- Backend module directory in `frontend/bindings/` = `github.com/SManriqueDev/poe-tool/backend/internal/<pkg>`
- Handler methods must be **public** (CapitalCase) for Wails to bind them
- `make dev` / `wails3 dev` runs `biome` only on `./src` (cwd is `frontend/`). Running biome directly from repo root on the wrong path will fail or format wrong files.
