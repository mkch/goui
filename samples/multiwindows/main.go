package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"
	"os"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui"
	"github.com/mkch/goui/messagebox"
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

var window3ID = goui.ValueID("window3")
var window3Closed bool
var mainWindowID = goui.ValueID("main window")
var closeWindow3ButtonID = goui.ValueID("close window3 button")

func ui() {
	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		ID:        mainWindowID,
		OnDestroy: func(*goui.Context) { goui.Exit(0) },
		Title:     "Window 1",
		Width:     800,
		Height:    600,
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
					goui.NewStatefulWidget(closeWindow3ButtonID, func(ctx *goui.StateContext) goui.WidgetState {
						return goui.NewWidgetState(ctx, func() goui.Widget {
							return &widgets.Button{
								Label:    gg.If(window3Closed, "Close Window3 (Already Closed)", "Close Window3"),
								Disabled: window3Closed,
								OnClick: func(ctx *goui.Context) {
									window3 := chkerr.Must(goui.WindowContext(window3ID))
									chkerr.MustOK(window3.CloseWindow())
								},
							}
						}, nil)
					}),
				},
			},
		},
	}))
	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		ID:     window3ID,
		Title:  "Window 3",
		Width:  600,
		Height: 400,
		Root: &widgets.Center{Widget: &widgets.Button{
			Label: "Hello from Window3",
			OnClick: func(ctx *goui.Context) {
				messagebox.Show(ctx, "Window 3", "Hello from Window3!", 0, 0)
			},
		}},
		OnDestroy: func(ctx *goui.Context) {
			mainWindow := chkerr.Must(goui.WindowContext(mainWindowID))
			chkerr.MustOK(mainWindow.UpdateWindow(func() { window3Closed = true }, closeWindow3ButtonID))
		},
	}))
}

func createWindow2() {
	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		Title:  "Window 2",
		Width:  700,
		Height: 600,
		Root:   &widgets.Center{Widget: goui.StatefulWidgetFunc(CountButton)},
		OnClose: func(ctx *goui.Context) bool {
			ret, _ := messagebox.Show(ctx, "Window2", "Are you sure you want to close this window?",
				messagebox.IconQuestion,
				messagebox.ButtonYesNo)
			return ret == messagebox.ReturnYes
		},
	}))
}

func CountButton(ctx *goui.StateContext) goui.WidgetState {
	var count int
	var updater = goui.NewStateUpdater(ctx)
	return goui.NewWidgetState(ctx, func() goui.Widget {
		return &widgets.Button{
			Label: fmt.Sprintf("Clicked %d times", count),
			OnClick: func(ctx *goui.Context) {
				chkerr.MustOK(updater.Update(func() { count++ }))
			},
		}
	}, nil)
}
