.PHONY: dev build-windows build-linux build-mac build-windows-cross package-windows-cross package-windows package-linux package-mac build-windows-exe build-windows-exe-cgo build-windows-exe-cgo-noconsole build-windows-wails build-windows-production build-windows-script clean deps lint format check help

# Run the application in development mode
dev:
	wails3 dev

# Build the application for Windows
build-windows:
	wails3 task build

# Build the application for Linux
build-linux:
	wails3 task build

# Build the application for macOS
build-mac:
	wails3 task build

# Cross-compile for Windows from any platform
build-windows-cross:
	@echo "Building with Wails3 (CGO_ENABLED=1 configured in Taskfile.yml)..."
	wails3 task windows:build

# Cross-compile package for Windows from any platform
package-windows-cross:
	wails3 task windows:package

# Build Windows executable with Go (no CGO)
build-windows-exe:
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Building Windows executable (no CGO)..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/poe-tool.exe .
	@echo "Windows executable created: bin/poe-tool.exe"

# Build Windows executable with Go (with CGO)
build-windows-exe-cgo:
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Building Windows executable (with CGO)..."
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s" -o bin/poe-tool-cgo.exe .
	@echo "Windows executable with CGO created: bin/poe-tool-cgo.exe"

# Build Windows executable with Go (with CGO, no console window)
build-windows-exe-cgo-noconsole:
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Building Windows executable (with CGO, no console window)..."
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s -H windowsgui" -o bin/poe-tool-cgo-noconsole.exe .
	@echo "Windows executable with CGO (no console) created: bin/poe-tool-cgo-noconsole.exe"

# Build Windows executable with Wails3 (full desktop features + CGO)
build-windows-wails:
	@echo "Building Windows executable with Wails3 (includes desktop features)..."
	@echo "Using: CGO_ENABLED=1 (configured in build/windows/Taskfile.yml)"
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 wails3 task windows:build
	@echo "Windows executable with Wails3 created: bin/poe-tool.exe"

# Build Windows executable with Wails3 in PRODUCTION mode (no console window)
build-windows-production:
	@echo "Building Windows executable in PRODUCTION mode (no console window)..."
	@echo "Using: CGO_ENABLED=1, PRODUCTION=true, -H windowsgui"
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 PRODUCTION=true wails3 task windows:build
	@echo "Windows production executable created: bin/poe-tool.exe (no console)"

# Build Windows executable using the interactive script
build-windows-script:
	@./build-windows.sh

# Package the application for Windows
package-windows:
	wails3 task package

# Package the application for Linux
package-linux:
	wails3 task package

# Package the application for macOS
package-mac:
	wails3 task package

# Clean build artifacts
clean:
	rm -rf build/bin
	cd frontend && rm -rf dist

# Install frontend dependencies
deps:
	cd frontend && npm install

# Lint frontend code
lint:
	cd frontend && npm run lint

# Format frontend code
format:
	cd frontend && npm run format

# Check frontend code
check:
	cd frontend && npm run check

# Help command to show available targets
help:
	@echo "Available commands:"
	@echo "  make dev                   - Run the application in development mode (Wails3)"
	@echo "  make build-windows         - Build the application (will use Windows when run on Windows)"
	@echo "  make build-linux           - Build the application (will use Linux when run on Linux)"
	@echo "  make build-mac             - Build the application (will use macOS when run on macOS)"
	@echo "  make build-windows-cross       - Cross-compile for Windows (Wails3 + CGO)"
	@echo "  make build-windows-exe         - Build Windows .exe with Go (CGO_ENABLED=0, smaller size)"
	@echo "  make build-windows-exe-cgo     - Build Windows .exe with Go (CGO_ENABLED=1, with console)"
	@echo "  make build-windows-exe-cgo-noconsole - Build Windows .exe with Go (CGO_ENABLED=1, no console)"
	@echo "  make build-windows-wails       - Build Windows .exe with Wails3 (CGO_ENABLED=1, with console)"
	@echo "  make build-windows-production  - Build Windows .exe with Wails3 (CGO_ENABLED=1, no console) 🏆"
	@echo "  make build-windows-script  - Interactive script to build all Windows versions"
	@echo "  make package-windows-cross - Cross-compile Windows package from any platform"
	@echo "  make package-windows       - Package the application (will use Windows when run on Windows)"
	@echo "  make package-linux         - Package the application (will use Linux when run on Linux)"
	@echo "  make package-mac           - Package the application (will use macOS when run on macOS)"
	@echo "  make clean                 - Clean build artifacts"
	@echo "  make deps                  - Install frontend dependencies"
	@echo "  make lint                  - Lint frontend code"
	@echo "  make format                - Format frontend code"
	@echo "  make check                 - Check and apply fixes to frontend code"
