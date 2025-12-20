package goui

import (
	"reflect"
	"slices"

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
	// updateChildren updates the children of the element to newChildren.
	// newChildren is the new slice of children, which may not have their parent set correctly.
	// unusedChildren contains the children that are no longer used and should be destroyed.
	updateChildren(newChildren []Element, unusedChildren []Element)
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

func (e *ElementBase) removeChildIndex(n int) {
	e.children[n].destroy()
	e.children = slices.Delete(e.children, n, n+1)
}

func (e *ElementBase) updateChildren(newChildren []Element, unusedChildren []Element) {
	for _, unused := range unusedChildren {
		unused.destroy()
	}
	for _, child := range newChildren {
		child.setParent(e)
	}
	e.children = newChildren
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
// Using this package-level function (taking the interface `Element`)
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

// buildElementTree builds the element tree for the given widget.
// The returned [Layouter] is the layouter of the returned [Element] or its nearest child.
func buildElementTree(ctx *Context, widget Widget) (elem Element, layouter Layouter, err error) {
	elem, err = performBuildElementTree(ctx, widget)
	if err != nil {
		return
	}
	layouter, err = buildLayouterTree(ctx, elem)
	return
}

// performBuildElementTree builds the element tree for the given widget.
func performBuildElementTree(ctx *Context, widget Widget) (Element, error) {
	elem, err := widget.CreateElement(ctx)
	if err != nil {
		return nil, err
	}

	elem.SetWidget(widget)

	if statefulWidget, ok := widget.(StatefulWidget); ok {
		return buildStatefulElement(ctx, elem, statefulWidget)
	}
	if statelessWidget, ok := widget.(StatelessWidget); ok {
		return buildStatelessElement(ctx, elem, statelessWidget)
	}
	if container, ok := widget.(Container); ok {
		return buildContainerElement(ctx, elem, container)
	}
	return elem, nil
}

func buildContainerElement(ctx *Context, elem Element, container Container) (Element, error) {
	numChildren := container.NumChildren()
	for i := range numChildren {
		childElem, err := performBuildElementTree(ctx, container.Child(i))
		if err != nil {
			return nil, err
		}
		element_AppendChild(elem, childElem)
	}
	return elem, nil
}

func buildStatelessElement(ctx *Context, elem Element, statelessWidget StatelessWidget) (Element, error) {
	childElem, err := performBuildElementTree(ctx, statelessWidget.Build(ctx))
	if err != nil {
		return nil, err
	}
	element_AppendChild(elem, childElem)
	return elem, nil
}

func buildStatefulElement(ctx *Context, elem Element, statefulWidget StatefulWidget) (Element, error) {
	statefulElem := elem.(*statefulElement)
	statefulElem.state = statefulWidget.CreateState(ctx)
	statefulElem.state.ctx = ctx
	statefulElem.state.element = elem
	childElem, err := performBuildElementTree(ctx, statefulElem.state.Build())
	if err != nil {
		return nil, err
	}
	element_AppendChild(elem, childElem)
	return elem, nil
}

// performUpdateElementTree is a helper of [updateElementTree] that performs the actual update.
// Parameter elem is the element to update.
// Parameter widget is the new widget to update to.
// If any error occurs during the update, the error is returned.
// The returned [Element] is the updated element(maybe the same as elem).
func performUpdateElementTree(ctx *Context, elem Element, widget Widget) (Element, error) {
	// Widgets do not match, recreate the entire element tree.
	if !widgetMatch(elem.Widget(), widget) {
		elem.destroy()
		return performBuildElementTree(ctx, widget)
	}
	// Widgets match, update the widget of the element.
	elem.SetWidget(widget)
	if container, ok := widget.(Container); ok {
		return updateContainerElement(ctx, elem, container)
	}
	if _, ok := widget.(StatefulWidget); ok {
		return updateStatefulWidget(ctx, elem)
	}
	if statelessWidget, ok := widget.(StatelessWidget); ok {
		return updateStatelessWidget(ctx, elem, statelessWidget)
	}
	return elem, nil
}

// updateStatelessWidget updates the stateless element elem to hold the new stateless widget.
func updateStatelessWidget(ctx *Context, elem Element, statelessWidget StatelessWidget) (Element, error) {
	childElem, err := performUpdateElementTree(ctx,
		elem.child(0), statelessWidget.Build(ctx))
	if err != nil {
		return nil, err
	}
	element_SetChild(elem, 0, childElem)
	return elem, nil
}

// updateStatefulWidget updates the stateful element elem to hold the new stateful widget.
func updateStatefulWidget(ctx *Context, elem Element) (Element, error) {
	statefulElement := elem.(*statefulElement)
	childElem, err := performUpdateElementTree(ctx,
		statefulElement.child(0), statefulElement.state.Build())
	if err != nil {
		return nil, err
	}
	element_SetChild(elem, 0, childElem)
	return elem, nil
}

// updateContainerElement updates the container element elem to hold the new container widget.
func updateContainerElement(ctx *Context, elem Element, container Container) (Element, error) {

	// Divide the children of elem into two parts:
	var oldElementMap map[ID]Element   // those with an ID
	var oldElementsWithoutID []Element // those without an ID
	for i := 0; i < elem.numChildren(); i++ {
		child := elem.child(i)
		id := child.Widget().WidgetID()
		if id == nil {
			oldElementsWithoutID = append(oldElementsWithoutID, child)
			continue
		}
		if oldElementMap == nil {
			oldElementMap = make(map[ID]Element)
		}
		oldElementMap[id] = child
	}

	var newChildren []Element // the updated children
	var withoutIDIndex = -1   // index of last used old element without ID

	// For each child widget in the new container, try to find a matching old element.
	// If found, update it in place; otherwise, create a new element.
	//
	// The matching is done by ID first, then by order for those without ID.
	for i := 0; i < container.NumChildren(); i++ {
		widget := container.Child(i)
		id := widget.WidgetID()
		if id != nil {
			// new widget has ID
			if oldElem, ok := oldElementMap[id]; ok {
				// found matching old element by that ID
				delete(oldElementMap, id)
				updatedElem, err := performUpdateElementTree(ctx, oldElem, widget)
				if err != nil {
					return nil, err
				}
				newChildren = append(newChildren, updatedElem)
				continue
			}
			// no matching old element by that ID, create a new one
			updatedElem, err := performBuildElementTree(ctx, widget)
			if err != nil {
				return nil, err
			}
			newChildren = append(newChildren, updatedElem)
			continue
		}
		// new widget has no ID
		// try to find the next old element without ID
		if withoutIDIndex+1 >= len(oldElementsWithoutID) {
			// no more old elements without ID, create a new one
			updatedElem, err := performBuildElementTree(ctx, widget)
			if err != nil {
				return nil, err
			}
			newChildren = append(newChildren, updatedElem)
			continue
		}
		// found next old element without ID
		withoutIDIndex++
		updatedElem, err := performUpdateElementTree(ctx, oldElementsWithoutID[withoutIDIndex], widget)
		if err != nil {
			return nil, err
		}
		newChildren = append(newChildren, updatedElem)
	}

	// Collect unused old elements
	var unusedChildren []Element
	for _, oldElem := range oldElementMap {
		unusedChildren = append(unusedChildren, oldElem)
	}
	if withoutIDIndex > 0 {
		for i := withoutIDIndex + 1; i < len(oldElementsWithoutID); i++ {
			unusedChildren = append(unusedChildren, oldElementsWithoutID[i])
		}
	}
	// Do the update
	elem.updateChildren(newChildren, unusedChildren)
	return elem, nil
}

// widgetMatch returns whether widget1 and widget2 are considered the same which
// means the element tree can be updated in place.
func widgetMatch(widget1, widget2 Widget) bool {
	return widget1.WidgetID() == widget2.WidgetID() && reflect.TypeOf(widget1) == reflect.TypeOf(widget2)
}

// updateElementTree updates the element tree to match the given widget.
// The returned updated is the updated element(maybe the same as elem).
// The returned layouter is the layouter of the updated element or its nearest child.
func updateElementTree(ctx *Context, elem Element, widget Widget) (updated Element, layouter Layouter, err error) {
	updated, err = performUpdateElementTree(ctx, elem, widget)
	if err != nil {
		return
	}
	layouter, err = buildLayouterTree(ctx, updated)
	return
}
