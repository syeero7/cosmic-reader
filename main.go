package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:      "cosmic-reader",
		MinWidth:   1024,
		MinHeight:  768,
		Fullscreen: true,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: app.assetServer,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "cosmic_reader.wails",
			OnSecondInstanceLaunch: app.secondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
