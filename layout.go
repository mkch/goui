package goui

import (
	"fmt"
	"iter"
	"slices"
	"time"

	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

// OverflowConstraintsError is returned when a widget's size exceeds its constraints in debug mode.
// Widget can be nil and if it is not nil, it is included in the error message for better debugging.
type OverflowConstraintsError struct {
	Widget      AbstractWidget
	Size        metrics.Size
	Constraints metrics.Constraints
}

func (e *OverflowConstraintsError) Error() string {
	if e.Widget == nil {
		return fmt.Sprintf("size %s overflows constraints %s", &e.Size, &e.Constraints)
	}
	return fmt.Sprintf("widget %T (ID = %v) with size %s overflows constraints %s",
		e.Widget, e.Widget.WidgetID(), &e.Size, &e.Constraints)
}

// Layouter is the interface for laying out elements.
type Layouter interface {
	// Layout computes the size of the element given the constraints.
	Layout(ctx *Context, constraints metrics.Constraints) (metrics.Size, error)
	// PositionAt puts the element at the given position.
	// The position is relative to the native parent element's top-left corner.
	PositionAt(ctx *Context, pt metrics.Point) error
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

// LayouterHelper is a helper struct for implementing Layouter.
// Embedding LayouterHelper in a struct and implementing
// [Layouter.Layout] and [Layouter.PositionAt] methods implements the Layouter interface.
type LayouterHelper struct {
	element Element
}

func (l *LayouterHelper) Element() Element {
	return l.element
}

func (l *LayouterHelper) setElement(element Element) {
	l.element = element
}

func (l *LayouterHelper) Children() iter.Seq[Layouter] {
	return func(yield func(Layouter) bool) {
		for i := 0; i < l.element.NumChildren(); i++ {
			childLayouter := layouterTree(l.element.Child(i))
			if childLayouter == nil {
				continue
			}
			if !yield(childLayouter) {
				return
			}
		}
	}
}

func (l *LayouterHelper) Parent() (parent Layouter) {
	for element := l.element.Parent(); element != nil; element = element.Parent() {
		parent = element.Layouter()
		if parent != nil {
			return
		}
	}
	return nil
}

func (l *LayouterHelper) Replayer() func(*Context) error {
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
	Size                 metrics.Size        // Last computed size
	Pos                  metrics.Point       // Last computed position
	Highlight            bool                // Whether to highlight the outline of this layouter
	highlightVer         uintptr             // Version of the highlight, used to avoid redundant redraws
	cancelHighlightBatch *[]debugLayouterVer // Batch of layouters to cancel highlight together
}

func (l *debugLayouter) Layout(ctx *Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	l.Highlight = true // Mark to highlight
	l.highlightVer++

	if debugParent, ok := l.Parent().(*debugLayouter); ok && // parent is debug layouter but can be nil
		debugParent.cancelHighlightBatch != nil && *debugParent.cancelHighlightBatch != nil {
		// Inherit and join the cancel highlight batch from parent
		l.cancelHighlightBatch = debugParent.cancelHighlightBatch
		*l.cancelHighlightBatch = append(*l.cancelHighlightBatch, debugLayouterVer{Layouter: l, Version: l.highlightVer})
	} else {
		// This is the root of laying out
		l.cancelHighlightBatch = &[]debugLayouterVer{{Layouter: l, Version: l.highlightVer}}
		defer func() {
			if err != nil {
				return // do not show highlight if layout fails
			}
			// Show highlight after laying out(include children) is done
			appOS.Window_Invalidate(ctx.window.DebugLayer, nil)
			// Schedule canceling all highlights in the batch after a delay
			const delay = 100 * time.Millisecond
			batch := *l.cancelHighlightBatch
			*l.cancelHighlightBatch = nil
			time.AfterFunc(delay, func() {
				appOS.App_Post(func() {
					// Cancel all highlights in a batch
					var cancelled bool
					for _, record := range batch {
						if record.Version < record.Layouter.highlightVer {
							continue // too late, already updated
						}
						record.Layouter.Highlight = false
						record.Layouter.highlightVer = 0
						cancelled = true
					}
					// Request a redraw to remove the highlights
					if cancelled {
						appOS.Window_Invalidate(ctx.window.DebugLayer, nil)
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

func (l *debugLayouter) PositionAt(ctx *Context, pt metrics.Point) (err error) {
	err = l.Layouter.PositionAt(ctx, pt)
	if err != nil {
		return
	}
	l.Pos = pt // Record position
	return
}

func (l *debugLayouter) Replayer() func(*Context) error {
	return l.Layouter.Replayer()
}

// allLayouterDebugOutlines returns an iterator of debug rectangles for the given layouter tree.
func allLayouterDebugOutlines(root Layouter) iter.Seq[native.DebugRect] {
	return func(yield func(native.DebugRect) bool) {
		// Use a stack to avoid recursive iterator calls
		type frame struct {
			layouters []Layouter
			offset    metrics.Point
		}
		stack := []frame{{layouters: []Layouter{root}, offset: metrics.Point{}}}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			for _, layouter := range current.layouters {
				debugLayouter, ok := layouter.(*debugLayouter)
				if !ok {
					continue
				}
				left := debugLayouter.Pos.X + current.offset.X
				top := debugLayouter.Pos.Y + current.offset.Y
				if !yield(native.DebugRect{
					Left:      left,
					Top:       top,
					Right:     left + debugLayouter.Size.Width,
					Bottom:    top + debugLayouter.Size.Height,
					Highlight: debugLayouter.Highlight}) {
					return
				}
				offset := current.offset
				if _, isNative := debugLayouter.Element().(ControlElement); isNative {
					offset.X += debugLayouter.Pos.X
					offset.Y += debugLayouter.Pos.Y
				}
				stack = append(stack, frame{layouters: slices.Collect(debugLayouter.Children()), offset: offset})
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
	// Container without a layouter usually indicates a programming error.
	// Provide a nop layouter if it really does not need to layout its children.
	if _, isContainer := element.Widget().(Container); isContainer {
		panic("container without a layouter")
	}
	if element.NumChildren() == 0 {
		return nil
	}
	return layouterTree(element.Child(0))
}
