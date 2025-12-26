package button

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/native"
)

type Button struct {
	ID      goui.ID
	Label   string
	Padding *goui.Size // Padding around the label text. If nil, default padding is used.
	OnClick func(*goui.Context)
}

func (btn *Button) WidgetID() goui.ID {
	return btn.ID
}

func (btn *Button) CreateElement(ctx *goui.Context) (goui.Element, error) {
	handle, err := native.CreateButton(ctx.NativeWindow(), btn.Label)
	if err != nil {
		return nil, err
	}
	layouter := &buttonLayouter{}
	elem := &buttonElement{
		goui.NativeElement{
			ElementBase: goui.ElementBase{
				ElementLayouter: layouter,
			},
			Handle:      handle,
			DestroyFunc: native.DestroyWindow,
		},
	}
	native.SetButtonOnClickListener(handle, func() {
		if btn.OnClick != nil {
			btn.OnClick(ctx)
		}
	})
	return elem, nil
}

type buttonElement struct {
	goui.NativeElement
}

func (e *buttonElement) SetWidget(ctx *goui.Context, widget goui.Widget) {
	newBtn := widget.(*Button)
	if oldWidget := e.Widget(); oldWidget != nil {
		oldBtn := oldWidget.(*Button)
		if oldBtn.Label != newBtn.Label {
			native.SetButtonLabel(e.Handle, newBtn.Label)
		}
	}
	// func type are not comparable, so we always reset the OnClick listener.
	native.SetButtonOnClickListener(e.Handle, func() {
		if newBtn.OnClick != nil {
			newBtn.OnClick(ctx)
		}
	})

	e.NativeElement.SetWidget(ctx, widget)
}

type buttonLayouter struct {
	goui.LayouterBase
	layoutSize goui.Size
}

var defaultButtonPadding = goui.Size{Width: 15, Height: 10}

func (l *buttonLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	elem := l.Element().(*buttonElement)
	if constraints.TightWidth() && constraints.TightHeight() {
		size = goui.Size{
			Width:  constraints.MinWidth,
			Height: constraints.MinHeight,
		}
		l.layoutSize = size
		return
	}
	widget := elem.Widget().(*Button)
	padding := widget.Padding
	if padding == nil {
		padding = &defaultButtonPadding
	}
	intrinsicWidth, intrinsicHeight, err := native.GetButtonMinimumSize(elem.Handle, widget.Label)
	if err != nil {
		return
	}
	size = constraints.Clamp(goui.Size{Width: intrinsicWidth + padding.Width, Height: intrinsicHeight + padding.Height})
	l.layoutSize = size
	return
}

func (l *buttonLayouter) PositionAt(x, y int) (err error) {
	return native.SetWidgetDimensions(l.Element().(*buttonElement).Handle, x, y, l.layoutSize.Width, l.layoutSize.Height)
}
