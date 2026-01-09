package menu_test

import (
	"fmt"

	"github.com/mkch/goui"
	"github.com/mkch/goui/menu"
)

// CountState is a state of selection count.
type CountState struct {
	goui.StateUpdater     // Will create in CountItem function.
	goui.NopDestroyer     // No special cleanup needed.
	count             int // The actual state value, selection count.
}

// Increment increases the count by 1 and triggers a state update.
func (s *CountState) Increment() {
	// Update the state to trigger rebuild.
	s.Update(func() { s.count++ })
}

// Build builds the menu item widget representing and mutating the current state.
func (s *CountState) Build() goui.MenuItem {
	return &menu.Item{
		Title: fmt.Sprintf("Count: %d", s.count),
		OnSelect: func(ctx *goui.Context) {
			s.Increment() // Increment count on selection.
		},
	}
}

// NewCounterItem creates a stateful menu item that tracks how many times it has been selected.
func NewCounterItem() menu.StatefulItem {
	return menu.StatefulItemFunc(
		func(ctx *goui.StateContext) menu.ItemState {
			return &CountState{
				StateUpdater: goui.NewStateUpdater(ctx),
			}
		},
	)
}
func Example_statefulMenuItem() {
	var app *goui.App // assume app is created elsewhere
	_ = app.CreateWindow(&goui.Window{
		Title:  "Stateful Menu Example",
		Width:  400,
		Height: 300,
		Menu: &menu.Menu{
			Items: []goui.MenuItem{&menu.Item{
				Title: "Stateful item inside",
				Submenu: &menu.Menu{
					Items: []goui.MenuItem{
						NewCounterItem(),
					},
				},
			}},
		},
	})
}
