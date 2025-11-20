package goui

import (
	"fmt"
	"iter"
	"slices"
	"time"
	"unsafe"

	"github.com/mkch/goui/native"
)

// OverflowParentError is returned when a widget's size exceeds its parent's constraints.
type OverflowParentError struct {
	Widget      Widget
	Size        Size
	Constraints Constraints
}

func (e *OverflowParentError) Error() string {
	return fmt.Sprintf("widget %T (ID = %v) with size %s overflows its parent constraints %s",
		e.Widget, e.Widget.WidgetID(), &e.Size, &e.Constraints)
}

// Infinity represents an infinite size(unbounded) constraint.
const Infinity = 1<<(unsafe.Sizeof(int(0))*8-1) - 1

// Constraints represents layout constraints.
type Constraints struct {
	MinWidth  int
	MinHeight int
	MaxWidth  int
	MaxHeight int
}

func (c *Constraints) String() string {
	return fmt.Sprintf("{MinWidth: %d, MinHeight: %d, MaxWidth: %d, MaxHeight: %d}",
		c.MinWidth, c.MinHeight, c.MaxWidth, c.MaxHeight)
}

// TightWidth returns true if the constraint has a finite and equal min and max width.
func (c *Constraints) TightWidth() bool {
	return c.MinWidth == c.MaxWidth && c.MinWidth != Infinity
}

// TightHeight returns true if the constraint has a finite and equal min and max height.
func (c *Constraints) TightHeight() bool {
	return c.MinHeight == c.MaxHeight && c.MinHeight != Infinity
}

// UnboundWidth returns true if no constraint is imposed on width.
func (c *Constraints) UnboundWidth() bool {
	return c.MaxWidth == Infinity
}

// UnboundHeight returns true if no constraint is imposed on height.
func (c *Constraints) UnboundHeight() bool {
	return c.MaxHeight == Infinity
}

type Size struct {
	Width  int
	Height int
}

func (s *Size) String() string {
	return fmt.Sprintf("{Width: %d, Height: %d}", s.Width, s.Height)
}

type Point struct {
	X, Y int
}

type Rect struct {
	Left, Top, Right, Bottom int
}

func (r *Rect) Width() int {
	return r.Right - r.Left
}

func (r *Rect) Height() int {
	return r.Bottom - r.Top
}

func (r *Rect) TopLeft() Point {
	return Point{X: r.Left, Y: r.Top}
}

func (r *Rect) BottomRight() Point {
	return Point{X: r.Right, Y: r.Bottom}
}

// Layouter is the interface for laying out elements.
type Layouter interface {
	// Layout computes the size of the element given the constraints.
	Layout(ctx *Context, constraints Constraints) (Size, error)
	// PositionAt puts the element at the given position.
	PositionAt(x, y int) error
	// Replayer returns a function that can replay the last layout operations,
	// or nil if replay is not supported (e.g., when the layout depends on children).
	Replayer() func(*Context) error
	NumChildren() int
	Child(n int) Layouter

	indexChildFunc(f func(Layouter) bool) int
	removeChild(child Layouter)
	removeChildIndex(n int)
	removeChildrenRange(start, end int)
	parent() Layouter
	setParent(parent Layouter)
	Element() Element
	setElement(e Element)

	// appendChildToSlice is a helper of [Layouter_AppendChild].
	// The implementation should just append child to the children slice or some equivalent.
	appendChildToSlice(child Layouter)
	// setChildInSlice is a helper of [Layouter_SetChild].
	// The implementation should just set child at index n in the children slice or some equivalent.
	setChildInSlice(i int, child Layouter)
}

// LayouterBase is a helper struct for implementing Layouter.
// Embedding LayouterBase in a struct and implementing
// Layout and Apply methods implements the Layouter interface.
type LayouterBase struct {
	theElement Element
	theParent  Layouter
	children   []Layouter
}

func (l *LayouterBase) Element() Element {
	return l.theElement
}

func (l *LayouterBase) setElement(e Element) {
	l.theElement = e
}

func (l *LayouterBase) parent() Layouter {
	return l.theParent
}

func (l *LayouterBase) NumChildren() int {
	return len(l.children)
}

func (l *LayouterBase) Child(n int) Layouter {
	return l.children[n]
}

func (l *LayouterBase) indexChildFunc(f func(Layouter) bool) int {
	return slices.IndexFunc(l.children, f)
}

func (l *LayouterBase) removeChildIndex(n int) {
	l.children = slices.Delete(l.children, n, n+1)
}

func (l *LayouterBase) removeChild(child Layouter) {
	l.children = slices.DeleteFunc(l.children, func(l Layouter) bool { return l == child })
}

func (l *LayouterBase) removeChildrenRange(start, end int) {
	l.children = slices.Delete(l.children, start, end)
}

func (l *LayouterBase) setChildInSlice(i int, child Layouter) {
	l.children[i] = child
}

func (l *LayouterBase) setParent(parent Layouter) {
	l.theParent = parent
}

func (l *LayouterBase) appendChildToSlice(child Layouter) {
	l.children = append(l.children, child)
}

func (l *LayouterBase) Replayer() func(*Context) error {
	return nil
}

// layouter_AppendChild appends child to parent Layouter.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func layouter_AppendChild(parent, child Layouter) {
	child.setParent(parent)
	parent.appendChildToSlice(child)
}

// layouter_SetChild sets child at index n of parent Layouter.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func layouter_SetChild(parent Layouter, n int, child Layouter) {
	if parent.Child(n) == child {
		return
	}
	parent.setChildInSlice(n, child)
	child.setParent(parent)
}

// debugLayouterVer records a debug layouter and its highlight version.
type debugLayouterVer struct {
	Layouter *debugLayouter
	Version  uintptr
}

// debugLayouter is a [Layouter] wrapper that records debugging information.
type debugLayouter struct {
	Layouter
	Ctx                  *Context
	Size                 Size                // Last computed size
	Pos                  Point               // Last computed position
	Highlight            bool                // Whether to highlight the outline of this layouter
	HighlightVer         uintptr             // Version of the highlight, used to avoid redundant redraws
	CancelHighlightBatch *[]debugLayouterVer // Batch of layouters to cancel highlight together
}

func (l *debugLayouter) Layout(ctx *Context, constraints Constraints) (size Size, err error) {
	l.Highlight = true // Mark to highlight
	l.HighlightVer++

	if debugParent, ok := l.parent().(*debugLayouter); ok && // parent is debug layouter but can be nil
		debugParent.CancelHighlightBatch != nil && *debugParent.CancelHighlightBatch != nil {
		// Inherit and join the cancel highlight batch from parent
		l.CancelHighlightBatch = debugParent.CancelHighlightBatch
		*l.CancelHighlightBatch = append(*l.CancelHighlightBatch, debugLayouterVer{Layouter: l, Version: l.HighlightVer})
	} else {
		// This is the root of laying out
		l.CancelHighlightBatch = &[]debugLayouterVer{{Layouter: l, Version: l.HighlightVer}}
		defer func() {
			if err != nil {
				return // do not show highlight if layout fails
			}
			// Show highlight after laying out(include children) is done
			native.InvalidWindow(l.Ctx.window.Handle)
			// Schedule canceling all highlights in the batch after a delay
			const delay = 100 * time.Millisecond
			batch := *l.CancelHighlightBatch
			*l.CancelHighlightBatch = nil
			time.AfterFunc(delay, func() {
				l.Ctx.app.Post(func() {
					// Cancel all highlights in a batch
					var cancelled bool
					for _, record := range batch {
						if record.Version < record.Layouter.HighlightVer {
							continue // too late, already updated
						}
						record.Layouter.Highlight = false
						record.Layouter.HighlightVer = 0
						cancelled = true
					}
					// Request a redraw to remove the highlights
					if cancelled {
						native.InvalidWindow(l.Ctx.window.Handle)
					}
				})
			})
		}()
	}

	size, err = l.Layouter.Layout(ctx, constraints)
	if err != nil {
		return
	}
	l.Size = size // Record size
	return
}

func (l *debugLayouter) PositionAt(x, y int) (err error) {
	err = l.Layouter.PositionAt(x, y)
	if err != nil {
		return
	}
	l.Pos = Point{X: x, Y: y} // Record position
	return
}

// allLayouterDebugOutlines returns an iterator of debug rectangles for the given layouter tree.
// The tree must be built with debugging([Window.DebugLayout]) on.
func allLayouterDebugOutlines(root Layouter) iter.Seq[native.DebugRect] {
	return func(yield func(native.DebugRect) bool) {
		// Use a stack to avoid recursive iterator calls
		stack := []Layouter{root}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if debugLayouter, ok := current.(*debugLayouter); ok {
				if !yield(native.DebugRect{
					Left:      debugLayouter.Pos.X,
					Top:       debugLayouter.Pos.Y,
					Right:     debugLayouter.Pos.X + debugLayouter.Size.Width,
					Bottom:    debugLayouter.Pos.Y + debugLayouter.Size.Height,
					Highlight: debugLayouter.Highlight}) {
					return
				}
				// Add children to stack in reverse order to maintain left-to-right traversal
				for i := debugLayouter.NumChildren() - 1; i >= 0; i-- {
					stack = append(stack, debugLayouter.Child(i))
				}
			}
		}
	}
}
