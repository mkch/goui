package textfield

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
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
	handle, err := ctx.App().OS().NewTextField(ctx.NativeWindow(), txt.InitialValue, txt.Obscure)
	if err != nil {
		return nil, err
	}

	layouter := &textFieldLayouter{}
	elem := &textFieldElement{
		goui.ControlElementHelper{
			ElementHelper: goui.ElementHelper{
				ElementLayouter: layouter,
			},
			Handle: handle,
			DestroyFunc: func(ctx *goui.Context, h native.Handle) error {
				if txt.Controller != nil {
					txt.Controller.setElement(ctx, nil)
				}
				return ctx.App().OS().Control_Destroy(h)
			},
		},
	}
	return elem, nil
}

func (*TextField) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

type textFieldElement struct {
	goui.ControlElementHelper
}

func (e *textFieldElement) SetWidget(ctx *goui.Context, widget goui.Component) (err error) {
	newWidget := widget.(*TextField)

	if err = e.ControlElementHelper.SetWidget(ctx, widget); err != nil {
		return
	}

	if newWidget.Controller != nil {
		newWidget.Controller.setElement(ctx, e)
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

func (l *textFieldLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
	return ctx.App().OS().Control_SetDimensions(l.Element().(*textFieldElement).Handle, metrics.NewRect(pt, l.layoutSize))
}
