package goui

import (
	"fmt"
	"iter"
	"time"
	"unsafe"

	"github.com/mkch/goui/native"
)

// OverflowConstraintsError is returned when a widget's size exceeds its constraints in debug mode.
// Widget can be nil and if it is not nil, it is included in the error message for better debugging.
type OverflowConstraintsError struct {
	Widget      Widget
	Size        Size
	Constraints Constraints
}

func (e *OverflowConstraintsError) Error() string {
	if e.Widget == nil {
		return fmt.Sprintf("size %s overflows constraints %s", &e.Size, &e.Constraints)
	}
	return fmt.Sprintf("widget %T (ID = %v) with size %s overflows constraints %s",
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

// MinSize returns the minimum size allowed by the constraints.
func (c *Constraints) MinSize() Size {
	return Size{Width: c.MinWidth, Height: c.MinHeight}
}

// MaxSize returns the maximum size allowed by the constraints.
func (c *Constraints) MaxSize() Size {
	return Size{Width: c.MaxWidth, Height: c.MaxHeight}
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
	// Children returns an iterator of child layouters.
	Children() iter.Seq[Layouter]
	// Parent returns the parent layouter, or nil.
	Parent() Layouter
	// Element returns the element that creates this layouter.
	Element() Element

	// setElement is a helper function to set the creator of this layouter.
	// The implementation should just set the element field or some equivalent.
	setElement(element Element)
}

// LayouterBase is a helper struct for implementing Layouter.
// Embedding LayouterBase in a struct and implementing
// Layout and PositionAt methods implements the Layouter interface.
type LayouterBase struct {
	element Element
}

func (l *LayouterBase) Element() Element {
	return l.element
}

func (l *LayouterBase) setElement(element Element) {
	l.element = element
}

func (l *LayouterBase) Children() iter.Seq[Layouter] {
	return func(yield func(Layouter) bool) {
		for i := 0; i < l.element.numChildren(); i++ {
			childLayouter := layouterTree(l.element.child(i))
			if childLayouter == nil {
				continue
			}
			if !yield(childLayouter) {
				return
			}
		}
	}
}

func (l *LayouterBase) Parent() (parent Layouter) {
	for element := l.element.parent(); element != nil; element = element.parent() {
		parent = element.Layouter()
		if parent != nil {
			return
		}
	}
	return nil
}

func (l *LayouterBase) Replayer() func(*Context) error {
	return nil
}

// debugLayouterVer records a debug layouter and its highlight version.
type debugLayouterVer struct {
	Layouter *debugLayouter
	Version  uintptr
}

// debugLayouter is a [Layouter] wrapper that records debugging information.
type debugLayouter struct {
	Layouter
	Size                 Size                // Last computed size
	Pos                  Point               // Last computed position
	Highlight            bool                // Whether to highlight the outline of this layouter
	HighlightVer         uintptr             // Version of the highlight, used to avoid redundant redraws
	CancelHighlightBatch *[]debugLayouterVer // Batch of layouters to cancel highlight together
}

func (l *debugLayouter) Layout(ctx *Context, constraints Constraints) (size Size, err error) {
	l.Highlight = true // Mark to highlight
	l.HighlightVer++

	if debugParent, ok := l.Parent().(*debugLayouter); ok && // parent is debug layouter but can be nil
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
			native.InvalidWindow(ctx.window.Handle)
			// Schedule canceling all highlights in the batch after a delay
			const delay = 100 * time.Millisecond
			batch := *l.CancelHighlightBatch
			*l.CancelHighlightBatch = nil
			time.AfterFunc(delay, func() {
				ctx.app.Post(func() {
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
						native.InvalidWindow(ctx.window.Handle)
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
				// The left-to-right traversal order for children is not maintained here.
				// Reversing debugLayouter.Children() would be inefficient.
				for child := range debugLayouter.Children() {
					stack = append(stack, child)
				}
			}
		}
	}
}

// layouterTree returns the layouter tree for the given element tree.
// The returned layouter is the layouter of the given element or its nearest child.
func layouterTree(element Element) (layouter Layouter) {
	layouter = element.Layouter()
	if layouter != nil {
		return
	}
	if _, isContainer := element.Widget().(Container); isContainer {
		panic("container without a layouter")
	}
	if element.numChildren() == 0 {
		return nil
	}
	return layouterTree(element.child(0))
}
