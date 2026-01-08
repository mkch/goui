package textfield

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type TextField struct {
	ID           goui.ID
	Controller   *Controller
	InitialValue string
	Obscure      bool // If true, the text field will obscure the input (e.g., for passwords).
}

func (txt *TextField) WidgetID() goui.ID {
	return txt.ID
}

func (txt *TextField) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	handle, err := native.CreateTextField(ctx.NativeWindow(), txt.InitialValue, txt.Obscure)
	if err != nil {
		return nil, err
	}

	layouter := &textFieldLayouter{}
	elem := &textFieldElement{
		goui.ControlElementBase{
			ElementBase: goui.ElementBase{
				ElementLayouter: layouter,
			},
			Handle: handle,
			DestroyFunc: func(h native.Handle) error {
				if txt.Controller != nil {
					txt.Controller.setElement(nil)
				}
				return native.DestroyWindow(h)
			},
		},
	}
	return elem, nil
}

type textFieldElement struct {
	goui.ControlElementBase
}

func (e *textFieldElement) SetWidget(ctx *goui.Context, widget goui.Widget) (err error) {
	newWidget := widget.(*TextField)

	if err = e.ControlElementBase.SetWidget(ctx, widget); err != nil {
		return
	}

	if newWidget.Controller != nil {
		newWidget.Controller.setElement(e)
	}
	return
}

type textFieldLayouter struct {
	goui.LayouterHelper
	layoutSize metrics.Size
}

func (l *textFieldLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	if constraints.TightWidth() && constraints.TightHeight() {
		size = constraints.MinSize()
		l.layoutSize = size
		return
	}
	intrinsicWidth, intrinsicHeight := metrics.DP(330), metrics.DP(50) // Default size for text field
	size = constraints.Clamp(metrics.Size{Width: intrinsicWidth, Height: intrinsicHeight})
	l.layoutSize = size
	return
}

func (l *textFieldLayouter) PositionAt(pt metrics.Point) (err error) {
	return native.SetWidgetDimensions(l.Element().(*textFieldElement).Handle, pt.X, pt.Y, l.layoutSize.Width, l.layoutSize.Height)
}
