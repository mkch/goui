package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"image/color"

	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/check"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	check.MustOK(app.CreateWindow(&goui.Window{
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "panel sample",
		Width:     800,
		Height:    600,
		Root:      rootWidget(),
	}))

	app.Run()
}

func rootWidget() goui.Widget {
	return &widgets.Padding{
		Left: 100,
		Widget: &widgets.Center{
			Widget: &widgets.Panel{
				BackgroundColor: &color.NRGBA{R: 200, G: 200},
				Widget: &widgets.Padding{
					Left: 40, Right: 80, Top: 30, Bottom: 60,
					Widget: &widgets.Row{
						MainAxisSize: axes.Min,
						Widgets: []goui.Widget{
							&widgets.Button{Label: "Click Me"},
							&widgets.Button{Label: "Click Me ~~~~"},
						},
					},
				},
			},
		},
	}
}
