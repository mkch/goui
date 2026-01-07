package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"
	"os"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/check"
	"github.com/mkch/goui/menu"
	"github.com/mkch/goui/messagebox"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/listener"
)

var app = goui.NewApp(&goui.AppConfig{
	Debug: &goui.Debug{
		LayoutOutline: true,
	},
})

var popupCount = 0

func main() {
	check.MustOK(app.CreateWindow(&goui.Window{
		Title:     "menu demo",
		Width:     800,
		Height:    600,
		OnDestroy: func(ctx *goui.Context) { app.Exit(0) },
		Menu:      MainMenu,
		Root: &widgets.Listener{
			OnPointerUp: func(ctx *goui.Context, event *listener.PointerEvent) {
				var spec *menu.PopupSpec
				if screenCoord {
					spec = &menu.PopupSpec{Pos: goui.Point{X: 10, Y: 20}}
				}
				if event.Button&listener.SecondaryMouseButton != 0 {
					err := menu.Popup(ctx, &menu.Menu{Items: []goui.Widget{
						&menu.Item{Title: "Item1"},
						&menu.Item{
							Title:    fmt.Sprintf("Popup Count: %d", popupCount),
							OnSelect: func(ctx *goui.Context) { popupCount++ },
						},
						&menu.Item{Title: "Hello Submenu",
							Submenu: &menu.Menu{Items: []goui.Widget{
								&menu.Item{
									Title: "Hello!",
									OnSelect: func(ctx *goui.Context) {
										messagebox.Show(ctx, "Hello", "Hello, popup!", messagebox.IconInfo, messagebox.ButtonOK)
									},
								},
							}},
						},
					}}, spec)
					if err != nil {
						errortrace.Panic(err)
					}
				}
			},
			Widget: &widgets.Expanded{},
		},
	}))
	os.Exit(app.Run())
}

var showHelp = true
var screenCoord = false // Use screen coordinates for menu popup

var MainMenu = &goui.StatefulWidget{
	StateCreator: func(ctx *goui.StateContext) (state goui.State) {
		state = goui.NewState(ctx, func() goui.Widget {
			m := &menu.WindowMenu{
				Items: []goui.Widget{
					&menu.Item{
						Title: "File",
						Submenu: &menu.Menu{
							Items: []goui.Widget{
								&menu.Item{
									Title:    "New",
									OnSelect: func(ctx *goui.Context) { fmt.Println("New selected") },
								},
								&menu.Separator{},
								CounterItem,
								&menu.Item{
									Title:    gg.If(showHelp, "Hide Help", "Show Help"),
									OnSelect: func(ctx *goui.Context) { state.Update(func() { showHelp = !showHelp }) },
								},
								&menu.Item{
									Title:    gg.If(screenCoord, "Popup at mouse pointer", "Popup at (10,20) on screen"),
									OnSelect: func(ctx *goui.Context) { state.Update(func() { screenCoord = !screenCoord }) },
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
	},
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

var CounterItem = &goui.StatefulWidget{
	StateCreator: func(ctx *goui.StateContext) goui.State {
		return &CountState{
			StateUpdater: goui.NewStateUpdater(ctx),
		}
	},
}
