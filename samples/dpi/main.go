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
	"github.com/mkch/goui/widgets/label"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {

	gg.MustOK(app.CreateWindow(&goui.Window{
		Title: "dpi test",
		Width: 800, Height: 850,
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
							TextAlignment:   label.Center,
							Text:            "7 CM x 5 CM",
						},
					},

					&widgets.SizedBox{Height: 20},

					&widgets.Label{
						Multiline: true,
						Text: `The size of above rectangle is approximately 7 CM x 5 CM.

Windows DPI changes with scaling: 
	1x: 96 DPI
	2x: 192 DPI
	……
If the result DPI is the same as the actual physical DPI of your monitor, the size of the gray box should be 7 CM x 5 CM.
For example, if the resolution of your monitor is 3840 x 2160 and the diagonal length is 27 inches, the physical DPI is about 163:
 	√(3840² + 2160²) / 27 ≈ 163
If you set Windows scaling to 175%, the effective DPI will be about 168:
 	96 * 1.75 = 168
In this case, the exact size of the gray box should be:
	(7 * 168 / 163) CM x (5 * 168 / 163) CM ≈ 7.2 CM x 5.2 CM`,
					},
				},
			},
		},
	}))

	os.Exit(app.Run())
}
