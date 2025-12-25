package main

import (
	"fmt"

	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
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
		Root: &widgets.Column{Widgets: []goui.Widget{
			&widgets.Center{
				HeightFactor: 120,
				Widget: &widgets.SizedBox{
					Width: 80, Height: 30,
					Widget: &widgets.Button{
						Label: "Click me!",
						OnClick: func(*goui.Context) {
							fmt.Println("Button clicked!")
						},
					},
				},
			},
			&widgets.Center{
				HeightFactor: 120,
				Widget: &widgets.SizedBox{
					Width: 300, Height: 30,
					Widget: &widgets.Button{
						Label: "Click\r\nme!",
						OnClick: func(*goui.Context) {
							fmt.Println("Button clicked~~~!")
						},
					},
				},
			},
			&widgets.Center{
				Widget: &widgets.Padding{
					Left:   50,
					Right:  100,
					Widget: CounterButton,
				},
			},
		}},
	})
	app.Run()
}

var CounterButton = goui.StatefulWidgetFunc(
	func(ctx *goui.Context) (state *goui.WidgetState) {
		var data int
		state = &goui.WidgetState{
			Build: func() goui.Widget {
				return &widgets.Button{
					Label: fmt.Sprintf("Clicked %d times", data),
					OnClick: func(ctx *goui.Context) {
						state.Update(func() {
							data++
						})
					},
				}
			}}
		return
	})
