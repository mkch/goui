// Package gouitest provides utilities for testing goui applications.
package gouitest

import (
	_ "unsafe" // for go:linkname

	"github.com/mkch/goui"
	"github.com/mkch/goui/native"
	"github.com/mkch/goui/native/mock"
)

//go:linkname BuildElementTree

// BuildElementTree builds the element tree for the given widget.
// Parameter parentLayouter is thee nearest recursive parent layouter,
// or nil if there is no recursive parent layouter.
// If any error occurs during the build, the error is returned.
// The returned Element is the root element of the built tree, and
// the returned Layouter is the layouter of the Element or its nearest child.
func BuildElementTree(ctx *goui.Context, widget goui.Component, parentLayouter goui.Layouter) (goui.Element, goui.Layouter, error)

//go:linkname app_RunOS
func app_RunOS(os native.OS, f func(), config *goui.AppConfig) int

// NewContext creates and returns a new mock goui.Context for testing.
// No OS specific resources are allocated.
func Run(f func(), config *goui.AppConfig) int {
	return app_RunOS(mock.NewOS(), f, config)
}

//go:linkname app_RunContext
func app_RunContext(os native.OS, f func(ctx *goui.Context), config *goui.AppConfig) int

func RunContext(f func(ctx *goui.Context), config *goui.AppConfig) int {
	return app_RunContext(mock.NewOS(), f, config)
}
