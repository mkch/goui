package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"image/color"
	"os"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {

	gg.MustOK(app.CreateWindow(goui.Window{
		Title: "dpi test",
		Width: 800, Height: 600,
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Root: &widgets.Center{
			Widget: &widgets.Column{
				CrossAxisAlignment: axes.Center,
				MainAxisSize:       axes.Min,
				Widgets: []goui.Widget{
					&widgets.SizedBox{
						Width: 70 * metrics.Millimeter, Height: 50 * metrics.Millimeter,
						Widget: &widgets.Label{
							BackgroundColor: &color.NRGBA{R: 100, G: 100, B: 100, A: 255},
							Text:            "7 CM x 5 CM",
						},
					},

					&widgets.SizedBox{Height: 20},

					&widgets.Label{
						Text: "Approximately 7 CM x 5 CM.",
					},
				},
			},
		},
	}))

	os.Exit(app.Run())
}
