package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"
	"os"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/menu"
	"github.com/mkch/goui/widgets"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

func main() {
	gg.MustOK(app.CreateWindow(goui.Window{
		Title:     "menu demo",
		Width:     400,
		Height:    300,
		OnDestroy: func(ctx *goui.Context) { app.Exit(0) },
		Menu:      goui.StatefulWidgetFunc(MainMenu),
	}))
	os.Exit(app.Run())
}

var showHelp = true

func MainMenu(ctx *goui.StateContext) (state goui.State) {
	state = goui.NewState(ctx, func() goui.Widget {
		m := &menu.WindowMenu{
			Items: []goui.Widget{
				&widgets.SizedBox{
					Width: 100, Height: 200,
					Widget: &widgets.Button{
						Label: "Wrong!",
					},
				},
				&menu.Item{
					Title: "File",
					Submenu: &menu.Menu{
						Items: []goui.Widget{
							&menu.Item{
								Title:    "New",
								OnSelect: func(ctx *goui.Context) { fmt.Println("New selected") },
							},
							&menu.Separator{},
							goui.StatefulWidgetFunc(CountItem),
							&menu.Item{
								Title:    gg.If(showHelp, "Hide Help", "Show Help"),
								OnSelect: func(ctx *goui.Context) { state.Update(func() { showHelp = !showHelp }) },
							},
						},
					},
				},
			},
		}
		if showHelp {
			m.Items = append(m.Items, &menu.Item{
				Title:    "Help",
				Disabled: true,
			},
			)
		}
		return m
	}, nil)
	return
}

type CountState struct {
	goui.StateUpdater
	goui.NopDestroyer
	count int
}

func (s *CountState) Increment() {
	s.Update(func() { s.count++ })
}

func (s *CountState) Build() goui.Widget {
	return &menu.Item{
		Title: fmt.Sprintf("Count: %d", s.count),
		OnSelect: func(ctx *goui.Context) {
			s.Increment()
		},
	}
}

func CountItem(ctx *goui.StateContext) goui.State {
	return &CountState{
		StateUpdater: goui.NewStateUpdater(ctx),
	}
}
