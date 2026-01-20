package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"

	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
)

func main() {
	goui.RunAndExit(ui, &goui.AppConfig{
		Debug: &goui.Debug{
			LayoutOutline: true,
		},
	})
}

func ui() {
	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		OnDestroy: func(*goui.Context) { goui.Exit(0) },
		Title:     "goui demo",
		Width:     1200,
		Height:    800,
		Root: &widgets.Column{Widgets: []goui.Widget{
			&widgets.Center{
				HeightFactor: 1.2,
				Widget: &widgets.SizedBox{
					Width: 200, Height: 50,
					Widget: goui.StatelessWidgetFunc(ClickMeButton),
				},
			},
			&widgets.Center{
				HeightFactor: 1.2,
				Widget: &widgets.SizedBox{
					Width: 600, Height: 50,
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
					Left:   100,
					Right:  200,
					Widget: goui.StatefulWidgetFunc(counterButton),
				},
			},
		}},
	}))
}

func ClickMeButton(ctx *goui.Context) goui.Widget {
	return &widgets.Button{
		Label: "Click me!",
		OnClick: func(ctx *goui.Context) {
			fmt.Println("Click me!")
		},
	}
}

func counterButton(ctx *goui.StateContext) (state goui.WidgetState) {
	// Click count, the real state.
	var count int
	state = goui.NewWidgetState(ctx, func() goui.Widget {
		return &widgets.Button{
			Label: fmt.Sprintf("Clicked %d times", count),
			OnClick: func(ctx *goui.Context) {
				chkerr.MustOK(state.Update(func() { count++ }))
			},
		}
	}, nil)
	return
}
