package goui

import "github.com/mkch/goui/native"

type Context struct {
	window *window // can't be nil
}

// NativeWindow returns the native window handle associated with this context.
func (ctx *Context) NativeWindow() native.Handle {
	return ctx.window.Handle
}
