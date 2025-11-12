package goui

import "github.com/mkch/goui/native"

type Window struct {
	ID     ID
	Title  string
	Width  int
	Height int
	Root   Widget
}

type window struct {
	Window
	ID       ID
	Handle   native.Handle
	Root     Element
	Layouter Layouter
}
