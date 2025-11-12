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
	e.children[i].setParent(nil)
	e.children = slices.Delete(e.children, i, i+1)
}

func (e *element) removeChildIndex(n int) {
	e.children[n].setParent(nil)
	e.children = slices.Delete(e.children, n, n+1)
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

// buildElementTree builds the element tree for the given widget recursively.
// The returned element is the root of the built element tree, which has a nil parent.
func buildElementTree(ctx *Context, widget Widget) (Element, error) {
	elem, err := widget.CreateElement(ctx)
	if err != nil {
		return nil, err
	}
	elem.setWidget(widget)
	if statefulWidget, ok := widget.(StatefulWidget); ok {
		statefulElem := elem.(*statefulElement)
		statefulElem.state = statefulWidget.CreateState(ctx)
		statefulElem.state.ctx = ctx
		statefulElem.state.element = elem
		childWidget := statefulElem.state.Build()
		childElem, err := buildElementTree(ctx, childWidget)
		if err != nil {
			return nil, err
		}
		element_AppendChild(elem, childElem)
	} else if statelessWidget, ok := widget.(StatelessWidget); ok {
		childWidget := statelessWidget.Build(ctx)
		childElem, err := buildElementTree(ctx, childWidget)
		if err != nil {
			return nil, err
		}
		element_AppendChild(elem, childElem)
	} else if container, ok := widget.(Container); ok {
		numChildren := container.NumChildren()
		for i := range numChildren {
			childElem, err := buildElementTree(ctx, container.Child(i))
			if err != nil {
				return nil, err
			}
			element_AppendChild(elem, childElem)
		}
	}
	return elem, nil
}

// performUpdateElementTree updates the element tree rooted at elem to match the given widget tree.
// Param elemIndex is the index of elem in its parent, if elem has no parent, elemIndex must be -1.
func performUpdateElementTree(ctx *Context, elem Element, elemIndex int, widget Widget) error {
	if widgetMatch(elem.widget(), widget) {
		// Widgets match, update the widget of the element.
		elem.setWidget(widget)
		if container, ok := widget.(Container); ok {
			// Update children.
			numWidgets := container.NumChildren()
			numElements := elem.numChildren()
			if numWidgets >= numElements {
				// Update existing elements.
				for i := range numElements {
					err := performUpdateElementTree(ctx, elem.child(i), i, container.Child(i))
					if err != nil {
						return err
					}
				}
				// Add new elements.
				for i := numElements; i < numWidgets; i++ {
					childElement, err := buildElementTree(ctx, container.Child(i))
					if err != nil {
						return err
					}
					element_AppendChild(elem, childElement)
				}
			} else if numElements > numWidgets {
				// Remove extra elements.
				for i := numElements - 1; i >= numWidgets; i-- {
					child := elem.child(i)
					child.destroy()
					elem.removeChildIndex(i)
				}
			}
			return nil
		} else if _, ok := widget.(StatefulWidget); ok {
			statefulElement := elem.(*statefulElement)
			err := performUpdateElementTree(ctx, statefulElement.child(0), 0, statefulElement.state.Build())
			if err != nil {
				return err
			}
		}
		return nil
	}
	// Widgets do not match, recreate the entire element tree.
	elem.destroy()
	elem, err := buildElementTree(ctx, widget)
	if err != nil {
		return err
	}
	if parent := elem.parent(); parent == nil {
		if elemIndex != -1 {
			panic("index out of range")
		}
		ctx.window.Root = elem
	} else {
		element_SetChild(parent, elemIndex, elem)
	}
	return nil
}

func widgetMatch(widget1, widget2 Widget) bool {
	return widget1.WidgetID() == widget2.WidgetID() && reflect.TypeOf(widget1) == reflect.TypeOf(widget2)
}

// updateElementTree updates the element tree rooted at elem to match the given widget tree.
// The parent of elem must not be nil.
func updateElementTree(ctx *Context, elem Element, widget Widget) error {
	if parent := elem.parent(); parent == nil {
		err := performUpdateElementTree(ctx, elem, -1, widget)
		if err != nil {
			return err
		}
	} else {
		err := performUpdateElementTree(ctx, elem, parent.indexChild(elem), widget)
		if err != nil {
			return err
		}
	}
	return nil
}
