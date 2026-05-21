package listener

//go:generate stringer -type=PointerKind
//go:generate stringer -type=Button

import (
	"fmt"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

// PointerKind represents the type of pointer device.
type PointerKind int

const (
	Mouse PointerKind = iota
)

// ButtonMask represents pointer buttons.
type Button int

const (
	PrimaryMouseButton Button = iota + 1
	SecondaryMouseButton
	MiddleMouseButton
)

// PointerEvent represents a pointer event.
type PointerEvent struct {
	Kind   PointerKind
	Button Button
	Pos    metrics.Point // Listener-local position

	ctx            *goui.Context
	listenerOffset metrics.Point
	nativeParent   native.Handle
	nativeWindow   native.Handle
}

func (evt *PointerEvent) String() string {
	return fmt.Sprintf("PointerEvent{Kind:%s, Button:%s, Pos:%s}", evt.Kind, evt.Button, evt.Pos)
}

// WindowClientPos returns the position of the pointer event in the coordinate system of the top-level native window's client area.
func (evt *PointerEvent) WindowClientPos() (pos metrics.Point, err error) {
	x, y, err := goui.OS().Util_ClientCoordinatesConv(evt.nativeParent, evt.nativeWindow,
		evt.listenerOffset.X+evt.Pos.X, evt.listenerOffset.Y+evt.Pos.Y)
	if err != nil {
		return
	}
	pos = metrics.Point{X: x, Y: y}
	return
}

// Listener is a widget that listens for pointer events.
// Listener stays as large as its child widget.
type Listener struct {
	ID     goui.ID
	Widget goui.Widget

	OnPointerDown func(ctx *goui.Context, event *PointerEvent)
	OnPointerUp   func(ctx *goui.Context, event *PointerEvent)
	OnPointerMove func(ctx *goui.Context, event *PointerEvent)
}

// WidgetID implements [goui.Widget.ID]
func (l *Listener) WidgetID() goui.ID {
	return l.ID
}

// NumChildren implements [goui.Container.NumChildren]
func (l *Listener) NumChildren() int {
	return gg.If(l.Widget != nil, 1, 0)
}

// Child implements [goui.Container.Child]
func (l *Listener) Child(index int) goui.Component {
	if l.Widget == nil || index != 0 {
		panic("index out of range")
	}
	return l.Widget
}

func (*Listener) ExclusiveType(marker.TypeWidget)    { /*NOP*/ }
func (*Listener) ExclusiveKind(marker.KindContainer) { /*NOP*/ }

// CreateElement implements [goui.Widget.CreateElement]
func (l *Listener) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	return &listenerElement{
		ElementHelper: goui.ElementHelper{ElementLayouter: &listenerLayouter{}},
	}, nil
}

type listenerElement struct {
	goui.ElementHelper
	offset metrics.Point // offset within native parent
	size   metrics.Size  // size of the element
	ctx    *goui.Context

	remove func() // removal function from native listeners
}

// SetWidget implements [goui.Element.SetWidget]
func (e *listenerElement) SetWidget(ctx *goui.Context, widget goui.Component) error {
	if e.Widget() == nil {
		// One time setup.
		e.ctx = ctx
		parent, err := goui.WidgetNativeParent(ctx, e.Parent())
		if err != nil {
			return errortrace.ErrorfStack("configure Listener failed: %w", err)
		}
		e.remove = goui.OS().App_AddMouseEventListener(parent, e)
	}
	return e.ElementHelper.SetWidget(ctx, widget)
}

// Destroy implements [goui.Element.Destroy]
func (e *listenerElement) Destroy(ctx *goui.Context) error {
	if err := e.ElementHelper.Destroy(ctx); err != nil {
		return err
	}
	e.remove()
	return nil
}

func (e *listenerElement) callPointerEventMethod(method func(ctx *goui.Context, event *PointerEvent),
	button Button, parent native.Handle, x, y metrics.DP) {
	if method != nil &&
		x >= e.offset.X && x < e.offset.X+e.size.Width &&
		y >= e.offset.Y && y < e.offset.Y+e.size.Height {
		method(e.ctx, &PointerEvent{
			Kind:           Mouse,
			Button:         button,
			Pos:            metrics.Point{X: x - e.offset.X, Y: y - e.offset.Y},
			ctx:            e.ctx,
			listenerOffset: e.offset,
			nativeParent:   parent,
			nativeWindow:   e.ctx.NativeWindow(),
		})
	}
}

func (e *listenerElement) OnMousePrimaryDown(parent native.Handle, x, y metrics.DP) {
	e.callPointerEventMethod(e.Widget().(*Listener).OnPointerDown, PrimaryMouseButton, parent, x, y)
}

func (e *listenerElement) OnMousePrimaryUp(parent native.Handle, x, y metrics.DP) {
	e.callPointerEventMethod(e.Widget().(*Listener).OnPointerUp, PrimaryMouseButton, parent, x, y)
}

func (e *listenerElement) OnMouseSecondaryDown(parent native.Handle, x, y metrics.DP) {
	e.callPointerEventMethod(e.Widget().(*Listener).OnPointerDown, SecondaryMouseButton, parent, x, y)
}

func (e *listenerElement) OnMouseSecondaryUp(parent native.Handle, x, y metrics.DP) {
	e.callPointerEventMethod(e.Widget().(*Listener).OnPointerUp, SecondaryMouseButton, parent, x, y)
}

func (e *listenerElement) OnMouseMiddleDown(parent native.Handle, x, y metrics.DP) {
	e.callPointerEventMethod(e.Widget().(*Listener).OnPointerDown, MiddleMouseButton, parent, x, y)
}

func (e *listenerElement) OnMouseMiddleUp(parent native.Handle, x, y metrics.DP) {
	e.callPointerEventMethod(e.Widget().(*Listener).OnPointerUp, MiddleMouseButton, parent, x, y)
}

func (e *listenerElement) OnMousePointerMove(parent native.Handle, x, y metrics.DP) {
	e.callPointerEventMethod(e.Widget().(*Listener).OnPointerMove, 0, parent, x, y)
}

type listenerLayouter struct {
	goui.LayouterHelper
}

// Layout implements [goui.Layouter.Layout]
func (l *listenerLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	defer func() {
		l.Element().(*listenerElement).size = size
	}()
	for child := range l.Children() {
		return child.Layout(ctx, constraints) // Use child size
	}
	return constraints.MinSize(), nil // No child, take minimum size
}

func (l *listenerLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) error {
	l.Element().(*listenerElement).offset = pt
	for child := range l.Children() {
		return child.PositionAt(ctx, pt)
	}
	return nil
}
