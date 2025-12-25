package label

import (
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
	"github.com/mkch/goui/native"
)

type Label struct {
	ID      goui.ID
	Text    string
	Padding *goui.Size // Padding around the label text. If nil, no padding is applied.
}

func (btn *Label) WidgetID() goui.ID {
	return btn.ID
}

func (btn *Label) CreateElement(ctx *goui.Context) (goui.Element, error) {
	handle, err := native.CreateLabel(ctx.NativeWindow(), btn.Text)
	if err != nil {
		return nil, err
	}
	layouter := &labelLayouter{}
	elem := &labelElement{
		goui.NativeElement{
			ElementBase: goui.ElementBase{
				ElementLayouter: layouter,
			},
			Handle:      handle,
			DestroyFunc: native.DestroyWindow,
		},
	}
	return elem, nil
}

type labelElement struct {
	goui.NativeElement
}

func (e *labelElement) SetWidget(ctx *goui.Context, widget goui.Widget) {
	newLabel := widget.(*Label)
	if oldWidget := e.Widget(); oldWidget != nil {
		oldLabel := oldWidget.(*Label)
		if oldLabel.Text != newLabel.Text {
			if err := native.SetLabelText(e.Handle, newLabel.Text); err != nil {
				errortrace.Panic(err)
			}
		}
	}
	e.NativeElement.SetWidget(ctx, widget)
}

type labelLayouter struct {
	goui.LayouterBase
	layoutSize goui.Size
}

func (l *labelLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	elem := l.Element().(*labelElement)
	if constraints.TightWidth() && constraints.TightHeight() {
		size = goui.Size{
			Width:  constraints.MinWidth,
			Height: constraints.MinHeight,
		}
		l.layoutSize = size
		return
	}
	widget := elem.Widget().(*Label)
	padding := widget.Padding
	if padding == nil {
		padding = &goui.Size{Width: 0, Height: 0}
	}
	intrinsicWidth, intrinsicHeight, err := native.GetTextDrawingSize(elem.Handle, widget.Text, false)
	if err != nil {
		return
	}
	size = goui.Size{
		Width:  layoututil.Clamp(intrinsicWidth+padding.Width, constraints.MinWidth, constraints.MaxWidth),
		Height: layoututil.Clamp(intrinsicHeight+padding.Height, constraints.MinHeight, constraints.MaxHeight),
	}
	l.layoutSize = size
	return
}

func (l *labelLayouter) PositionAt(x, y int) (err error) {
	return native.SetWidgetDimensions(l.Element().(*labelElement).Handle, x, y, l.layoutSize.Width, l.layoutSize.Height)
}
