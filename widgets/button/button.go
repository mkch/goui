package button

import (
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type Button struct {
	ID       goui.ID
	Label    string
	Padding  *metrics.Size // Padding around the label text. If nil, default padding is used.
	Disabled bool

	OnClick func(*goui.Context)
}

func (btn *Button) WidgetID() goui.ID {
	return btn.ID
}

func (btn *Button) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	parentHandle, err := goui.WidgetNativeParent(ctx, parent)
	if err != nil {
		err = errortrace.ErrorfStack("Create Button failed: %w", err)
		return
	}
	handle, err := native.CreateButton(parentHandle, btn.Label)
	if err != nil {
		err = errortrace.ErrorfStack("Create Button failed: %w", err)
		return
	}
	native.SetWidgetEnabled(handle, !btn.Disabled)
	layouter := &buttonLayouter{}
	element = &buttonElement{
		goui.ControlElementHelper{
			ElementHelper: goui.ElementHelper{
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
	return
}

func (*Button) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

type buttonElement struct {
	goui.ControlElementHelper
}

func (e *buttonElement) SetWidget(ctx *goui.Context, widget goui.AbstractWidget) (err error) {
	oldBtn, _ := e.Widget().(*Button)
	newBtn := widget.(*Button)

	if err = e.ControlElementHelper.SetWidget(ctx, widget); err != nil {
		return
	}

	if oldBtn != nil {
		if oldBtn.Label != newBtn.Label {
			if err = native.SetButtonLabel(e.Handle, newBtn.Label); err != nil {
				return
			}
		}
		if oldBtn.Disabled != newBtn.Disabled {
			if err = native.SetWidgetEnabled(e.Handle, !newBtn.Disabled); err != nil {
				return
			}
		}
	} else {
		if err = native.SetWidgetEnabled(e.Handle, !newBtn.Disabled); err != nil {
			return
		}
	}

	if newBtn.OnClick != nil {
		native.SetButtonOnClickListener(e.Handle, func() { newBtn.OnClick(ctx) })
	} else {
		native.SetButtonOnClickListener(e.Handle, nil)
	}

	return
}

type buttonLayouter struct {
	goui.LayouterHelper
	layoutSize metrics.Size
}

var defaultButtonPadding = metrics.Size{Width: 30, Height: 20}

func (l *buttonLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	elem := l.Element().(*buttonElement)
	if constraints.TightWidth() && constraints.TightHeight() {
		size = metrics.Size{
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
	size = constraints.Clamp(metrics.Size{Width: intrinsicWidth + padding.Width, Height: intrinsicHeight + padding.Height})
	l.layoutSize = size
	return
}

func (l *buttonLayouter) PositionAt(pt metrics.Point) (err error) {
	return native.SetWidgetDimensions(l.Element().(*buttonElement).Handle, pt.X, pt.Y, l.layoutSize.Width, l.layoutSize.Height)
}
