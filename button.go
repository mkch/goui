package goui

import "github.com/mkch/goui/native"

type Button struct {
	ID      ID
	Label   string
	Padding *Size // Padding around the label text. If nil, default padding is used.
	OnClick func()
}

func (btn *Button) WidgetID() ID {
	return btn.ID
}

func (btn *Button) CreateElement(ctx *Context) (Element, error) {
	handle, err := native.CreateButton(ctx.window.Handle, btn.Label)
	if err != nil {
		return nil, err
	}
	elem := &buttonElement{
		nativeElement{
			layouter:    &buttonLayouter{},
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
	nativeElement
}

func (e *buttonElement) setWidget(widget Widget) {
	newBtn := widget.(*Button)
	if oldWidget := e.widget(); oldWidget != nil {
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

	e.nativeElement.setWidget(widget)
}

type buttonLayouter struct {
	LayouterBase
}

var defaultButtonPadding = Size{Width: 15, Height: 10}

func (l *buttonLayouter) Layout(ctx *Context, constraints Constraints) (size Size, err error) {
	l.LayouterBase.Layout(ctx, constraints)
	if constraints.Tight() {
		l.size = Size{
			Width:  constraints.MinWidth,
			Height: constraints.MinHeight,
		}
		return l.size, nil
	}
	elem := l.element().(*buttonElement)
	widget := elem.widget().(*Button)
	padding := widget.Padding
	if padding == nil {
		padding = &defaultButtonPadding
	}
	intrinsicWidth, intrinsicHeight, err := native.GetButtonMinimumSize(elem.Handle, widget.Label)
	if err != nil {
		return
	}
	l.size = Size{
		Width:  clampInt(intrinsicWidth+padding.Width, constraints.MinWidth, constraints.MaxWidth),
		Height: clampInt(intrinsicHeight+padding.Height, constraints.MinHeight, constraints.MaxHeight),
	}
	return l.size, nil
}

func (l *buttonLayouter) PositionAt(x, y int) (err error) {
	l.LayouterBase.PositionAt(x, y)
	return native.SetWidgetDimensions(l.element().(*buttonElement).Handle, x, y, l.size.Width, l.size.Height)
}
