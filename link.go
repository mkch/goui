package goui

// This file contains functions linked from the goui/widgets/widgetstest package.

import _ "unsafe" // for go:linkname

//go:linkname link_BuildElementTree github.com/mkch/goui/widgets/widgetstest.BuildElementTree
func link_BuildElementTree(ctx *Context, widget Widget, parentLayouter Layouter) (Element, Layouter, error) {
	return buildElementTree(ctx, widget, parentLayouter)
}

//go:linkname link_NewMockContext github.com/mkch/goui/widgets/widgetstest.NewContext
func link_NewMockContext() *Context {
	return newMockContext()
}
