package main

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	app.CreateWindow(goui.Window{
		OnClose: func() { app.Exit(0) },
		Title:   "goui visibility sample",
		Width:   400,
		Height:  300,
		Root:    rootWidget(),
	})

	app.Run()
}

func rootWidget() goui.Widget {
	return &widgets.Center{Widget: demoWidget()}
}

func demoWidget() goui.StatefulWidget {
	return goui.StatefulWidgetFunc(func(ctx *goui.Context, updateState goui.UpdateStateFunc) *goui.WidgetState {
		var visible bool = true
		var maintainSize bool
		return &goui.WidgetState{
			Build: func() goui.Widget {
				return &widgets.Column{
					CrossAxisAlignment: axes.Center,
					Widgets: []goui.Widget{
						&widgets.SizedBox{Height: 10},

						&widgets.Visibility{
							Visible:      visible,
							MaintainSize: maintainSize,
							Widget: &widgets.Padding{
								Left: 5, Right: 5,
								Widget: &widgets.Label{
									Text: "The quick brown fox jumps over the lazy dog.",
								},
							},
						},

						&widgets.SizedBox{Height: 20},

						&widgets.Button{
							Label: "Show",
							OnClick: func(ctx *goui.Context) {
								if !visible {
									gg.MustOK(updateState(func() { visible = true }))
								}
							},
						},
						&widgets.Button{
							Label: "Hide",
							OnClick: func(ctx *goui.Context) {
								if visible || maintainSize {
									gg.MustOK(updateState(func() { visible = false; maintainSize = false }))
								}
							},
						},
						&widgets.Button{
							Label: "Hide, maintain size",
							OnClick: func(ctx *goui.Context) {
								if visible || !maintainSize {
									gg.MustOK(updateState(func() { visible = false; maintainSize = true }))
								}
							},
						},
					},
				}
			},
		}
	})
}
