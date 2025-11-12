package goui

import "github.com/mkch/goui/native"

type Button struct {
	ID      ID
	Label   string
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
	size Size
}

func (l *buttonLayouter) Layout(ctx *Context, constraints Constraints) Size {
	l.size = Size{
		Width:  clampInt(200, constraints.MinWidth, constraints.MaxWidth),
		Height: clampInt(40, constraints.MinHeight, constraints.MaxHeight),
	}
	return l.size
}

func (l *buttonLayouter) Apply(x, y int) error {
	return native.SetWidgetDimensions(l.element().(*buttonElement).Handle, x, y, l.size.Width, l.size.Height)
}
