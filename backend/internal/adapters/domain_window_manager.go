package adapters

import (
	"context"
	"fmt"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// DomainWindowManager implementa domain.WindowManager de forma pura
type DomainWindowManager struct {
	logger domain.Logger

	// Window tracking
	activeWindows map[string]*application.WebviewWindow
	windowsMu     sync.RWMutex

	// Application instance
	app   *application.App
	appMu sync.RWMutex

	// Configuration
	defaultWidth  int
	defaultHeight int
}

// NewDomainWindowManager crea una nueva instancia del window manager
func NewDomainWindowManager(logger domain.Logger, config WindowManagerConfig) *DomainWindowManager {
	return &DomainWindowManager{
		logger:        logger,
		activeWindows: make(map[string]*application.WebviewWindow),
		defaultWidth:  config.DefaultWidth,
		defaultHeight: config.DefaultHeight,
	}
}

// SetApplication establece la instancia de la aplicación
func (wm *DomainWindowManager) SetApplication(app *application.App) {
	wm.appMu.Lock()
	defer wm.appMu.Unlock()
	wm.app = app

	wm.logger.Info("window_manager", "Application instance configured", map[string]interface{}{
		"app_configured": app != nil,
	})
}

// OpenLogsWindow abre una nueva ventana dedicada para mostrar los logs
func (wm *DomainWindowManager) OpenLogsWindow(ctx context.Context) error {
	wm.appMu.RLock()
	app := wm.app
	wm.appMu.RUnlock()

	if app == nil {
		wm.logger.Error("window_manager", "Cannot open logs window: application not configured", nil)
		return fmt.Errorf("application not configured")
	}

	windowID := "logs-window"

	// Check if window already exists
	wm.windowsMu.RLock()
	if _, exists := wm.activeWindows[windowID]; exists {
		wm.windowsMu.RUnlock()
		// Bring existing window to front
		if err := wm.BringToFront(ctx, windowID); err != nil {
			wm.logger.Warning("window_manager", "Failed to bring logs window to front", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return nil
	}
	wm.windowsMu.RUnlock()

	// Create new window using Wails v3 API
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Logs - Poe Tool",
		Width:            1000,
		Height:           700,
			URL:              "/#/logs", // Route to logs page
		BackgroundColour: application.NewRGB(27, 38, 54),
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 0,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarDefault,
		},
	})

	if window == nil {
		wm.logger.Error("window_manager", "Failed to create logs window", nil)
		return fmt.Errorf("failed to create window")
	}

	// Register listener to clean up when user closes the window via X
	window.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		wm.windowsMu.Lock()
		delete(wm.activeWindows, windowID)
		wm.windowsMu.Unlock()
	})

	// Track the window
	wm.windowsMu.Lock()
	wm.activeWindows[windowID] = window
	wm.windowsMu.Unlock()

	wm.logger.Info("window_manager", "Logs window opened successfully", map[string]interface{}{
		"window_id": windowID,
		"width":     1000,
		"height":    700,
	})

	return nil
}

// CloseWindow cierra una ventana específica
func (wm *DomainWindowManager) CloseWindow(ctx context.Context, windowID string) error {
	wm.windowsMu.Lock()
	defer wm.windowsMu.Unlock()

	window, exists := wm.activeWindows[windowID]
	if !exists {
		return fmt.Errorf("window not found: %s", windowID)
	}

	// Close the window
	window.Close()

	// Remove from tracking
	delete(wm.activeWindows, windowID)

	wm.logger.Info("window_manager", "Window closed successfully", map[string]interface{}{
		"window_id": windowID,
	})

	return nil
}

// ShowWindow muestra una ventana específica
func (wm *DomainWindowManager) ShowWindow(ctx context.Context, windowID string) error {
	wm.windowsMu.RLock()
	window, exists := wm.activeWindows[windowID]
	wm.windowsMu.RUnlock()

	if !exists {
		return fmt.Errorf("window not found: %s", windowID)
	}

	window.Show()

	wm.logger.Debug("window_manager", "Window shown", map[string]interface{}{
		"window_id": windowID,
	})

	return nil
}

// HideWindow oculta una ventana específica
func (wm *DomainWindowManager) HideWindow(ctx context.Context, windowID string) error {
	wm.windowsMu.RLock()
	window, exists := wm.activeWindows[windowID]
	wm.windowsMu.RUnlock()

	if !exists {
		return fmt.Errorf("window not found: %s", windowID)
	}

	window.Hide()

	wm.logger.Debug("window_manager", "Window hidden", map[string]interface{}{
		"window_id": windowID,
	})

	return nil
}

// GetActiveWindows retorna información de todas las ventanas activas
func (wm *DomainWindowManager) GetActiveWindows(ctx context.Context) ([]domain.WindowInfo, error) {
	wm.windowsMu.RLock()
	defer wm.windowsMu.RUnlock()

	var windows []domain.WindowInfo

	for windowID, window := range wm.activeWindows {
		// Get window properties (basic implementation)
		windowInfo := domain.WindowInfo{
			ID:      windowID,
			Title:   "Poe Tool Window", // Could be enhanced to get actual title
			Visible: true,              // Assuming visible if tracked
			Active:  false,             // Could be enhanced to check if focused
		}

		// Set default position and size
		windowInfo.Position.X = 100
		windowInfo.Position.Y = 100
		windowInfo.Size.Width = wm.defaultWidth
		windowInfo.Size.Height = wm.defaultHeight

		// Override for logs window
		if windowID == "logs-window" {
			windowInfo.Title = "Logs - Poe Tool"
			windowInfo.Size.Width = 1000
			windowInfo.Size.Height = 700
		}

		windows = append(windows, windowInfo)

		// Prevent unused variable warning
		_ = window
	}

	wm.logger.Debug("window_manager", "Retrieved active windows", map[string]interface{}{
		"count": len(windows),
	})

	return windows, nil
}

// BringToFront trae una ventana al frente
func (wm *DomainWindowManager) BringToFront(ctx context.Context, windowID string) error {
	wm.windowsMu.RLock()
	window, exists := wm.activeWindows[windowID]
	wm.windowsMu.RUnlock()

	if !exists {
		return fmt.Errorf("window not found: %s", windowID)
	}

	// Show and focus the window
	window.Show()
	// Note: Wails v3 might have additional methods for focusing/bringing to front
	// This is a basic implementation that shows the window

	wm.logger.Debug("window_manager", "Window brought to front", map[string]interface{}{
		"window_id": windowID,
	})

	return nil
}

// GetWindowCount retorna el número de ventanas activas
func (wm *DomainWindowManager) GetWindowCount() int {
	wm.windowsMu.RLock()
	defer wm.windowsMu.RUnlock()
	return len(wm.activeWindows)
}


