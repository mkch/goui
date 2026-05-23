package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"image/color"
	"os"

	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

func main() {
	os.Exit(goui.Run(ui, &goui.AppConfig{
		Debug: &goui.Debug{
			LayoutOutline: true,
		},
	}))
}

func ui(app *goui.App) {
	chkerr.MustOK(app.CreateWindow(&goui.Window{
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "panel sample",
		Width:     800,
		Height:    600,
		Root:      rootWidget(),
	}))
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
