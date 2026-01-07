package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/check"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/axes"
	"github.com/mkch/goui/widgets/label"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	app.CreateWindow(&goui.Window{
		OnDestroy: func(*goui.Context) { app.Exit(0) },
		Title:     "goui demo",
		Width:     800,
		Height:    600,
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
		func() goui.Widget {
			return &widgets.Label{
				TextAlignment: label.Center,
				Text:          fmt.Sprintf("Label: %v", s.number)}
		},
		func() goui.Widget {
			return &widgets.Button{
				Label: fmt.Sprintf("Button: %v", s.number)}
		},
	)
}

func (s *numberState) Inc() {
	check.MustOK(s.Update(func() { s.number++ }))
}

var state *numberState

var stateful = &goui.StatefulWidget{
	StateCreator: func(ctx *goui.StateContext) goui.State {
		state = NewState(ctx)
		return state
	},
}
