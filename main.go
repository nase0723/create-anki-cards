package main

import (
	"embed"
	"log"

	"create-anki-cards/internal/cache"
	"create-anki-cards/internal/config"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	c, err := cache.New("cache")
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	app := NewApp(cfg, c)

	err = wails.Run(&options.App{
		Title:         "Anki Card Creator",
		Width:         600,
		Height:        1000,
		AlwaysOnTop:   true,
		StartHidden:   true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
