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

// CounterButton is a stateful widget that displays a button showing the click count.
// Each time the button is clicked, the click count increases by one.
type CounterButton struct {
	goui.StatefulWidgetHelper
	goui.ID
}

// NewCounterButton creates a new [CounterButton] with the given ID.
func NewCounterButton(ID goui.ID) *CounterButton {
	return &CounterButton{
		ID: ID,
	}
}

// CreateState implements [goui.StatefulWidget.CreateState] method.
func (*CounterButton) CreateState(ctx *goui.StateContext) goui.State {
	return NewCountState(ctx)
}

// ID implements [goui.Widget.ID] method.
func (b *CounterButton) WidgetID() goui.ID {
	return b.ID
}

func ExampleStatefulWidget() {
	var counterButton goui.StatefulWidget = NewCounterButton(goui.ValueID("id"))
	// Or simply use StatefulWidgetFunc:
	counterButton = goui.NewStatefulWidget(goui.ValueID("id"), func(ctx *goui.StateContext) goui.State {
		return NewCountState(ctx)
	})
	// Or StatefulWidgetFunc if ID is nil:
	counterButton = goui.StatefulWidgetFunc(func(sc *goui.StateContext) goui.State {
		return NewCountState(sc)
	})

	_ = counterButton
}
