package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	var window *application.WebviewWindow
	app := application.New(application.Options{
		Name: "cosmic-reader",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
			Middleware: func(next http.Handler) http.Handler {
				return newAssetsServer(next)
			},
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "cosmic_reader.wails",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				if window != nil {
					window.Restore()
					window.Focus()
				}
			},
		},
	})

	window = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "cosmic-reader",
		MinWidth:  1024,
		MinHeight: 768,
	})

	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(event *application.ApplicationEvent) {
		window.Fullscreen()
		data := event.Context().Data()
		log.Println("app data: ", data)
	})

	app.RegisterService(application.NewService(NewMainService(app)))

	if err := app.Run(); err != nil {
		println("Error:", err.Error())
	}
}
