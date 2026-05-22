package goui

import (
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type Window struct {
	ID       ID
	Title    string
	Width    metrics.DP
	Height   metrics.DP
	Disabled bool
	// The root widget to display in the window.
	// If Root is nil, an empty window is created.
	Root Widget
	// The menu widget for the window.
	// If nil, no menu is created.
	// Only widgets representing menus should be assigned; others will not display correctly.
	Menu Menu
	// OnClose is called when the window is requested to close, if not nil.
	// If OnClose is not nil and returns false, the window will not close.
	OnClose   func(ctx *Context) bool
	OnDestroy func(ctx *Context)
}

type window struct {
	ID         ID
	Handle     native.Handle
	Root       Element       // Root element.
	Menu       Element       // Menu element.
	Layouter   Layouter      // Layouter for the root element.
	DebugLayer native.Handle // Layer for drawing debug outlines.
}
