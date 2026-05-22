package goui_test

import (
	"fmt"

	"github.com/mkch/goui"
	"github.com/mkch/goui/menu"
	"github.com/mkch/goui/widgets"
)

func ExampleContext_UpdateWindow() {
	var count int
	var rootWidgetID = goui.UniqueID()

	goui.CreateWindow(&goui.Window{
		// Assign rootWidgetID to the root widget of the window.
		Root: goui.NewStatefulWidget(rootWidgetID, func(ctx *goui.StateContext) goui.WidgetState {
			return goui.NewWidgetState(ctx, func() goui.Widget {
				return &widgets.Label{Text: fmt.Sprintf("Clicked %d times", count)}
			}, nil)
		}),

		Menu: &menu.WindowMenuItems{
			&menu.Item{
				Title: "Increment Count",
				OnSelect: func(ctx *goui.Context) {
					// Update the root widget by rootWidgetID.
					ctx.UpdateWindow(func() { count++ }, rootWidgetID)
				},
			},
		},
	})
}

func ExampleContext_UpdateMenu() {
	var count int
	var menuItemID = goui.ValueID(1)

	goui.CreateWindow(&goui.Window{
		Menu: &menu.WindowMenuItems{
			// Assign menuItemID to the menu item.
			menu.NewStatefulItem(menuItemID, func(ctx *goui.StateContext) menu.ItemState {
				return menu.NewItemState(ctx, func() goui.MenuItem {
					return &menu.Item{Title: fmt.Sprintf("Clicked %d times", count)}
				}, nil)
			}),
		},

		Root: &widgets.Button{
			Label: "Increment Count",
			OnClick: func(ctx *goui.Context) {
				// Update the menu item by menuItemID.
				ctx.UpdateMenu(func() { count++ }, menuItemID)
			},
		},
	})
}
