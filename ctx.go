package goui

import "github.com/mkch/goui/native"

type Context struct {
	window *window // can't be nil
}

// NativeWindow returns the native window handle associated with this context.
func (ctx *Context) NativeWindow() native.Handle {
	return ctx.window.Handle
}

// WindowEnabled returns whether the window associated with this context is enabled.
func (ctx *Context) WindowEnabled() (bool, error) {
	return appOS.Window_Enabled(ctx.window.Handle)
}

// SetWindowEnabled sets whether the window associated with this context is enabled.
func (ctx *Context) SetWindowEnabled(enabled bool) error {
	return appOS.Window_SetEnabled(ctx.window.Handle, enabled)
}
