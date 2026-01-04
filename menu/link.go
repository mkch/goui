package menu

import (
	_ "unsafe"

	"github.com/mkch/goui" // for go:linkname
)

//go:linkname buildMenuElementTree
func buildMenuElementTree(ctx *goui.Context, w goui.Widget) (goui.Element, error)
