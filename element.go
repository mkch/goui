package goui

import (
	"reflect"
	"slices"

	"github.com/mkch/goui/native"
)

type Element interface {
	widget() Widget
	setWidget(widget Widget)
	parent() Element
	numChildren() int
	child(n int) Element
	indexChild(child Element) int
	removeChild(child Element)
	removeChildIndex(n int)
	removeChildrenRange(start, end int)
	destroy()

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

type element struct {
	theWidget Widget
	theParent Element
	children  []Element
}

func (e *element) widget() Widget {
	return e.theWidget
}

func (e *element) setWidget(widget Widget) {
	e.theWidget = widget
}

func (e *element) parent() Element {
	return e.theParent
}

func (e *element) numChildren() int {
	return len(e.children)
}

func (e *element) child(n int) Element {
	return e.children[n]
}

func (e *element) indexChild(child Element) int {
	return slices.Index(e.children, child)
}

func (e *element) removeChild(child Element) {
	i := slices.Index(e.children, child)
	if i == -1 {
		return
	}
	e.children[i].destroy()
	e.children = slices.Delete(e.children, i, i+1)
}

func (e *element) removeChildIndex(n int) {
	e.children[n].destroy()
	e.children = slices.Delete(e.children, n, n+1)
}

func (e *element) removeChildrenRange(start, end int) {
	for i := start; i < end; i++ {
		e.children[i].destroy()
	}
	e.children = slices.Delete(e.children, start, end)
}

func (e *element) destroy() {
	for _, child := range e.children {
		child.destroy()
	}
}

func (e *element) setParent(parent Element) {
	e.theParent = parent
}

func (e *element) appendChildToSlice(child Element) {
	e.children = append(e.children, child)
}

func (e *element) setChildInSlice(n int, child Element) {
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

type nativeElement struct {
	element
	layouter Layouter
	Handle   native.Handle
	// DestroyFunc is called to destroy the native handle.
	// A nil value means no special destruction is needed.
	DestroyFunc func(native.Handle) error
}

func (e *nativeElement) Layouter() Layouter {
	return e.layouter
}

func (e *nativeElement) NativeHandle(*Context) native.Handle {
	return e.Handle
}

func (e *nativeElement) destroy() {
	if e.DestroyFunc != nil {
		e.DestroyFunc(e.Handle)
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
	elem.setWidget(widget)

	var layouter Layouter
	if holder, ok := elem.(LayouterHolder); ok {
		layouter = holder.Layouter()
		layouter.setElement(elem)
	}
	if statefulWidget, ok := widget.(StatefulWidget); ok {
		return buildStatefulElement(ctx, elem, statefulWidget, parentLayouter)
	}
	if statelessWidget, ok := widget.(StatelessWidget); ok {
		return buildStatelessElement(ctx, elem, statelessWidget, parentLayouter)
	}
	if container, ok := widget.(Container); ok {
		// Container must have a Layouter
		return buildContainerElement(ctx, elem, container, layouter)
	}
	return elem, layouter, nil
}

func buildContainerElement(ctx *Context, elem Element, container Container, layouter Layouter) (Element, Layouter, error) {
	numChildren := container.NumChildren()
	for i := range numChildren {
		childElem, childLayouter, err := buildElementTree(ctx, container.Child(i), layouter)
		if err != nil {
			return nil, nil, err
		}
		element_AppendChild(elem, childElem)
		if childLayouter != nil {
			Layouter_AppendChild(layouter, childLayouter)
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
	if !widgetMatch(elem.widget(), widget) {
		elem.destroy()
		return buildElementTree(ctx, widget, parentLayouter)
	}
	// Widgets match, update the widget of the element.
	elem.setWidget(widget)
	if container, ok := widget.(Container); ok {
		return updateContainerElement(ctx, elem, container)
	}
	if _, ok := widget.(StatefulWidget); ok {
		return updateStatefulWidget(ctx, elem, parentLayouter, layouterIndex)
	}
	if statelessWidget, ok := widget.(StatelessWidget); ok {
		return updateStatelessWidget(ctx, elem, statelessWidget, parentLayouter, layouterIndex)
	}
	if holder, ok := elem.(LayouterHolder); ok {
		return elem, holder.Layouter(), nil
	}
	return elem, nil, nil
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
	layouter := elem.(LayouterHolder).Layouter()
	// Update children.
	numWidgets := container.NumChildren()
	numElements := elem.numChildren()
	var oldChildrenLayoutCount = layouter.numChildren()
	var childrenLayoutCount = 0
	var layouterIndex = -1

	var updateElement = func(i int) error {
		child := elem.child(i)
		if _, ok := child.(LayouterHolder); ok {
			layouterIndex++
		}
		childElem, childLayouter, err := performUpdateElementTree(ctx, child, container.Child(i), layouter, layouterIndex)
		if err != nil {
			return err
		}
		element_SetChild(elem, i, childElem)
		if childLayouter != nil {
			if childrenLayoutCount < oldChildrenLayoutCount {
				Layouter_SetChild(layouter, childrenLayoutCount, childLayouter)
			} else {
				Layouter_AppendChild(layouter, childLayouter)
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
					Layouter_SetChild(layouter, childrenLayoutCount, childLayouter)
				} else {
					Layouter_AppendChild(layouter, childLayouter)
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
		Layouter_AppendChild(parentLayouter, newLayouter)
	} else {
		Layouter_SetChild(parentLayouter, oldIndex, newLayouter)
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
		if holder, ok := parent.(LayouterHolder); ok {
			parentLayouter = holder.Layouter()
			break
		}
	}
	var layouterIndex = -1
	if parentLayouter != nil {
		layouterIndex = parentLayouter.indexChildFunc(func(l Layouter) bool { return l.element() == elem })
	}
	return performUpdateElementTree(ctx, elem, widget, parentLayouter, layouterIndex)
}
