package goui

import (
	"fmt"
	"slices"
	"unsafe"
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
	// insertChildInSlice is a helper of [Layouter_InsertChild].
	// The implementation should just insert child at index n in the children slice or some equivalent.
	insertChildInSlice(i int, child Layouter)
}

// LayouterBase is a helper struct for implementing Layouter.
// Embedding LayouterBase in a struct and implementing
// Layout and Apply methods implements the Layouter interface.
type LayouterBase struct {
	Position   Point
	Size       Size
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

func (l *LayouterBase) insertChildInSlice(i int, child Layouter) {
	l.children = slices.Insert(l.children, i, child)
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

// Layout_InsertChild inserts child at index n of parent Layouter.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func layouter_InsertChild(parent Layouter, i int, child Layouter) {
	parent.insertChildInSlice(i, child)
	child.setParent(parent)
}

// LayouterHolder is an interface that [Element] can implement to provide a Layouter.
type LayouterHolder interface {
	ElementLayouter() Layouter
}
