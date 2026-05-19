// Package debug provides debugging utilities.
package debug

import (
	_ "unsafe" // for go:linkname

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/metrics"
)

//go:linkname debugEnabled

// debugEnabled returns whether debug mode is enabled.
// It is implemented in goui/link.go and is linked here to avoid import cycle.
func debugEnabled() bool

// CheckLayoutOverflow returns an [goui.OverflowConstraintsError] if the given size exceeds the given constraints and debug mode is on.
// Widget can be nil and if widget is not nil, it is included in the error for better debugging.
// It is recommended to call this function in the Layout method of a custom layouter to detect children layout overflow issues.
func CheckLayoutOverflow(ctx *goui.Context, widget goui.Component, size metrics.Size, constraints metrics.Constraints) error {
	if !debugEnabled() {
		return nil
	}
	if size.Width < constraints.MinWidth || size.Width > constraints.MaxWidth ||
		size.Height < constraints.MinHeight || size.Height > constraints.MaxHeight {
		return errortrace.WithStack(&goui.OverflowConstraintsError{
			Widget:      widget,
			Size:        size,
			Constraints: constraints,
		})
	}
	return nil
}
