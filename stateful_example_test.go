package goui_test

import (
	"fmt"

	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
)

// CountState is a [goui.State] that holds the click count.
type CountState struct {
	goui.StateUpdater // Implements Update method.
	goui.NopDestroyer // No cleanup needed.

	count int // The actual state.
}

// NewCountState creates a new [CountState].
func NewCountState(ctx *goui.StateContext) goui.State {
	return &CountState{
		StateUpdater: goui.NewStateUpdater(ctx),
	}
}

// Build implements [goui.State.Build] method.
func (state *CountState) Build() goui.Widget {
	return &widgets.Button{
		Label: fmt.Sprintf("Clicked %d times", state.count),
		OnClick: func(ctx *goui.Context) {
			state.Update(func() {
				state.count++
			})
		},
	}
}

// NewCounterButton creates a new counter button with the given ID.
// The returned counter button is a stateful widget that displays a button showing the click count.
// Each time the button is clicked, the click count increases by one.
func NewCounterButton(ID goui.ID) goui.StatefulWidget {
	return goui.NewStatefulWidget(
		ID,
		func(ctx *goui.StateContext) goui.State { return NewCountState(ctx) },
	)
}

func ExampleStatefulWidget() {
	var counterButton goui.StatefulWidget = NewCounterButton(goui.ValueID("id"))

	// If there is no need to set an id, StatefulWidgetFunc is very convenient.
	counterButton = goui.StatefulWidgetFunc(NewCountState)

	_ = counterButton
}
