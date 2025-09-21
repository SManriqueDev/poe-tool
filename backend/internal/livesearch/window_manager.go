package livesearch

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// GetAppInstance necesita ser implementada para acceder a la instancia global
// Esto se resolverá cuando importemos desde main
var GetAppInstance func() *application.App

// OpenLogsWindow abre una nueva ventana dedicada para mostrar los logs de LiveSearch
func OpenLogsWindow() error {
	if GetAppInstance == nil {
		return fmt.Errorf("app instance getter not initialized")
	}

	app := GetAppInstance()
	if app == nil {
		return fmt.Errorf("app instance not available")
	}

	// Crear la nueva ventana para logs
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "LiveSearch Logs",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 0,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarDefault,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/livesearch-logs",
		Width:            1000,
		Height:           700,
		MinWidth:         800,
		MinHeight:        600,
		MaxWidth:         1400,
		MaxHeight:        1000,
	})

	return nil
}
