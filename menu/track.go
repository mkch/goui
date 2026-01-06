package menu

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/check"
	"github.com/mkch/goui/native"
)

type PopupSpec struct {
	// Position where the track popup should appear.
	// In screen coordinates.
	Pos goui.Point
}

// Popup displays the given menu as a popup menu.
// If spec is nil, the default settings are used.
func Popup(ctx *goui.Context, menu *Menu, spec *PopupSpec) (err error) {
	elem, err := goui.BuildElementTree(ctx, menu)
	if err != nil {
		return
	}
	// Do not use
	//
	//  defer elem.Destroy()
	//
	// here.
	// because the window does not receive a WM_COMMAND message from the menu until
	// TrackPopupMenu(win32.TrackPopupMenuEx) returns.
	// If the menu is destroyed before that, the menu commands will not be delivered
	// to the menu item callback.
	defer ctx.App().Post(func() {
		check.MustOK(elem.Destroy())
	})

	var nativeSpec *native.TrackPopupSpec
	if spec != nil {
		nativeSpec = &native.TrackPopupSpec{
			X: spec.Pos.X,
			Y: spec.Pos.Y,
		}
	}
	return native.TrackPopupMenu(elem.(*nativeMenuElement).Handle, ctx.NativeWindow(), nativeSpec)
}
