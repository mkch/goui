package goui_test

import (
	"fmt"

	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets"
)

func ExampleNewState() {
	// counterButton is a stateful widget that displays a button showing the click count.
	// Each time the button is clicked, the click count increases by one.
	var counterButton goui.Widget = goui.NewStatefulWidget(
		goui.ValueID(123),
		func(ctx *goui.StateContext) (state goui.State) {
			var count int // Click count, the actual state.
			return goui.NewState(ctx, func() goui.Widget {
				return &widgets.Button{
					Label: fmt.Sprintf("Clicked %d times", count),
					OnClick: func(ctx *goui.Context) {
						state.Update(func() { count++ })
					},
				}
			}, nil)
		},
	)

	_ = counterButton
}
