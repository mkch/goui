package layoututil

import (
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
)

// CheckOverflow returns an [goui.OverflowParentError] if the given size exceeds the given constraints.
// Widget can be nil and if widget is not nil, it is included in the error for better debugging.
func CheckOverflow(widget goui.Widget, size goui.Size, constraints goui.Constraints) error {
	if size.Width < constraints.MinWidth || size.Width > constraints.MaxWidth ||
		size.Height < constraints.MinHeight || size.Height > constraints.MaxHeight {
		return errortrace.WithStack(&goui.OverflowParentError{
			Widget:      widget,
			Size:        size,
			Constraints: constraints,
		})
	}
	return nil
}

// Clamp clamps value between min and max.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampSize clamps the given size between the given constraints.
func ClampSize(size goui.Size, constraints goui.Constraints) goui.Size {
	return goui.Size{
		Width:  Clamp(size.Width, constraints.MinWidth, constraints.MaxWidth),
		Height: Clamp(size.Height, constraints.MinHeight, constraints.MaxHeight),
	}
}
