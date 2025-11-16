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

const Infinite = 1<<(unsafe.Sizeof(int(0))*8-1) - 1

// clampInt clamps value between min and max.
func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

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

// Tight returns true if the constraints are tight (min and max are equal).
func (c *Constraints) Tight() bool {
	return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
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

type layouterReplayData struct {
	constraints Constraints
	position    Point
}

// Layouter is the interface for laying out elements.
type Layouter interface {
	// Layout computes the size of the element given the constraints.
	Layout(ctx *Context, constraints Constraints) (Size, error)
	// PositionAt puts the element at the given position.
	PositionAt(x, y int) error
	// ChildrenIndependent returns true if the size of layout does not depend on its children.
	ChildrenIndependent() bool

	numChildren() int
	child(n int) Layouter
	indexChildFunc(f func(Layouter) bool) int
	removeChild(child Layouter)
	removeChildIndex(n int)
	removeChildrenRange(start, end int)
	parent() Layouter
	setParent(parent Layouter)
	element() Element
	setElement(e Element)

	// replayData is a helper of [Layouter_Replayer].
	// If the current state of the Layouter does not support replaying,
	// it should return nil.
	replayData() *layouterReplayData
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
	theElement      Element
	theParent       Layouter
	lastConstraints *Constraints
	position        Point
	size            Size
	children        []Layouter
}

func (l *LayouterBase) element() Element {
	return l.theElement
}

func (l *LayouterBase) setElement(e Element) {
	l.theElement = e
}

func (l *LayouterBase) parent() Layouter {
	return l.theParent
}

func (l *LayouterBase) numChildren() int {
	return len(l.children)
}

func (l *LayouterBase) child(n int) Layouter {
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

// Layout should be called by the embedding struct's Layout method.
// Note that this Layout method does not satisfy the Layouter interface
// due to the different return type.
func (l *LayouterBase) Layout(ctx *Context, constraints Constraints) {
	l.lastConstraints = &constraints
}

// PositionAt should be called by the embedding struct's PositionAt method.
// Note that this PositionAt method does not satisfy the Layouter interface
// due to the different return type.
func (l *LayouterBase) PositionAt(x, y int) {
	l.position = Point{X: x, Y: y}
}

func (l *LayouterBase) replayData() *layouterReplayData {
	if l.lastConstraints == nil {
		return nil
	}
	return &layouterReplayData{
		constraints: *l.lastConstraints,
		position:    l.position,
	}
}

func (l *LayouterBase) ChildrenIndependent() bool {
	return false
}

// Layouter_Replayer returns a function that replays the lat layout operations of the given Layouter.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func Layouter_Replayer(l Layouter) func(*Context) error {
	if !l.ChildrenIndependent() {
		return nil
	}
	data := l.replayData()
	if data == nil {
		return nil
	}
	return func(ctx *Context) error {
		if _, err := l.Layout(ctx, data.constraints); err != nil {
			return err
		}
		return l.PositionAt(data.position.X, data.position.Y)
	}
}

// Layouter_AppendChild appends child to parent Layouter.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func Layouter_AppendChild(parent, child Layouter) {
	child.setParent(parent)
	parent.appendChildToSlice(child)
}

// Layouter_SetChild sets child at index n of parent Layouter.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func Layouter_SetChild(parent Layouter, n int, child Layouter) {
	if parent.child(n) == child {
		return
	}
	parent.setChildInSlice(n, child)
	child.setParent(parent)
}

// Layout_InsertChild inserts child at index n of parent Layouter.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func Layouter_InsertChild(parent Layouter, i int, child Layouter) {
	parent.insertChildInSlice(i, child)
	child.setParent(parent)
}

// LayouterHolder is an interface that [Element] can implement to provide a Layouter.
type LayouterHolder interface {
	Layouter() Layouter
}
