package main

import (
	"fmt"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/messagebox"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	var window3ID = goui.ValueID("window3")
	closeWindow3Button, enableCloseWindow3Button := newEnableDisableButton("Close Window3", func() {
		app.CloseWindow(window3ID)
	})
	app.CreateWindow(goui.Window{
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "Window 1",
		Width:     600,
		Height:    400,
		Root: &widgets.Center{
			Widget: &widgets.Row{
				MainAxisSize:       axes.Min,
				CrossAxisAlignment: axes.Center,
				Widgets: []goui.Widget{
					&widgets.Button{
						Label: "Create New Window2",
						OnClick: func(ctx *goui.Context) {
							createWindow2()
						},
					},
					&widgets.SizedBox{Width: 20},
					closeWindow3Button,
				},
			},
		},
	})
	createWindow3(window3ID, func(*goui.Context) {
		enableCloseWindow3Button(false)
	})
	app.Run()
}

func newEnableDisableButton(label string, onClick func()) (btn goui.StatefulWidget, setEnabled func(bool)) {
	var enabled bool = true
	var updater goui.StateUpdater
	btn = goui.StatefulWidgetFunc(func(ctx *goui.StateContext) goui.State {
		updater = goui.NewStateUpdater(ctx)
		return goui.NewState(ctx, func() goui.Widget {
			return &widgets.Button{
				Label:    label,
				Disabled: !enabled,
				OnClick: func(ctx *goui.Context) {
					onClick()
				},
			}
		}, nil)
	})
	setEnabled = func(value bool) {
		updater.Update(func() { enabled = value })
	}
	return
}

func createWindow2() {
	app.CreateWindow(goui.Window{
		Title:  "Window 2",
		Width:  400,
		Height: 300,
		Root:   &widgets.Center{Widget: newCountButton()},
		OnClose: func(ctx *goui.Context) bool {
			ret, _ := messagebox.Show(ctx, "Window2", "Are you sure you want to close this window?",
				messagebox.IconQuestion,
				messagebox.ButtonYesNo)
			return ret == messagebox.ReturnYes
		},
	})
}

func newCountButton() goui.StatefulWidget {
	return goui.StatefulWidgetFunc(func(ctx *goui.StateContext) goui.State {
		var count int
		var updater = goui.NewStateUpdater(ctx)
		return goui.NewState(ctx, func() goui.Widget {
			return &widgets.Button{
				Label: fmt.Sprintf("Clicked %d times", count),
				OnClick: func(ctx *goui.Context) {
					gg.MustOK(updater.Update(func() { count++ }))
				},
			}
		}, nil)
	})
}

func createWindow3(id goui.ID, onDestroy func(*goui.Context)) {
	app.CreateWindow(goui.Window{
		ID:     id,
		Title:  "Window 3",
		Width:  400,
		Height: 300,
		Root: &widgets.Center{Widget: &widgets.Button{
			Label: "Hello from Window3",
			OnClick: func(ctx *goui.Context) {
				messagebox.Show(ctx, "Window 3", "Hello from Window3!", 0, 0)
			},
		}},
		OnDestroy: onDestroy,
	})
}
