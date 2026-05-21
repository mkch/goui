package menu

import (
	"github.com/mkch/gg/errortrace/chkerr"
	"github.com/mkch/goui"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type PopupSpec struct {
	// Position where the track popup should appear.
	// In screen coordinates.
	Pos metrics.Point
}

// Popup displays the given menu as a popup menu.
// If spec is nil, the default settings are used.
// If Popup is called from [listener.OnPointerUp] with [listener.SecondaryMouseButton],
// Popup may fail with golang.org/x/sys/windows.ERROR_POPUP_ALREADY_ACTIVE error which can be ignored.
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
	defer goui.Post(func() {
		chkerr.MustOK(elem.Destroy(ctx))
	})

	var nativeSpec *native.TrackPopupSpec
	if spec != nil {
		nativeSpec = &native.TrackPopupSpec{
			X: spec.Pos.X,
			Y: spec.Pos.Y,
		}
	}
	return goui.OS().Window_TrackPopupMenu(ctx.NativeWindow(), elem.(*nativeMenuElement).Handle, nativeSpec)
}
