package label

import (
	"image/color"

	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
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
	Padding         *goui.Size   // Padding around the label text. If nil, no padding is applied.
	BackgroundColor *color.NRGBA // Background color of the label. If nil, default is used.
}

func (btn *Label) WidgetID() goui.ID {
	return btn.ID
}

func (btn *Label) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	handle, err := native.CreateLabel(goui.NativeControl(ctx, parent), btn.Text)
	if err != nil {
		return nil, err
	}
	layouter := &labelLayouter{}
	elem := &labelElement{
		goui.ControlElementBase{
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
	goui.ControlElementBase
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
		if oldLabel.BackgroundColor != newLabel.BackgroundColor ||
			oldLabel.BackgroundColor != nil && *newLabel.BackgroundColor != *oldLabel.BackgroundColor {
			if err := native.SetLabelBackgroundColor(e.Handle, newLabel.BackgroundColor); err != nil {
				errortrace.Panic(err)
			}
		}
		if oldLabel.Multiline != newLabel.Multiline {
			if err := native.SetLabelMultiline(e.Handle, newLabel.Multiline); err != nil {
				errortrace.Panic(err)
			}
		}
		if oldLabel.TextAlignment != newLabel.TextAlignment {
			if err := native.SetLabelTextAlignment(e.Handle, native.TextAlignment(newLabel.TextAlignment)); err != nil {
				errortrace.Panic(err)
			}
		}
		return
	}

	if newLabel.BackgroundColor != nil {
		if err := native.SetLabelBackgroundColor(e.Handle, newLabel.BackgroundColor); err != nil {
			errortrace.Panic(err)
		}
	}

	if err := native.SetLabelMultiline(e.Handle, newLabel.Multiline); err != nil {
		errortrace.Panic(err)
	}

	if err := native.SetLabelTextAlignment(e.Handle, native.TextAlignment(newLabel.TextAlignment)); err != nil {
		errortrace.Panic(err)
	}

	e.ControlElementBase.SetWidget(ctx, widget)
}

type labelLayouter struct {
	goui.LayouterHelper
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
	intrinsicWidth, intrinsicHeight, err := native.GetTextDrawingSize(elem.Handle, widget.Text, widget.Multiline, constraints.MaxWidth-padding.Width)
	if err != nil {
		return
	}
	size = constraints.Clamp(goui.Size{Width: intrinsicWidth + padding.Width, Height: intrinsicHeight + padding.Height})
	l.layoutSize = size
	return
}

func (l *labelLayouter) PositionAt(x, y metrics.DP) (err error) {
	return native.SetWidgetDimensions(l.Element().(*labelElement).Handle, x, y, l.layoutSize.Width, l.layoutSize.Height)
}
