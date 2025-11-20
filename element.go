package goui

import (
	"iter"
	"reflect"
	"slices"
	"time"

	"github.com/mkch/goui/native"
)

// Element is the persistent representation of a [Widget] in the GUI tree.
type Element interface {
	Widget() Widget
	SetWidget(widget Widget)
	// Layouter returns the layouter of the element. Can be nil.
	Layouter() Layouter
	parent() Element
	numChildren() int
	child(n int) Element
	indexChild(child Element) int
	removeChild(child Element)
	removeChildIndex(n int)
	removeChildrenRange(start, end int)
	destroy()

	// setLayouter sets the layouter of the element. For debug purposes only.
	setLayouter(layouter Layouter)
	// setParent is a helper of [element_AppendChild].
	// The implementation should just set the parent field or some equivalent.
	setParent(parent Element)
	// appendChildToSlice is a helper of [element_AppendChild].
	// The implementation should just append child to the children slice or some equivalent.
	appendChildToSlice(child Element)
	// setChildInSlice is a helper of [element_SetChild].
	// The implementation should just set child at index n in the children slice or some equivalent.
	setChildInSlice(n int, child Element)
}

// ElementBase implements [Element], and is the building block for other Element types.
type ElementBase struct {
	// ElementLayouter is the layouter of the element. Can be nil.
	// This field is returned by Layouter() method.
	ElementLayouter Layouter
	theWidget       Widget
	theParent       Element
	children        []Element
}

func (e *ElementBase) Widget() Widget {
	return e.theWidget
}

func (e *ElementBase) SetWidget(widget Widget) {
	e.theWidget = widget
}

func (e *ElementBase) Layouter() Layouter {
	return e.ElementLayouter
}

func (e *ElementBase) setLayouter(layouter Layouter) {
	e.ElementLayouter = layouter
}

func (e *ElementBase) parent() Element {
	return e.theParent
}

func (e *ElementBase) numChildren() int {
	return len(e.children)
}

func (e *ElementBase) child(n int) Element {
	return e.children[n]
}

func (e *ElementBase) indexChild(child Element) int {
	return slices.Index(e.children, child)
}

func (e *ElementBase) removeChild(child Element) {
	i := slices.Index(e.children, child)
	if i == -1 {
		return
	}
	e.children[i].destroy()
	e.children = slices.Delete(e.children, i, i+1)
}

func (e *ElementBase) removeChildIndex(n int) {
	e.children[n].destroy()
	e.children = slices.Delete(e.children, n, n+1)
}

func (e *ElementBase) removeChildrenRange(start, end int) {
	for i := start; i < end; i++ {
		e.children[i].destroy()
	}
	e.children = slices.Delete(e.children, start, end)
}

func (e *ElementBase) destroy() {
	for _, child := range e.children {
		child.destroy()
	}
}

func (e *ElementBase) setParent(parent Element) {
	e.theParent = parent
}

func (e *ElementBase) appendChildToSlice(child Element) {
	e.children = append(e.children, child)
}

func (e *ElementBase) setChildInSlice(n int, child Element) {
	e.children[n] = child
}

// element_AppendChild appends child to parent and sets child's parent to parent.
//
// We keep this as a package-level function (instead of a method like
// parent.AppendChild(child)) because Go does not have polymorphic
// receiver. If AppendChild were implemented as a method on *elementBase
// (the embedded base type), calling it would use the *elementBase
// receiver value — and child.setParent(e) would set the child's parent
// dynamic type to *elementBase, not the outer concrete type that embeds
// elementBase (e.g. *nativeElement).
//
// Using this package-level function (taking the interface `element`)
// preserves the original parent's dynamic type when calling
// child.setParent(parent).
func element_AppendChild(parent, child Element) {
	parent.appendChildToSlice(child)
	child.setParent(parent)
}

// element_SetChild sets the nth child of parent to child.
//
// See [element_AppendChild] for explanation why this is a package-level function.
func element_SetChild(parent Element, n int, child Element) {
	if parent.child(n) == child {
		return
	}
	parent.child(n).destroy()
	parent.setChildInSlice(n, child)
	child.setParent(parent)
}

// NativeElement is an [Element] that represents a native GUI widget.
type NativeElement struct {
	ElementBase
	Handle native.Handle
	// DestroyFunc is called to destroy the native handle.
	// A nil value means no special destruction is needed.
	DestroyFunc func(native.Handle) error
}

func (e *NativeElement) NativeHandle(*Context) native.Handle {
	return e.Handle
}

func (e *NativeElement) destroy() {
	if e.DestroyFunc != nil {
		e.DestroyFunc(e.Handle)
	}
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

// buildElementTree builds the element tree for the given widget.
// Parameter parentLayouter is thee nearest recursive parent layouter,
// or nil if there is no recursive parent layouter.
// If any error occurs during the build, the error is returned.
// The returned [Element] is the root element of the built tree.
// The returned [Layouter] is the layouter of the returned [Element] or its nearest child.
func buildElementTree(ctx *Context, widget Widget, parentLayouter Layouter) (Element, Layouter, error) {
	elem, err := widget.CreateElement(ctx)
	if err != nil {
		return nil, nil, err
	}

	layouter := elem.Layouter()
	if layouter != nil {
		if ctx.window.DebugLayout {
			layouter = &debugLayouter{
				Layouter: layouter,
				Ctx:      ctx,
			}
			elem.setLayouter(layouter)
		}
		layouter.setElement(elem)
	}

	elem.SetWidget(widget)

	if statefulWidget, ok := widget.(StatefulWidget); ok {
		return buildStatefulElement(ctx, elem, statefulWidget, parentLayouter)
	}
	if statelessWidget, ok := widget.(StatelessWidget); ok {
		return buildStatelessElement(ctx, elem, statelessWidget, parentLayouter)
	}
	if container, ok := widget.(Container); ok {
		return buildContainerElement(ctx, elem, container)
	}
	return elem, layouter, nil
}

func buildContainerElement(ctx *Context, elem Element, container Container) (Element, Layouter, error) {
	// Container must have a Layouter
	layouter := elem.Layouter()
	numChildren := container.NumChildren()
	for i := range numChildren {
		childElem, childLayouter, err := buildElementTree(ctx, container.Child(i), layouter)
		if err != nil {
			return nil, nil, err
		}
		element_AppendChild(elem, childElem)
		if childLayouter != nil {
			layouter_AppendChild(layouter, childLayouter)
		}
	}
	return elem, layouter, nil
}

func buildStatelessElement(ctx *Context, elem Element, statelessWidget StatelessWidget, parentLayouter Layouter) (Element, Layouter, error) {
	childElem, childLayouter, err := buildElementTree(ctx, statelessWidget.Build(ctx), parentLayouter)
	if err != nil {
		return nil, nil, err
	}
	element_AppendChild(elem, childElem)
	return elem, childLayouter, nil
}

func buildStatefulElement(ctx *Context, elem Element, statefulWidget StatefulWidget, parentLayouter Layouter) (Element, Layouter, error) {
	statefulElem := elem.(*statefulElement)
	statefulElem.state = statefulWidget.CreateState(ctx)
	statefulElem.state.ctx = ctx
	statefulElem.state.element = elem
	childElem, childLayouter, err := buildElementTree(ctx, statefulElem.state.Build(), parentLayouter)
	if err != nil {
		return nil, nil, err
	}
	element_AppendChild(elem, childElem)
	return elem, childLayouter, nil
}

// performUpdateElementTree is a helper of [updateElementTree] that performs the actual update.
// Parameter elem is the element to update.
// Parameter widget is the new widget to update to.
// Parameter parentLayouter is thee nearest recursive parent layouter, or nil if there is no recursive parent layouter.
// Parameter layouterIndex is the index of the old layouter in the parentLayouter,
// or -1 if parentLayouter is nil.
// If any error occurs during the update, the error is returned.
// The returned [Element] is the updated element(maybe the same as elem).
// The returned [Layouter] is the layouter of the returned [Element] or its nearest child.
func performUpdateElementTree(ctx *Context, elem Element, widget Widget, parentLayouter Layouter, layouterIndex int) (Element, Layouter, error) {
	// Widgets do not match, recreate the entire element tree.
	if !widgetMatch(elem.Widget(), widget) {
		elem.destroy()
		return buildElementTree(ctx, widget, parentLayouter)
	}
	// Widgets match, update the widget of the element.
	elem.SetWidget(widget)
	if container, ok := widget.(Container); ok {
		return updateContainerElement(ctx, elem, container)
	}
	if _, ok := widget.(StatefulWidget); ok {
		return updateStatefulWidget(ctx, elem, parentLayouter, layouterIndex)
	}
	if statelessWidget, ok := widget.(StatelessWidget); ok {
		return updateStatelessWidget(ctx, elem, statelessWidget, parentLayouter, layouterIndex)
	}
	return elem, elem.Layouter(), nil
}

// updateStatelessWidget updates the stateless element elem to hold the new stateless widget.
func updateStatelessWidget(ctx *Context, elem Element, statelessWidget StatelessWidget, parentLayouter Layouter, layouterIndex int) (Element, Layouter, error) {
	childElem, childLayouter, err := performUpdateElementTree(ctx,
		elem.child(0), statelessWidget.Build(ctx),
		parentLayouter, layouterIndex)
	if err != nil {
		return nil, nil, err
	}
	element_SetChild(elem, 0, childElem)
	updateLayouter(childLayouter, parentLayouter, layouterIndex)
	return elem, childLayouter, nil
}

// updateStatefulWidget updates the stateful element elem to hold the new stateful widget.
func updateStatefulWidget(ctx *Context, elem Element, parentLayouter Layouter, layouterIndex int) (Element, Layouter, error) {
	statefulElement := elem.(*statefulElement)
	childElem, childLayouter, err := performUpdateElementTree(ctx,
		statefulElement.child(0), statefulElement.state.Build(),
		parentLayouter, layouterIndex)
	if err != nil {
		return nil, nil, err
	}
	element_SetChild(elem, 0, childElem)
	updateLayouter(childLayouter, parentLayouter, layouterIndex)
	return elem, childLayouter, nil
}

// updateContainerElement updates the container element elem to hold the new container widget.
func updateContainerElement(ctx *Context, elem Element, container Container) (Element, Layouter, error) {
	// Container must have a Layouter
	layouter := elem.Layouter()
	// Update children.
	numWidgets := container.NumChildren()
	numElements := elem.numChildren()
	var oldChildrenLayoutCount = layouter.NumChildren()
	var childrenLayoutCount = 0
	var layouterIndex = -1

	var updateElement = func(i int) error {
		child := elem.child(i)
		if child.Layouter() != nil {
			layouterIndex++
		}
		childElem, childLayouter, err := performUpdateElementTree(ctx, child, container.Child(i), layouter, layouterIndex)
		if err != nil {
			return err
		}
		element_SetChild(elem, i, childElem)
		if childLayouter != nil {
			if childrenLayoutCount < oldChildrenLayoutCount {
				layouter_SetChild(layouter, childrenLayoutCount, childLayouter)
			} else {
				layouter_AppendChild(layouter, childLayouter)
			}
			childrenLayoutCount++
		}
		return nil
	}
	if numElements <= numWidgets {
		// Update existing elements.
		for i := range numElements {
			if err := updateElement(i); err != nil {
				return nil, nil, err
			}
		}
		// Add new elements.
		for i := numElements; i < numWidgets; i++ {
			childElement, childLayouter, err := buildElementTree(ctx, container.Child(i), layouter)
			if err != nil {
				return nil, nil, err
			}
			element_AppendChild(elem, childElement)
			if childLayouter != nil {
				if childrenLayoutCount < oldChildrenLayoutCount {
					layouter_SetChild(layouter, childrenLayoutCount, childLayouter)
				} else {
					layouter_AppendChild(layouter, childLayouter)
				}
				childrenLayoutCount++
			}
		}
	} else {
		// Update existing elements.
		for i := range numWidgets {
			if err := updateElement(i); err != nil {
				return nil, nil, err
			}
		}
		// Remove extra elements.
		elem.removeChildrenRange(numWidgets, numElements)
	}
	// Remove extra layouts.
	if childrenLayoutCount < oldChildrenLayoutCount {
		layouter.removeChildrenRange(childrenLayoutCount, oldChildrenLayoutCount)
	}
	return elem, layouter, nil
}

// updateLayouter updates the layouter in the parentLayouter.
// Parameter oldIndex is the index of the old layouter in the parentLayouter,
// or -1 if there was no corresponding old layouter.
func updateLayouter(newLayouter Layouter, parentLayouter Layouter, oldIndex int) {
	if newLayouter == nil {
		if oldIndex >= 0 {
			parentLayouter.removeChildIndex(oldIndex)
		}
		return
	}
	if oldIndex == -1 {
		layouter_AppendChild(parentLayouter, newLayouter)
	} else {
		layouter_SetChild(parentLayouter, oldIndex, newLayouter)
	}
}

// widgetMatch returns whether widget1 and widget2 are considered the same which
// means the element tree can be updated in place.
func widgetMatch(widget1, widget2 Widget) bool {
	return widget1.WidgetID() == widget2.WidgetID() && reflect.TypeOf(widget1) == reflect.TypeOf(widget2)
}

// updateElementTree updates the element tree to match the given widget.
// The returned [Element] is the updated element(maybe the same as elem).
// The returned [Layouter] is the layouter of the returned [Element] or its nearest child.
func updateElementTree(ctx *Context, elem Element, widget Widget) (Element, Layouter, error) {
	var parentLayouter Layouter
	for parent := elem.parent(); parent != nil; parent = parent.parent() {
		if parentLayouter = parent.Layouter(); parentLayouter != nil {
			break
		}
	}
	var layouterIndex = -1
	if parentLayouter != nil {
		layouterIndex = parentLayouter.indexChildFunc(func(l Layouter) bool { return l.Element() == elem })
	}
	return performUpdateElementTree(ctx, elem, widget, parentLayouter, layouterIndex)
}
