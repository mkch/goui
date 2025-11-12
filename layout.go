package goui

import (
	"slices"
	"unsafe"
)

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

type Constraints struct {
	MinWidth  int
	MinHeight int
	MaxWidth  int
	MaxHeight int
}

type Size struct {
	Width  int
	Height int
}

type Point struct {
	X, Y int
}

// Layouter is the interface for laying out elements.
type Layouter interface {
	// Layout computes the size of the element given the constraints.
	Layout(ctx *Context, constraints Constraints) Size
	// Apply puts the element at the given position.
	Apply(x, y int) error

	numChildren() int
	child(n int) Layouter
	removeChild(child Layouter)
	removeChildIndex(n int)
	parent() Layouter
	setParent(parent Layouter)
	element() Element
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
	theElement Element
	theParent  Layouter
	children   []Layouter
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

func (l *LayouterBase) removeChildIndex(n int) {
	l.children[n].setParent(nil)
	l.children = slices.Delete(l.children, n, n+1)
}

func (l *LayouterBase) removeChild(child Layouter) {
	l.children = slices.DeleteFunc(l.children, func(l Layouter) bool { return l == child })
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

func buildLayouterTree(ctx *Context, elem Element) (layouter Layouter, err error) {
	if layouterHolder, ok := elem.(LayouterHolder); ok {
		layouter = layouterHolder.Layouter()
		layouter.setElement(elem)
	}
	for i := 0; i < elem.numChildren(); i++ {
		childElem := elem.child(i)
		childLayouter, err := buildLayouterTree(ctx, childElem)
		if err != nil {
			return nil, err
		}
		if layouter == nil {
			// buildElementTree ensures that Container widget must have a Layouter,
			// so when this happens, the childLayouter must be the only child of StatefulWidget or StatelessWidget.
			layouter = childLayouter
		}
		Layouter_AppendChild(layouter, childLayouter)
	}
	return layouter, nil
}
