package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"os"

	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui"
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

func ui() {
	chkerr.MustOK(goui.CreateWindow(&goui.Window{
		OnDestroy: func(*goui.Context) { goui.Exit(0) },
		Title:     "goui visibility sample",
		Width:     800,
		Height:    600,
		Root:      rootWidget(),
	}))
}

func rootWidget() goui.Widget {
	return &widgets.Center{Widget: goui.StatefulWidgetFunc(demoWidget)}
}

type state struct {
	goui.StateUpdater
	goui.NopDestroyer
	visible      bool
	maintainSize bool
}

func NewState(ctx *goui.StateContext) *state {
	return &state{
		StateUpdater: goui.NewStateUpdater(ctx),
		visible:      true,
	}
}

func (s *state) Build() goui.Widget {
	return &widgets.Column{
		CrossAxisAlignment: axes.Center,
		Widgets: []goui.Widget{
			&widgets.SizedBox{Height: 20},

			&widgets.Visibility{
				Visible:      s.visible,
				MaintainSize: s.maintainSize,
				Widget: &widgets.Padding{
					Left: 10, Right: 10,
					Widget: &widgets.Label{
						Text: "The quick brown fox jumps over the lazy dog.",
					},
				},
			},

			&widgets.SizedBox{Height: 40},

			&widgets.Button{
				Label: "Show",
				OnClick: func(ctx *goui.Context) {
					if !s.visible {
						chkerr.MustOK(s.Update(func() { s.visible = true }))
					}
				},
			},
			&widgets.Button{
				Label: "Hide",
				OnClick: func(ctx *goui.Context) {
					if s.visible || s.maintainSize {
						chkerr.MustOK(s.Update(func() { s.visible = false; s.maintainSize = false }))
					}
				},
			},
			&widgets.Button{
				Label: "Hide, maintain size",
				OnClick: func(ctx *goui.Context) {
					if s.visible || !s.maintainSize {
						chkerr.MustOK(s.Update(func() { s.visible = false; s.maintainSize = true }))
					}
				},
			},
		},
	}
}

func demoWidget(ctx *goui.StateContext) goui.State {
	return NewState(ctx)
}
