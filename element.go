package goui

import (
	"reflect"
	"slices"

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/native"
)

// Element is the persistent representation of a [Widget] in the GUI tree.
type Element interface {
	Widget() Widget
	SetWidget(ctx *Context, widget Widget)
	// Layouter returns the layouter of the element. Can be nil.
	Layouter() Layouter
	// Parent returns the parent element, or nil if this is the root element of a window.
	Parent() Element

	numChildren() int
	child(n int) Element
	indexChild(child Element) int
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
	// setChildrenSlice is a helper of [element_UpdateChildren].
	// The implementation should just set the children slice to children or some equivalent.
	setChildrenSlice(children []Element)
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

func (e *ElementBase) SetWidget(ctx *Context, widget Widget) {
	e.theWidget = widget
}

func (e *ElementBase) Layouter() Layouter {
	return e.ElementLayouter
}

// NativeHandle returns the native handle associated with the given element
// or its nearest ancestor that is a native element.
// If element is nil or there is no such ancestor, ctx.NativeWindow()
// is returned.
func NativeHandle(ctx *Context, element Element) native.Handle {
	for elem := element; elem != nil; elem = elem.Parent() {
		if nativeElem, ok := elem.(nativeElement); ok {
			return nativeElem.NativeHandle()
		}
	}
	return ctx.NativeWindow()
}

func (e *ElementBase) setLayouter(layouter Layouter) {
	e.ElementLayouter = layouter
}

func (e *ElementBase) Parent() Element {
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

func (e *ElementBase) setChildrenSlice(children []Element) {
	e.children = children
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
// If the old child is different from child, the old child is destroyed.
// Then child's parent is set to parent.
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

// nativeElement is the interface all native elements implement.
// Currently all native elements embed [NativeElement].
type nativeElement interface {
	NativeHandle() native.Handle
	destroy()
}

// NativeElement is an [Element] that represents a native GUI widget.
type NativeElement struct {
	ElementBase
	Handle native.Handle
	// DestroyFunc is called to destroy the native handle.
	// A nil value means no special destruction is needed.
	DestroyFunc func(native.Handle) error
}

// NativeElement returns e.Handle.
func (e *NativeElement) NativeHandle() native.Handle {
	// This method implements the nativeElement interface.
	return e.Handle
}

func (e *NativeElement) destroy() {
	if e.DestroyFunc != nil {
		err := e.DestroyFunc(e.Handle)
		if err != nil {
			errortrace.Panic(err)
		}
	}
}

// buildElementTree builds the element tree for the given widget.
// The returned layouter is the layouter of the returned element or its nearest child.
func buildElementTree(ctx *Context, widget Widget) (element Element, layouter Layouter, err error) {
	element, err = buildElementTreeImpl(ctx, nil, widget)
	if err != nil {
		return
	}
	layouter = layouterTree(element)
	return
}

// buildElementTreeImpl builds the element tree for the given widget.
// The parent element can be nil. If parent is not nil, the ancestor
// tree of parent must already be established.
// The returned element is the root element of the built tree.
func buildElementTreeImpl(ctx *Context, parent Element, widget Widget) (Element, error) {
	elem, err := widget.CreateElement(ctx, parent)
	if err != nil {
		return nil, err
	}

	if parent != nil {
		// Appending elem to parent before elem is fully constructed because
		// CreateElement method of native widget looks for its native parent
		// handle up the element tree. See [NativeHandle] function.
		element_AppendChild(parent, elem)
	}

	if layouter := elem.Layouter(); layouter != nil {
		layouter.setElement(elem)
		if ctx.app.debug.LayoutOutlineEnabled() {
			layouter = &debugLayouter{
				Layouter: layouter,
			}
			elem.setLayouter(layouter)
		}
	}

	elem.SetWidget(ctx, widget)

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
		// The child element is already appended to elem in buildElementTreeImpl.
		_, err := buildElementTreeImpl(ctx, elem, container.Child(i))
		if err != nil {
			return nil, err
		}
	}
	return elem, nil
}

func buildStatelessElement(ctx *Context, elem Element, statelessWidget StatelessWidget) (Element, error) {
	// The child element is already appended to elem in buildElementTreeImpl.
	_, err := buildElementTreeImpl(ctx, elem, statelessWidget.Build(ctx))
	if err != nil {
		return nil, err
	}
	return elem, nil
}

func buildStatefulElement(ctx *Context, elem Element, statefulWidget StatefulWidget) (Element, error) {
	statefulElement := elem.(*statefulElement)
	statefulElement.state = statefulWidget.CreateState(&StateContext{ctx, statefulElement})
	// The child element is already appended to elem in buildElementTreeImpl.
	_, err := buildElementTreeImpl(ctx, elem, statefulElement.state.Build())
	if err != nil {
		return nil, err
	}
	return elem, nil
}

// updateElementTree is a helper of [reconcileElementTree] that performs the in-place update.
// This function must be called when [widgetMatch] returns true for elem.Widget() and widget.
// The elem will be updated to hold widget.
// If any error occurs during the update, the error is returned.
func updateElementTree(ctx *Context, elem Element, widget Widget) (err error) {
	elem.SetWidget(ctx, widget)
	if container, ok := widget.(Container); ok {
		return updateContainerElement(ctx, elem, container)
	}
	if _, ok := widget.(StatefulWidget); ok {
		return updateStatefulWidget(ctx, elem)
	}
	if statelessWidget, ok := widget.(StatelessWidget); ok {
		return updateStatelessWidget(ctx, elem, statelessWidget)
	}
	return nil
}

// updateStatelessWidget updates the stateless element elem to hold the new stateless widget.
func updateStatelessWidget(ctx *Context, elem Element, statelessWidget StatelessWidget) error {
	err := reconciledChildElement(ctx, elem, 0, statelessWidget.Build(ctx))
	if err != nil {
		return err
	}
	return nil
}

// updateStatefulWidget updates the stateful element elem to hold the new stateful widget.
func updateStatefulWidget(ctx *Context, elem Element) error {
	statefulElement := elem.(*statefulElement)
	// rebuild the child widget and reconcile.
	err := reconciledChildElement(
		ctx,
		statefulElement, 0,
		statefulElement.state.Build(),
	)
	if err != nil {
		return err
	}
	return nil
}

// updateContainerElement updates the container element to hold the new container widget.
func updateContainerElement(ctx *Context, element Element, container Container) error {
	var newChildren = make([]Element, container.NumChildren()) // the updated children

	numElem := element.numChildren()
	numWidget := container.NumChildren()

	// Phase 1: Top-down match

	var topDownCount = 0 // number of matched elements(widgets) from the top
	for i := 0; i < min(numElem, numWidget); i++ {
		widget := container.Child(i)
		elem := element.child(i)
		if !widgetMatch(widget, elem.Widget()) {
			break
		}
		err := updateElementTree(ctx, elem, widget)
		if err != nil {
			return err
		}
		newChildren[i] = elem
		topDownCount++
	}

	// Phase 2: Bottom-up match

	var bottomUpCount = 0 // number of matched elements(widgets) from the bottom
	for i := 0; numElem-i > topDownCount && numWidget-i > topDownCount; i++ {
		widgetIndex := numWidget - 1 - i
		elemIndex := numElem - 1 - i
		widget := container.Child(widgetIndex)
		elem := element.child(elemIndex)
		if !widgetMatch(widget, elem.Widget()) {
			break
		}
		err := updateElementTree(ctx, elem, widget)
		if err != nil {
			return err
		}
		newChildren[widgetIndex] = elem
		bottomUpCount++
	}

	// Phase 3: Handle the middle part:
	//   Widgets and elements with ID are matched by ID.
	//   Widgets without ID and unmatched widgets are treated as new, and new elements are created for them.
	//   Elements without ID and unmatched elements are destroyed.

	var unmatchedKeyedElements map[ID]Element // old elements with ID in the middle
	var unusedElements []Element              // old elements without ID in the middle
	if topDownCount+bottomUpCount < numElem { // if there are old elements left
		unmatchedKeyedElements = make(map[ID]Element, numElem-topDownCount-bottomUpCount)
		// collect old elements with ID
		for i := topDownCount; i <= numElem-1-bottomUpCount; i++ {
			elem := element.child(i)
			id := elem.Widget().WidgetID()
			if id != nil {
				unmatchedKeyedElements[id] = elem
			} else {
				unusedElements = append(unusedElements, elem)
			}
		}
	}
	// Process widgets in the middle part
	for i := topDownCount; i <= numWidget-1-bottomUpCount; i++ {
		widget := container.Child(i)
		widgetID := widget.WidgetID()
		matchedElem := unmatchedKeyedElements[widgetID] // no need to handle nil ID here
		var updatedElem Element
		var err error
		if matchedElem == nil {
			updatedElem, err = buildElementTreeImpl(ctx, element, widget)
		} else {
			updatedElem, err = reconcileElementTreeImpl(ctx, matchedElem, widget)
			if updatedElem != matchedElem {
				matchedElem.destroy()
			}
			delete(unmatchedKeyedElements, widgetID)
		}
		if err != nil {
			return err
		}
		newChildren[i] = updatedElem
	}
	// Destroy unmatched old elements
	for _, unmatched := range unmatchedKeyedElements {
		unmatched.destroy()
	}
	// Destroy unused old elements
	for _, unusedElem := range unusedElements {
		unusedElem.destroy()
	}
	// Update the children
	for _, child := range newChildren {
		child.setParent(element)
	}
	element.setChildrenSlice(newChildren)
	return nil
}

// widgetMatch returns whether widget1 and widget2 are considered the same which
// means the element tree can be updated in place.
func widgetMatch(widget1, widget2 Widget) bool {
	return widget1.WidgetID() == widget2.WidgetID() && reflect.TypeOf(widget1) == reflect.TypeOf(widget2)
}

// reconcileElementTreeImpl performs the actual reconciliation.
// It recreates the element tree if the widgets do not match, or updates it in place if they match.
// The reconciled element and any error occurred during the process are returned.
// If a new element is created, it is returned as the reconciled and the old element is not destroyed.
func reconcileElementTreeImpl(ctx *Context, element Element, widget Widget) (reconciled Element, err error) {
	// Widgets do not match, recreate the entire element tree.
	if !widgetMatch(element.Widget(), widget) {
		return buildElementTreeImpl(ctx, element.Parent(), widget)
	}
	// Widgets match, update the widget of the element.
	err = updateElementTree(ctx, element, widget)
	if err != nil {
		return
	}
	return element, nil
}

// reconciledChildElement reconciles the child element at childIndex of parent with widget.
func reconciledChildElement(ctx *Context, parent Element, childIndex int, widget Widget) (err error) {
	oldChild := parent.child(childIndex)
	newChild, err := reconcileElementTreeImpl(ctx, parent.child(childIndex), widget)
	if err != nil {
		return err
	}
	if newChild != oldChild {
		element_SetChild(parent, childIndex, newChild)
	}
	return
}
