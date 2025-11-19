// Package widgettest provides utilities for testing widgets.
package widgetstest

import (
	_ "unsafe" // for go:linkname

	"github.com/mkch/goui"
)

//go:linkname BuildElementTree

// BuildElementTree builds the element tree for the given widget.
// Parameter parentLayouter is thee nearest recursive parent layouter,
// or nil if there is no recursive parent layouter.
// If any error occurs during the build, the error is returned.
// The returned Element is the root element of the built tree, and
// the returned Layouter is the layouter of the Element or its nearest child.
func BuildElementTree(ctx *goui.Context, widget goui.Widget, parentLayouter goui.Layouter) (goui.Element, goui.Layouter, error)

//go:linkname NewContext

// NewContext creates and returns a new mock goui.Context for testing.
// No OS specific resources are allocated.
func NewContext() *goui.Context
