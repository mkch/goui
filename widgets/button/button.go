package button

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
	"github.com/mkch/goui/native"
)

type Button struct {
	ID      goui.ID
	Label   string
	Padding *goui.Size // Padding around the label text. If nil, default padding is used.
	OnClick func()
}

func (btn *Button) WidgetID() goui.ID {
	return btn.ID
}

func (btn *Button) CreateElement(ctx *goui.Context) (goui.Element, error) {
	handle, err := native.CreateButton(ctx.NativeWindow(), btn.Label)
	if err != nil {
		return nil, err
	}
	elem := &buttonElement{
		goui.NativeElement{
			ElementBase: goui.ElementBase{
				ElementLayouter: &buttonLayouter{},
			},
			Handle:      handle,
			DestroyFunc: native.DestroyWindow,
		},
	}
	native.SetButtonOnClickListener(handle, func() {
		if btn.OnClick != nil {
			btn.OnClick()
		}
	})
	return elem, nil
}

type buttonElement struct {
	goui.NativeElement
}

func (e *buttonElement) SetWidget(widget goui.Widget) {
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
			newBtn.OnClick()
		}
	})

	e.NativeElement.SetWidget(widget)
}

type buttonLayouter struct {
	goui.LayouterBase
}

var defaultButtonPadding = goui.Size{Width: 15, Height: 10}

func (l *buttonLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	elem := l.Element().(*buttonElement)
	defer func() {
		if err == nil {
			err = native.SetWidgetSize(elem.Handle, size.Width, size.Height)
		}
	}()
	if constraints.TightWidth() && constraints.TightHeight() {
		size = goui.Size{
			Width:  constraints.MinWidth,
			Height: constraints.MinHeight,
		}
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
	size = goui.Size{
		Width:  layoututil.Clamp(intrinsicWidth+padding.Width, constraints.MinWidth, constraints.MaxWidth),
		Height: layoututil.Clamp(intrinsicHeight+padding.Height, constraints.MinHeight, constraints.MaxHeight),
	}
	return
}

func (l *buttonLayouter) PositionAt(x, y int) (err error) {
	return native.SetWidgetPosition(l.Element().(*buttonElement).Handle, x, y)
}
