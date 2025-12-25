package goui

import "github.com/mkch/goui/native"

type Window struct {
	ID      ID
	Title   string
	Width   int
	Height  int
	Root    Widget
	OnClose func()
}

type window struct {
	Window
	ID       ID
	Handle   native.Handle
	Root     Element  // Root element.
	Layouter Layouter // Layouter for the root element.
}
