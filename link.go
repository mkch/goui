package goui

// This file contains functions linked from the goui/gouitest package.

import (
	_ "unsafe" // for go:linkname

	"github.com/mkch/goui/internal/tricks"
	"github.com/mkch/goui/native"
)

//go:linkname link_BuildElementTree github.com/mkch/goui/gouitest.BuildElementTree
func link_BuildElementTree(ctx *Context, widget AbstractWidget) (Element, Layouter, error) {
	return buildElementTree(ctx, widget)
}

//go:linkname link_RunOS github.com/mkch/goui/gouitest.app_RunOS
func link_RunOS(f func(), config *AppConfig) int {
	return runOS(newOS(), f, config)
}

//go:linkname link_RunContext github.com/mkch/goui/gouitest.app_RunContext
func link_RunContext(os native.OS, f func(ctx *Context), config *AppConfig) int {
	return runContext(os, f, config)
}

//go:linkname link_debug github.com/mkch/goui/internal/debug.debug
func link_debug() *tricks.Debug {
	return appDebug
}
