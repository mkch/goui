package main

import (
	"fmt"

	"github.com/mkch/goui"
)

var app = goui.NewApp()

func main() {
	app.CreateWindow(goui.Window{
		Title:  "goui demo",
		Width:  600,
		Height: 400,
		Root: &goui.Column{Widgets: []goui.Widget{
			&goui.Button{
				Label: "Click me!",
				OnClick: func() {
					fmt.Println("Button clicked!")
				},
			},
			&goui.Button{
				Label: "Click me!",
				OnClick: func() {
					fmt.Println("Button clicked~~~!")
				},
			},
			CounterButton,
		}},
	})
	app.Run()
}

var CounterButton = goui.StatefulWidgetFuc(
	func(ctx *goui.Context) (state *goui.WidgetState) {
		var data int
		state = &goui.WidgetState{
			Build: func() goui.Widget {
				return &goui.Button{
					Label: fmt.Sprintf("Clicked %d times", data),
					OnClick: func() {
						state.Update(func() {
							data++
						})
					},
				}
			}}
		return
	})
