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
	handle, err := goui.OS().NewButton(parentHandle, btn.Label)
	if err != nil {
		err = errortrace.ErrorfStack("Create Button failed: %w", err)
		return
	}
	goui.OS().Control_SetEnabled(handle, !btn.Disabled)
	layouter := &buttonLayouter{}
	element = &buttonElement{
		goui.ControlElementHelper{
			ElementHelper: goui.ElementHelper{
				ElementLayouter: layouter,
			},
			Handle:      handle,
			DestroyFunc: func(ctx *goui.Context, handle native.Handle) error { return goui.OS().Window_Destroy(handle) },
		},
	}
	goui.OS().Button_SetOnClickListener(handle, func() {
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
			if err = goui.OS().Button_SetLabel(e.Handle, newBtn.Label); err != nil {
				return
			}
		}
		if oldBtn.Disabled != newBtn.Disabled {
			if err = goui.OS().Control_SetEnabled(e.Handle, !newBtn.Disabled); err != nil {
				return
			}
		}
	} else {
		if err = goui.OS().Control_SetEnabled(e.Handle, !newBtn.Disabled); err != nil {
			return
		}
	}

	if newBtn.OnClick != nil {
		goui.OS().Button_SetOnClickListener(e.Handle, func() { newBtn.OnClick(ctx) })
	} else {
		goui.OS().Button_SetOnClickListener(e.Handle, nil)
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
	intrinsicSize, err := goui.OS().Button_MinimumSize(elem.Handle, widget.Label)
	if err != nil {
		return
	}
	size = constraints.Clamp(metrics.Size{Width: intrinsicSize.Width + padding.Width, Height: intrinsicSize.Height + padding.Height})
	l.layoutSize = size
	return
}

func (l *buttonLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
	return goui.OS().Control_SetDimensions(l.Element().(*buttonElement).Handle, metrics.NewRect(pt, l.layoutSize))
}
