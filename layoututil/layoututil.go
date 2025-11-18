package layoututil

import "github.com/mkch/goui"

// CheckOverflow returns an [goui.OverflowParentError] if the given size exceeds the given constraints.
func CheckOverflow(widget goui.Widget, size goui.Size, constraints goui.Constraints) error {
	if size.Width < constraints.MinWidth || size.Width > constraints.MaxWidth ||
		size.Height < constraints.MinHeight || size.Height > constraints.MaxHeight {
		return &goui.OverflowParentError{
			Widget:      widget,
			Size:        size,
			Constraints: constraints,
		}
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
