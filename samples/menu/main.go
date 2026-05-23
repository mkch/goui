package main

//go:generate rsrc -arch amd64 -manifest manifest.xml
//go:generate rsrc -arch 386 -manifest manifest.xml

import (
	"fmt"
	"os"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui"
	"github.com/mkch/goui/menu"
	"github.com/mkch/goui/messagebox"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/widgets"
	"github.com/mkch/goui/widgets/listener"
)

func main() {
	os.Exit(goui.Run(ui, &goui.AppConfig{
		Debug: &goui.Debug{
			LayoutOutline: true,
		},
	}))
}

var popupCount = 0
var rootWidgetID = goui.UniqueID()

func ui(app *goui.App) {
	chkerr.MustOK(app.CreateWindow(&goui.Window{
		Title:     "menu demo",
		Width:     800,
		Height:    600,
		OnDestroy: func(ctx *goui.Context) { app.Exit(0) },
		Menu:      rootMenu,
		Root:      goui.NewStatefulWidget(rootWidgetID, rootWidget),
	}))
}

var showHelp = true
var screenCoord = false // Use screen coordinates for menu popup
var useListener = true

func rootWidget(ctx *goui.StateContext) goui.WidgetState {
	return goui.NewWidgetState(ctx, func() goui.Widget {
		if !useListener {
			return &widgets.Center{
				Widget: &widgets.Label{Text: "Empty"},
			}
		}
		return &widgets.Listener{
			OnPointerUp: func(ctx *goui.Context, event *listener.PointerEvent) {
				var spec *menu.PopupSpec
				if screenCoord {
					spec = &menu.PopupSpec{Pos: metrics.Point{X: 10, Y: 20}}
				}
				if event.Button == listener.SecondaryMouseButton {
					err := menu.Popup(ctx, &menu.Menu{Items: []goui.MenuItem{
						&menu.Item{Title: "Item1"},
						&menu.Item{
							Title:    fmt.Sprintf("Popup Count: %d", popupCount),
							OnSelect: func(ctx *goui.Context) { popupCount++ },
						},
						&menu.Item{Title: "Hello Submenu",
							Submenu: &menu.Menu{Items: []goui.MenuItem{
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
		}
	}, nil)
}

var rootMenu = menu.StatefulMenuFunc(
	func(ctx *goui.StateContext) (state menu.MenuState) {
		state = menu.NewMenuState(ctx, func() goui.Menu {
			m := &menu.WindowMenu{
				Items: []goui.MenuItem{
					&menu.Item{
						Title: "File",
						Submenu: menu.Items{
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
							&menu.Item{
								Title: gg.If(useListener, "Use empty root", "Use listener root"),
								OnSelect: func(ctx *goui.Context) {
									state.Update(func() { useListener = !useListener })
									ctx.UpdateWindow(func() {}, rootWidgetID)
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
)

type CountState struct {
	goui.StateUpdater
	goui.NopDestroyer
	count int
}

func (s *CountState) Increment() {
	s.Update(func() { s.count++ })
}

func (s *CountState) Build() goui.MenuItem {
	return &menu.Item{
		Title: "Counter item inside",
		Submenu: menu.Items{
			&menu.Item{
				Title: fmt.Sprintf("Count: %d", s.count),
				OnSelect: func(ctx *goui.Context) {
					s.Increment()
				},
			},
		},
	}
}

var CounterItem = menu.NewStatefulItem(
	nil,
	func(ctx *goui.StateContext) menu.ItemState {
		return &CountState{
			StateUpdater: goui.NewStateUpdater(ctx),
		}
	},
)
