package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"image/color"

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	err := app.CreateWindow(goui.Window{
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "panel sample",
		Width:     400,
		Height:    300,
		Root:      rootWidget(),
	})

	if err != nil {
		errortrace.Panic(err)
	}

	app.Run()
}

func rootWidget() goui.Widget {
	return &widgets.Padding{
		Left: 100,
		Widget: &widgets.Center{
			Widget: &widgets.Panel{
				BackgroundColor: &color.NRGBA{R: 200, G: 200},
				Widget: &widgets.Padding{
					Left: 20, Right: 20, Top: 10, Bottom: 10,
					Widget: &widgets.Button{Label: "Click Me"},
				},
			},
		},
	}
}
