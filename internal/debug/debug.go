// Package debug provides debugging utilities.
package debug

import (
	_ "unsafe" // for go:linkname

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/tricks"
)

//go:linkname debug

// debug returns the debug configuration for the given context.
func debug(ctx *goui.Context) *tricks.Debug

// CheckLayoutOverflow returns an [goui.OverflowConstraintsError] if the given size exceeds the given constraints.
// Widget can be nil and if widget is not nil, it is included in the error for better debugging.
// This function is intended to be used when
func CheckLayoutOverflow(ctx *goui.Context, widget goui.Widget, size goui.Size, constraints goui.Constraints) error {
	if debug(ctx) == nil {
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
