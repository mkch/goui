package label

import (
	"image/color"

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

type TextAlignment int

// Sync with native.TextAlignment

const (
	Left TextAlignment = iota
	Right
	Center
)

// Label is a widget that displays a text string.
type Label struct {
	ID              goui.ID
	Text            string
	Multiline       bool // If true, the label supports multiple lines of text.
	TextAlignment   TextAlignment
	BackgroundColor *color.NRGBA // Background color of the label. If nil, default is used.
}

func (btn *Label) WidgetID() goui.ID {
	return btn.ID
}

func (btn *Label) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	parentHandle, err := goui.WidgetNativeParent(ctx, parent)
	if err != nil {
		err = errortrace.ErrorfStack("Create Label failed: %w", err)
		return
	}
	handle, err := goui.OS().NewLabel(parentHandle, btn.Text)
	if err != nil {
		err = errortrace.ErrorfStack("Create Label failed: %w", err)
		return
	}
	layouter := &labelLayouter{}
	element = &labelElement{
		goui.ControlElementHelper{
			ElementHelper: goui.ElementHelper{
				ElementLayouter: layouter,
			},
			Handle:      handle,
			DestroyFunc: func(ctx *goui.Context, handle native.Handle) error { return goui.OS().Control_Destroy(handle) },
		},
	}
	return
}

func (*Label) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

type labelElement struct {
	goui.ControlElementHelper
}

func (e *labelElement) SetWidget(ctx *goui.Context, widget goui.Component) (err error) {
	oldLabel, _ := e.Widget().(*Label)
	newLabel := widget.(*Label)

	if err = e.ControlElementHelper.SetWidget(ctx, widget); err != nil {
		return
	}

	if oldLabel != nil {
		if oldLabel.Text != newLabel.Text {
			if err = goui.OS().Label_SetText(e.Handle, newLabel.Text); err != nil {
				return
			}
		}
		if oldLabel.BackgroundColor != newLabel.BackgroundColor &&
			(oldLabel.BackgroundColor == nil || newLabel.BackgroundColor == nil || *oldLabel.BackgroundColor != *newLabel.BackgroundColor) {
			if err = goui.OS().Label_SetBackgroundColor(e.Handle, newLabel.BackgroundColor); err != nil {
				return
			}
		}
		if oldLabel.Multiline != newLabel.Multiline {
			if err = goui.OS().Label_SetMultiline(e.Handle, newLabel.Multiline); err != nil {
				return
			}
		}
		if oldLabel.TextAlignment != newLabel.TextAlignment {
			if err = goui.OS().Label_SetTextAlignment(e.Handle, native.TextAlignment(newLabel.TextAlignment)); err != nil {
				return
			}
		}
		return
	}

	if err = goui.OS().Label_SetBackgroundColor(e.Handle, newLabel.BackgroundColor); err != nil {
		return
	}

	if err = goui.OS().Label_SetMultiline(e.Handle, newLabel.Multiline); err != nil {
		return
	}

	return goui.OS().Label_SetTextAlignment(e.Handle, native.TextAlignment(newLabel.TextAlignment))
}

type labelLayouter struct {
	goui.LayouterHelper
	layoutSize metrics.Size
}

func (l *labelLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	elem := l.Element().(*labelElement)
	if constraints.TightWidth() && constraints.TightHeight() {
		size = metrics.Size{
			Width:  constraints.MinWidth,
			Height: constraints.MinHeight,
		}
		l.layoutSize = size
		return
	}
	widget := elem.Widget().(*Label)
	intrinsicSize, err := goui.OS().Control_TextDrawingSize(elem.Handle, widget.Text, widget.Multiline, constraints.MaxWidth)
	if err != nil {
		return
	}
	size = constraints.Clamp(intrinsicSize)
	l.layoutSize = size
	return
}

func (l *labelLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
	return goui.OS().Control_SetDimensions(l.Element().(*labelElement).Handle, metrics.NewRect(pt, l.layoutSize))
}
