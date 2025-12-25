package layoututil

import (
	"github.com/mkch/goui"
)

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
