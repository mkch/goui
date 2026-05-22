package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"
	"image/color"
	"os"

	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/label"
	"github.com/mkch/goui/widgets/listener"
)

func main() {
	os.Exit(goui.Run(ui, &goui.AppConfig{
		Debug: &goui.Debug{
			LayoutOutline: true,
		},
	}))
}

func ui() {
	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		Title: "Listener Sample Disabled Window",
		Width: 1200, Height: 400,
		Disabled:  true,
		OnDestroy: func(ctx *goui.Context) { goui.Exit(0) },
		Root: &widgets.Center{
			Widget: goui.StatefulWidgetFunc(func(ctx *goui.StateContext) goui.WidgetState {
				return &ListenerLabelState{
					StateUpdater: goui.NewStateUpdater(ctx),
				}
			}),
		},
	}))

	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		Title: "Listener Sample window2",
		Width: 1200, Height: 400,
		OnDestroy: func(ctx *goui.Context) { goui.Exit(0) },
		Root: &widgets.Center{
			Widget: goui.StatefulWidgetFunc(func(ctx *goui.StateContext) goui.WidgetState {
				return &ListenerLabelState{
					StateUpdater: goui.NewStateUpdater(ctx),
				}
			}),
		},
	}))
}

type ListenerLabelState struct {
	goui.StateUpdater
	goui.NopDestroyer

	event *listener.PointerEvent
}

func (state *ListenerLabelState) SetEvent(evt *listener.PointerEvent) {
	state.StateUpdater.Update(func() { state.event = evt })
}

// Build implements [goui.WidgetState.Build]
func (state *ListenerLabelState) Build() goui.Widget {
	f := func(ctx *goui.Context, evt *listener.PointerEvent) {
		state.SetEvent(evt)
	}
	var labelText = "Nothing yet."
	if state.event != nil {
		labelText = fmt.Sprintf("%s\nWindowPos: %s", state.event, chkerr.Must(state.event.WindowClientPos()))
	}
	return &widgets.Padding{
		Left: 50, Right: 50, Top: 30, Bottom: 30,
		Widget: &widgets.Listener{
			OnPointerDown: f, OnPointerUp: f, OnPointerMove: f,
			Widget: &widgets.Expanded{
				Widget: &widgets.Column{
					Widgets: []goui.Widget{
						&widgets.Button{
							Label: "Enabled",
						},
						&widgets.Button{
							Label:    "Disabled",
							Disabled: true,
						},
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
