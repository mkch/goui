package panel

import (
	"image/color"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

// Panel is a simple container widget that uses its child's size as its own size.
type Panel struct {
	ID              goui.ID
	Widget          goui.Widget
	BackgroundColor *color.NRGBA
}

func (p *Panel) WidgetID() goui.ID {
	return p.ID
}

func (p *Panel) CreateElement(ctx *goui.Context, parent goui.Element) (element goui.Element, err error) {
	parentHandle, err := goui.LookupNativeParent(ctx, parent)
	if err != nil {
		err = errortrace.ErrorfStack("Create Panel failed: %w", err)
		return
	}
	handle, err := native.CreatePanel(parentHandle)
	if err != nil {
		err = errortrace.ErrorfStack("Create Panel failed: %w", err)
		return
	}
	element = &panelElement{
		goui.ControlElementBase{
			ElementBase: goui.ElementBase{
				ElementLayouter: &panelLayouter{},
			},
			Handle:      handle,
			DestroyFunc: native.DestroyWindow,
		},
	}
	return
}

func (p *Panel) NumChildren() int {
	return gg.If(p.Widget != nil, 1, 0)
}

func (p *Panel) Child(n int) goui.Widget {
	return p.Widget
}

func (p *Panel) Exclusive(goui.Container) { /*Nop*/ }

type panelElement struct {
	goui.ControlElementBase
}

func (elem *panelElement) SetWidget(ctx *goui.Context, widget goui.Widget) (err error) {
	oldPanel, _ := elem.Widget().(*Panel)
	newPanel := widget.(*Panel)

	if err = elem.ElementBase.SetWidget(ctx, widget); err != nil {
		return
	}

	if oldPanel != nil {
		if oldPanel.BackgroundColor != newPanel.BackgroundColor {
			if err = native.SetPanelBackgroundColor(elem.Handle, newPanel.BackgroundColor); err != nil {
				return
			}
		}
		return
	}
	return native.SetPanelBackgroundColor(elem.Handle, newPanel.BackgroundColor)
}

type panelLayouter struct {
	goui.LayouterHelper
	layoutSize metrics.Size
}

func (l *panelLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	defer func() { l.layoutSize = size }()
	for child := range l.Children() {
		size, err = child.Layout(ctx, constraints) // Use child's size as is.
		if err != nil {
			return
		}
		return size, debug.CheckLayoutOverflow(ctx, child.Element().Widget(), size, constraints)
	}
	return constraints.MinSize(), nil // Use min size when no child.
}

func (l *panelLayouter) PositionAt(x, y metrics.DP) (err error) {
	if err = native.SetWidgetDimensions(l.Element().(*panelElement).Handle, x, y, l.layoutSize.Width, l.layoutSize.Height); err != nil {
		return
	}
	for child := range l.Children() {
		return child.PositionAt(0, 0) // Position child at (0,0) relative to panel.
	}
	return
}
