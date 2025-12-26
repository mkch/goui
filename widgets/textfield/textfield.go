package textfield

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/native"
)

type TextField struct {
	ID           goui.ID
	Controller   *Controller
	InitialValue string
}

func (txt *TextField) WidgetID() goui.ID {
	return txt.ID
}

func (txt *TextField) CreateElement(ctx *goui.Context) (goui.Element, error) {
	handle, err := native.CreateTextField(ctx.NativeWindow(), txt.InitialValue)
	if err != nil {
		return nil, err
	}

	layouter := &textFieldLayouter{}
	elem := &textFieldElement{
		goui.NativeElement{
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
	if txt.Controller != nil {
		txt.Controller.setElement(elem)
	}
	return elem, nil
}

type textFieldElement struct {
	goui.NativeElement
}

func (e *textFieldElement) SetWidget(ctx *goui.Context, widget goui.Widget) {
	oldWidget := e.Widget()
	if oldWidget != widget {
		if newTextField := widget.(*TextField); newTextField.Controller != nil {
			newTextField.Controller.setElement(e)
		}
	}
	e.NativeElement.SetWidget(ctx, widget)
}

type textFieldLayouter struct {
	goui.LayouterBase
	layoutSize goui.Size
}

func (l *textFieldLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	if constraints.TightWidth() && constraints.TightHeight() {
		size = constraints.MinSize()
		l.layoutSize = size
		return
	}
	intrinsicWidth, intrinsicHeight := 200, 30 // Default size for text field
	size = constraints.Clamp(goui.Size{Width: intrinsicWidth, Height: intrinsicHeight})
	l.layoutSize = size
	return
}

func (l *textFieldLayouter) PositionAt(x, y int) (err error) {
	return native.SetWidgetDimensions(l.Element().(*textFieldElement).Handle, x, y, l.layoutSize.Width, l.layoutSize.Height)
}
