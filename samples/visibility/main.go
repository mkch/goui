package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/check"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	app.CreateWindow(&goui.Window{
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "goui visibility sample",
		Width:     800,
		Height:    600,
		Root:      rootWidget(),
	})

	app.Run()
}

func rootWidget() goui.Widget {
	return &widgets.Center{Widget: demoWidget()}
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
						check.MustOK(s.Update(func() { s.visible = true }))
					}
				},
			},
			&widgets.Button{
				Label: "Hide",
				OnClick: func(ctx *goui.Context) {
					if s.visible || s.maintainSize {
						check.MustOK(s.Update(func() { s.visible = false; s.maintainSize = false }))
					}
				},
			},
			&widgets.Button{
				Label: "Hide, maintain size",
				OnClick: func(ctx *goui.Context) {
					if s.visible || !s.maintainSize {
						check.MustOK(s.Update(func() { s.visible = false; s.maintainSize = true }))
					}
				},
			},
		},
	}
}

func demoWidget() *goui.Stateful {
	return &goui.Stateful{
		StateCreator: func(ctx *goui.StateContext) goui.State {
			return NewState(ctx)
		},
	}
}
