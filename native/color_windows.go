package native

import (
	"image/color"

	"github.com/mkch/gw/win32"
)

// nativeColor converts a color.NRGBA to a native OS color representation.
func nativeColor(c *color.NRGBA) win32.COLORREF {
	return win32.RGB(c.R, c.G, c.B)
}
