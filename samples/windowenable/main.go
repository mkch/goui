package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"github.com/mkch/gg"
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

func safeWindowEnabled(id goui.ID) bool {
	win, err := goui.WindowContext(id)
	if err != nil {
		return true
	}
	enabled, err := win.WindowEnabled()
	if err != nil {
		return true
	}
	return enabled
}

func ui() {
	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		ID:        goui.ValueID("window1"),
		Title:     "Window 1",
		Width:     400,
		Height:    300,
		OnDestroy: func(ctx *goui.Context) { goui.Exit(0) },
		Root: &widgets.Center{
			Widget: goui.StatefulWidgetFunc(func(ctx *goui.StateContext) (state goui.WidgetState) {
				state = goui.NewWidgetState(ctx, func() goui.Widget {
					return &widgets.Button{
						Label: gg.If(safeWindowEnabled(goui.ValueID("window2")), "Disable Window 2", "Enable Window 2"),
						OnClick: func(ctx *goui.Context) {
							window2 := chkerr.Must(goui.WindowContext(goui.ValueID("window2")))
							enabled := chkerr.Must(window2.WindowEnabled())
							chkerr.MustOK(window2.SetWindowEnabled(!enabled))
							state.Update(func() {})
						},
					}
				}, nil)
				return
			}),
		},
	}))

	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		ID:        goui.ValueID("window2"),
		Title:     "Window 2",
		Width:     400,
		Height:    300,
		OnDestroy: func(ctx *goui.Context) { goui.Exit(0) },
		Root: &widgets.Center{
			Widget: goui.StatefulWidgetFunc(func(ctx *goui.StateContext) (state goui.WidgetState) {
				state = goui.NewWidgetState(ctx, func() goui.Widget {
					return &widgets.Button{
						Label: gg.If(safeWindowEnabled(goui.ValueID("window1")), "Disable Window 1", "Enable Window 1"),
						OnClick: func(ctx *goui.Context) {
							window1 := chkerr.Must(goui.WindowContext(goui.ValueID("window1")))
							enabled := chkerr.Must(window1.WindowEnabled())
							chkerr.MustOK(window1.SetWindowEnabled(!enabled))
							state.Update(func() {})
						},
					}
				}, nil)
				return
			}),
		},
	}))
}
