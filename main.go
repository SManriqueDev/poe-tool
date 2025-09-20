package main

import (
	"embed"
	"log"

	"github.com/SManriqueDev/poe-tool/backend"
	"github.com/SManriqueDev/poe-tool/backend/db"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	err := db.Init("poe_tool.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	app := backend.NewApp()
	err = wails.Run(&options.App{
		Title:  "Poe Tool",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app.SettingsHandler,
			app.LiveSearchHandler,
			app.LoggingHandler,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
