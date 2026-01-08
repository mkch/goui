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
	BackgroundColor *color.NRGBA // Background color of the label. If nil, default is used.
}

func (btn *Label) WidgetID() goui.ID {
	return btn.ID
}

func (btn *Label) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	parentHandle, err := goui.LookupNativeParent(ctx, parent)
	if err != nil {
		err = errortrace.ErrorfStack("Create Label failed: %w", err)
		return
	}
	handle, err := native.CreateLabel(parentHandle, btn.Text)
	if err != nil {
		err = errortrace.ErrorfStack("Create Label failed: %w", err)
		return
	}
	layouter := &labelLayouter{}
	element = &labelElement{
		goui.ControlElementBase{
			ElementBase: goui.ElementBase{
				ElementLayouter: layouter,
			},
			Handle:      handle,
			DestroyFunc: native.DestroyWindow,
		},
	}
	return
}

func (*Label) ExclusiveWidgetMenu(goui.Widget) { /*Nop*/ }

type labelElement struct {
	goui.ControlElementBase
}

func (e *labelElement) SetWidget(ctx *goui.Context, widget goui.WidgetBase) (err error) {
	oldLabel, _ := e.Widget().(*Label)
	newLabel := widget.(*Label)

	if err = e.ControlElementBase.SetWidget(ctx, widget); err != nil {
		return
	}

	if oldLabel != nil {
		if oldLabel.Text != newLabel.Text {
			if err = native.SetLabelText(e.Handle, newLabel.Text); err != nil {
				return
			}
		}
		if oldLabel.BackgroundColor != newLabel.BackgroundColor ||
			oldLabel.BackgroundColor != nil && *newLabel.BackgroundColor != *oldLabel.BackgroundColor {
			if err = native.SetLabelBackgroundColor(e.Handle, newLabel.BackgroundColor); err != nil {
				return
			}
		}
		if oldLabel.Multiline != newLabel.Multiline {
			if err = native.SetLabelMultiline(e.Handle, newLabel.Multiline); err != nil {
				return
			}
		}
		if oldLabel.TextAlignment != newLabel.TextAlignment {
			if err = native.SetLabelTextAlignment(e.Handle, native.TextAlignment(newLabel.TextAlignment)); err != nil {
				return
			}
		}
		return
	}

	if err = native.SetLabelBackgroundColor(e.Handle, newLabel.BackgroundColor); err != nil {
		return
	}

	if err = native.SetLabelMultiline(e.Handle, newLabel.Multiline); err != nil {
		return
	}

	return native.SetLabelTextAlignment(e.Handle, native.TextAlignment(newLabel.TextAlignment))
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
	intrinsicWidth, intrinsicHeight, err := native.GetTextDrawingSize(elem.Handle, widget.Text, widget.Multiline, constraints.MaxWidth)
	if err != nil {
		return
	}
	size = constraints.Clamp(metrics.Size{Width: intrinsicWidth, Height: intrinsicHeight})
	l.layoutSize = size
	return
}

func (l *labelLayouter) PositionAt(pt metrics.Point) (err error) {
	return native.SetWidgetDimensions(l.Element().(*labelElement).Handle, pt.X, pt.Y, l.layoutSize.Width, l.layoutSize.Height)
}
