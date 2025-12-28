package main

import (
	"fmt"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
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
		Root: &widgets.Center{
			Widget: &widgets.Column{
				CrossAxisAlignment: axes.Center,
				Widgets: []goui.Widget{
					&widgets.SizedBox{Height: 10},
					&widgets.SizedBox{
						Width:  200,
						Height: 40,
						Widget: stateful,
					},
					&widgets.SizedBox{Height: 10},
					&widgets.Button{
						Label:   "Increase State (Even: Label, Odd: Button)",
						OnClick: func(ctx *goui.Context) { state.Inc() },
					},
				},
			},
		},
	})
	app.Run()
}

type numberState struct {
	goui.StateUpdater
	goui.NopDestroyer
	number int
}

func NewState(ctx *goui.StateContext) *numberState {
	return &numberState{
		StateUpdater: goui.NewStateUpdater(ctx),
		number:       0,
	}
}

func (s *numberState) Build() goui.Widget {
	return gg.IfFunc(s.number%2 == 0,
		func() goui.Widget { return &widgets.Label{Text: fmt.Sprintf("Label: %v", s.number)} },
		func() goui.Widget { return &widgets.Button{Label: fmt.Sprintf("Button: %v", s.number)} },
	)
}

func (s *numberState) Inc() {
	gg.MustOK(s.Update(func() { s.number++ }))
}

var state *numberState

var stateful = goui.StatefulWidgetFunc(func(ctx *goui.StateContext) goui.State {
	state = NewState(ctx)
	return state
})
