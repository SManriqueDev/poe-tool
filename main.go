package main

import (
	"context"
	"embed"
	"log"

	"github.com/SManriqueDev/poe-tool/backend"
	"github.com/SManriqueDev/poe-tool/backend/db"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Variable global para acceder a la aplicación desde los servicios
var appInstance *application.App

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	err := db.Init("poe_tool.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	app := backend.NewApp()

	// Create a new Wails application
	wailsApp := application.New(application.Options{
		Name:        "Poe Tool",
		Description: "A tool for Path of Exile game players to manage live searches",
		Services: []application.Service{
			application.NewService(app.SettingsHandler),
			application.NewService(app.LiveSearchHandler),
			application.NewService(app.LoggingHandler),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Asignar la instancia de app a la variable global
	appInstance = wailsApp

	// Configurar la aplicación para que los servicios puedan acceder a la instancia
	app.SetAppInstance(wailsApp)

	// Configurar contextos de los servicios
	ctx := context.Background()
	app.SetupContexts(ctx)

	// Create a new window
	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Poe Tool",
		Width:  1024,
		Height: 768,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 0,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarDefault,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Initialize the app startup
	app.Startup()

	// Run the application
	err = wailsApp.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
