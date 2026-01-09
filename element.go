package goui

import (
	"errors"
	"reflect"
	"slices"

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/native"
)

// Element is the persistent representation of a [Widget] in the GUI tree.
type Element interface {
	Widget() AbstractWidget
	SetWidget(ctx *Context, widget AbstractWidget) error
	// Layouter returns the layouter of the element. Can be nil.
	Layouter() Layouter
	// Parent returns the parent element, or nil if this is the root element of a window.
	Parent() Element
	// Destroy destroys the element and releases any associated resources.
	Destroy() error
	// NumChildren returns the number of child elements.
	NumChildren() int
	// Child returns the nth child element.
	Child(n int) Element

	indexChild(child Element) int

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

// ElementHelper implements [Element], and is the building block for other Element types.
type ElementHelper struct {
	// ElementLayouter is the layouter of the element. Can be nil.
	// This field is returned by Layouter() method.
	ElementLayouter Layouter
	theWidget       AbstractWidget
	theParent       Element
	children        []Element
}

func (e *ElementHelper) Widget() AbstractWidget {
	return e.theWidget
}

func (e *ElementHelper) SetWidget(ctx *Context, widget AbstractWidget) error {
	e.theWidget = widget
	return nil
}

func (e *ElementHelper) Layouter() Layouter {
	return e.ElementLayouter
}

// ErrWrongParent is returned when build element tree fails due to invalid parent element.
// For example, a native control widget is created under a native menu element or vice versa.
var ErrWrongParent = errors.New("wrong parent")

// WidgetNativeParent returns the native handle of the given widget element
// or its nearest ancestor that is a [ControlElement].
// If element is nil or there is no such ancestor, ctx.NativeWindow()
// is returned.
// If element does not belong to a widget tree, [ErrWrongParent] is returned.
func WidgetNativeParent(ctx *Context, element Element) (native.Handle, error) {
	if element == nil {
		return ctx.NativeWindow(), nil
	}
	if _, ok := element.Widget().(interface{ ExclusiveType(marker.TypeWidget) }); !ok {
		// Not a widget tree
		return nil, errortrace.WithStack(ErrWrongParent)
	}
	type R struct {
		h   native.Handle
		err error
	}
	r, found := LookupParent(element, func(e Element) (R, bool) {
		if elem, ok := e.(ControlElement); ok {
			return R{elem.NativeControl(), nil}, true
		}
		return R{}, false // Continue searching
	})
	if found {
		return r.h, r.err
	}
	// If no parent found, use the window handle
	return ctx.NativeWindow(), nil
}

func (e *ElementHelper) setLayouter(layouter Layouter) {
	e.ElementLayouter = layouter
}

func (e *ElementHelper) Parent() Element {
	return e.theParent
}

func (e *ElementHelper) NumChildren() int {
	return len(e.children)
}

func (e *ElementHelper) Child(n int) Element {
	return e.children[n]
}

func (e *ElementHelper) indexChild(child Element) int {
	return slices.Index(e.children, child)
}

func (e *ElementHelper) setChildrenSlice(children []Element) {
	e.children = children
}

// Destroy implements [Element.Destroy].
func (e *ElementHelper) Destroy() (err error) {
	for _, child := range e.children {
		if err = child.Destroy(); err != nil {
			return
		}
	}
	return
}

func (e *ElementHelper) setParent(parent Element) {
	e.theParent = parent
}

func (e *ElementHelper) appendChildToSlice(child Element) {
	e.children = append(e.children, child)
}

func (e *ElementHelper) setChildInSlice(n int, child Element) {
	e.children[n] = child
}

// element_AppendChild appends child to parent and sets child's parent to parent.
//
// We keep this as a package-level function (instead of a method like
// parent.AppendChild(child)) because Go does not have polymorphic
// receiver. If AppendChild were implemented as a method on *[ElementHelper]
// (the embedded helper type), calling it would use the *[ElementHelper]
// receiver value — and child.setParent(e) would set the child's parent
// dynamic type to *[ElementHelper], not the outer concrete type that embeds
// [ElementHelper] (e.g. *nativeElement).
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
func element_SetChild(parent Element, n int, child Element) (err error) {
	if parent.Child(n) == child {
		return
	}
	if err = parent.Child(n).Destroy(); err != nil {
		return
	}
	parent.setChildInSlice(n, child)
	child.setParent(parent)
	return
}

// ControlElement is the element that represent a native GUI control.
type ControlElement interface {
	Element
	// Returns the native handle of the control.
	NativeControl() native.Handle
}

// ControlElementHelper is an [Element] that represents a native GUI control.
// This type can be used to implement [ControlElement] for native control widgets.
type ControlElementHelper struct {
	ElementHelper
	Handle native.Handle
	// DestroyFunc is called to destroy the native handle.
	// A nil value means no special destruction is needed.
	DestroyFunc func(native.Handle) error
}

// NativeControl implements the [ControlElement.NativeControl] method.
func (e *ControlElementHelper) NativeControl() native.Handle {
	return e.Handle
}

// Destroy implements the [ControlElement.Destroy] method.
func (e *ControlElementHelper) Destroy() error {
	if e.DestroyFunc != nil {
		err := e.DestroyFunc(e.Handle)
		if err != nil {
			return err
		}
	}
	return nil
}

// buildElementTree builds the element tree for the given widget.
// The returned layouter is the layouter of the returned element or its nearest child.
func buildElementTree(ctx *Context, widget AbstractWidget) (element Element, layouter Layouter, err error) {
	element, err = buildElementTreeImpl(ctx, nil, widget)
	if err != nil {
		return
	}
	layouter = layouterTree(element)
	return
}

// BuildElementTree builds the element tree for the given widget.
// The returned element is the root element of the built tree.
func BuildElementTree(ctx *Context, widget AbstractWidget) (Element, error) {
	return buildElementTreeImpl(ctx, nil, widget)
}

// buildElementTreeImpl builds the element tree for the given widget.
// The parent element can be nil. If parent is not nil, the ancestor
// tree of parent must already be established.
// The returned element is the root element of the built tree.
func buildElementTreeImpl(ctx *Context, parent Element, widget AbstractWidget) (Element, error) {
	elem, err := widget.CreateElement(ctx, parent)
	if err != nil {
		return nil, err
	}

	if parent != nil {
		// Appending elem to parent before elem is fully constructed because
		// CreateElement method of native widget looks for its native parent
		// handle up the element tree. See [NativeControl] function.
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

	if err := elem.SetWidget(ctx, widget); err != nil {
		return nil, err
	}

	if statefulWidget, ok := widget.(AbstractStatefulWidget); ok {
		return buildStatefulElement(ctx, elem, statefulWidget)
	}
	if statelessWidget, ok := widget.(AbstractStatelessWidget); ok {
		return buildStatelessElement(ctx, elem, statelessWidget)
	}
	if container, ok := widget.(AbstractContainer); ok {
		return buildContainerElement(ctx, elem, container)
	}
	return elem, nil
}

func buildContainerElement(ctx *Context, elem Element, container AbstractContainer) (Element, error) {
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

func buildStatelessElement(ctx *Context, elem Element, statelessWidget AbstractStatelessWidget) (Element, error) {
	// The child element is already appended to elem in buildElementTreeImpl.
	_, err := buildElementTreeImpl(ctx, elem, statelessWidget.Build(ctx))
	if err != nil {
		return nil, err
	}
	return elem, nil
}

func buildStatefulElement(ctx *Context, elem Element, statefulWidget AbstractStatefulWidget) (Element, error) {
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
func updateElementTree(ctx *Context, elem Element, widget AbstractWidget) (err error) {
	if err = elem.SetWidget(ctx, widget); err != nil {
		return
	}
	if container, ok := widget.(AbstractContainer); ok {
		return updateContainerElement(ctx, elem, container)
	}
	if _, ok := widget.(AbstractStatefulWidget); ok {
		return updateStatefulWidget(ctx, elem)
	}
	if statelessWidget, ok := widget.(AbstractStatelessWidget); ok {
		return updateStatelessWidget(ctx, elem, statelessWidget)
	}
	return nil
}

// updateStatelessWidget updates the stateless element elem to hold the new stateless widget.
func updateStatelessWidget(ctx *Context, elem Element, statelessWidget AbstractStatelessWidget) error {
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
func updateContainerElement(ctx *Context, element Element, container AbstractContainer) error {
	var newChildren = make([]Element, container.NumChildren()) // the updated children

	numElem := element.NumChildren()
	numWidget := container.NumChildren()

	// Phase 1: Top-down match

	var topDownCount = 0 // number of matched elements(widgets) from the top
	for i := 0; i < min(numElem, numWidget); i++ {
		widget := container.Child(i)
		elem := element.Child(i)
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
		elem := element.Child(elemIndex)
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
			elem := element.Child(i)
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
				if err = matchedElem.Destroy(); err != nil {
					return err
				}
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
		if err := unmatched.Destroy(); err != nil {
			return err
		}
	}
	// Destroy unused old elements
	for _, unusedElem := range unusedElements {
		if err := unusedElem.Destroy(); err != nil {
			return err
		}
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
func widgetMatch(widget1, widget2 AbstractWidget) bool {
	return widget1.WidgetID() == widget2.WidgetID() && reflect.TypeOf(widget1) == reflect.TypeOf(widget2)
}

// reconcileElementTreeImpl performs the actual reconciliation.
// It recreates the element tree if the widgets do not match, or updates it in place if they match.
// The reconciled element and any error occurred during the process are returned.
// If a new element is created, it is returned as the reconciled and the old element is not destroyed.
func reconcileElementTreeImpl(ctx *Context, element Element, widget AbstractWidget) (reconciled Element, err error) {
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
func reconciledChildElement(ctx *Context, parent Element, childIndex int, widget AbstractWidget) (err error) {
	oldChild := parent.Child(childIndex)
	newChild, err := reconcileElementTreeImpl(ctx, parent.Child(childIndex), widget)
	if err != nil {
		return err
	}
	if newChild != oldChild {
		return element_SetChild(parent, childIndex, newChild)
	}
	return
}
