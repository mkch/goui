package goui

import (
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type Window struct {
	ID        ID
	Title     string
	Width     metrics.DP
	Height    metrics.DP
	Root      Widget
	OnClose   func(ctx *Context) bool // Return true to allow closing, false to prevent.
	OnDestroy func(ctx *Context)
}

type window struct {
	Window
	ID         ID
	Handle     native.Handle
	Root       Element       // Root element.
	Layouter   Layouter      // Layouter for the root element.
	DebugLayer native.Handle // Layer for drawing debug outlines.
}
