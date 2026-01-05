package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"
	"image/color"
	"os"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/label"
	"github.com/mkch/goui/widgets/listener"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	err := app.CreateWindow(&goui.Window{
		Title: "Listener Sample",
		Width: 1200, Height: 400,
		OnDestroy: func(ctx *goui.Context) { app.Exit(0) },
		Root: &widgets.Center{
			Widget: goui.StatefulWidgetFunc(func(ctx *goui.StateContext) goui.State {
				return &ListenerLabelState{
					StateUpdater: goui.NewStateUpdater(ctx),
				}
			}),
		},
	})

	if err != nil {
		errortrace.Panic(err)
	}

	os.Exit(app.Run())
}

type ListenerLabelState struct {
	goui.StateUpdater
	goui.NopDestroyer

	event *listener.PointerEvent
}

func (state *ListenerLabelState) SetEvent(evt *listener.PointerEvent) {
	state.StateUpdater.Update(func() { state.event = evt })
}

// Build implements [goui.State.Build]
func (state *ListenerLabelState) Build() goui.Widget {
	f := func(ctx *goui.Context, evt *listener.PointerEvent) {
		state.SetEvent(evt)
	}
	var labelText = "Nothing yet."
	if state.event != nil {
		labelText = fmt.Sprintf("%s\nWindowPos: %s", state.event, gg.Must(state.event.WindowClientPos()))
	}
	return &widgets.Padding{
		Left: 50, Right: 50, Top: 30, Bottom: 30,
		Widget: &widgets.Listener{
			OnPointerDown: f, OnPointerUp: f, OnPointerMove: f,
			Widget: &widgets.Expanded{
				Widget: &widgets.Column{
					Widgets: []goui.Widget{
						&widgets.Label{
							Multiline:     true,
							TextAlignment: label.Center,
							Text:          labelText,
						},
						&widgets.Expanded{
							Widget: &widgets.Panel{
								BackgroundColor: &color.NRGBA{B: 255},
							},
						},
					},
				},
			},
		},
	}
}
