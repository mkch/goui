package goui

import (
	"github.com/mkch/goui/internal/util"
	"github.com/mkch/goui/native"
)

type Context struct {
	window *window // can't be nil
}

// NativeWindow returns the native window handle associated with this context.
func (ctx *Context) NativeWindow() native.Handle {
	return ctx.window.Handle
}

// CloseWindow closes the window associated with this context.
// If no such window exists, it returns [ErrNoSuchWindow].
func (ctx *Context) CloseWindow() error {
	return appOS.Window_Close(ctx.window.Handle)
}

// WindowEnabled returns whether the window associated with this context is enabled.
func (ctx *Context) WindowEnabled() (bool, error) {
	return appOS.Window_Enabled(ctx.window.Handle)
}

// SetWindowEnabled sets whether the window associated with this context is enabled.
func (ctx *Context) SetWindowEnabled(enabled bool) error {
	return appOS.Window_SetEnabled(ctx.window.Handle, enabled)
}

// updateElementState updates the state of an element by calling f.
// If widgetID is nil, the target is elem itself;
// otherwise, the target is the first element in the element tree rooted at elem that has the given ID.
// If the target element cannot be found or is not a [*StatefulElement], it does nothing and returns nil.
func updateElementState(ctx *Context, elem Element, f func(), widgetID ID) error {
	var stateful *StatefulElement
	if widgetID == nil {
		// try to update element itself
		stateful, _ = elem.(*StatefulElement)
	} else {
		// try to find the stateful element with the given widget ID
		stateful, _ = util.LookupChild(elem, func(e Element) (stateful *StatefulElement, found bool) {
			found = e.Widget().WidgetID() == widgetID
			if found {
				stateful, _ = e.(*StatefulElement)
			}
			return
		})
	}
	if stateful == nil {
		return nil
	}
	return updateWidgetState(ctx, f, stateful)
}

// UpdateWindow updates the state of a [Widget] in the root widget tree of the window
// associated with ctx by calling f.
// If widgetID is nil, the target is the root widget itself;
// otherwise, the target is the first widget in the root widget tree that has the given ID.
// If the target widget cannot be found or is not a [StatefulWidget], it does nothing and returns nil.
func (ctx *Context) UpdateWindow(f func(), widgetID ID) error {
	return updateElementState(ctx, ctx.window.Root, f, widgetID)
}

// UpdateMenu updates the state of a [Menu] or [MenuItem] in the root menu tree of the window
// associated with ctx by calling f.
// If widgetID is nil, the target is the root menu itself;
// otherwise, the target is the first menu or menu item in the root menu tree that has the given ID.
// If the window has no menu, the target cannot be found or is not a stateful component,
// it does nothing and returns nil.
func (ctx *Context) UpdateMenu(f func(), widgetID ID) error {
	if ctx.window.Menu == nil {
		return nil
	}
	return updateElementState(ctx, ctx.window.Menu, f, widgetID)
}
