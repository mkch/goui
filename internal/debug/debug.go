// Package debug provides debugging utilities.
package debug

import (
	_ "unsafe" // for go:linkname

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/tricks"
	"github.com/mkch/goui/metrics"
)

//go:linkname debug

// debug returns the debug configuration for the given context.
func debug() *tricks.Debug

// CheckLayoutOverflow returns an [goui.OverflowConstraintsError] if the given size exceeds the given constraints.
// Widget can be nil and if widget is not nil, it is included in the error for better debugging.
func CheckLayoutOverflow(ctx *goui.Context, widget goui.Component, size metrics.Size, constraints metrics.Constraints) error {
	if debug() == nil {
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
