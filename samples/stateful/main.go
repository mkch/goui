package main

import (
	"fmt"

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
		Title:   "goui demo",
		Width:   600,
		Height:  400,
		Root: &widgets.Center{
			Widget: &widgets.Column{
				CrossAxisAlignment: axes.Center,
				Widgets: []goui.Widget{
					&widgets.SizedBox{Height: 10},
					&widgets.SizedBox{
						Width:  200,
						Height: 40,
						Widget: stateful,
					},
					&widgets.SizedBox{Height: 10},
					&widgets.Button{
						Label: "Increase State (Even: Label, Odd: Button)",
						OnClick: func(ctx *goui.Context) {
							gg.MustOK(updateNumber(func() { number++ }))
						},
					},
				},
			},
		},
	})
	app.Run()
}

var number = 0
var updateNumber goui.UpdateStateFunc

var stateful = goui.StatefulWidgetFunc(
	func(ctx *goui.Context, updateState goui.UpdateStateFunc) *goui.WidgetState {
		return &goui.WidgetState{
			Build: func() goui.Widget {
				updateNumber = updateState
				return gg.IfFunc(number%2 == 0,
					func() goui.Widget { return &widgets.Label{Text: fmt.Sprintf("Label: %v", number)} },
					func() goui.Widget { return &widgets.Button{Label: fmt.Sprintf("Button: %v", number)} },
				)
			}}
	})
