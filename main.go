package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
			Middleware: newAssetsServer,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			app.emitFileOpening(nil)
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "cosmic_reader.wails",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				app.emitFileOpening(data.Args)
				runtime.WindowShow(app.ctx)
			},
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
