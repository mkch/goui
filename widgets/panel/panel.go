package panel

import (
	"image/color"

	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
	"github.com/mkch/goui/marker"
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
	parentHandle, err := goui.WidgetNativeParent(ctx, parent)
	if err != nil {
		err = errortrace.ErrorfStack("Create Panel failed: %w", err)
		return
	}
	handle, err := goui.OS().NewPanel(parentHandle)
	if err != nil {
		err = errortrace.ErrorfStack("Create Panel failed: %w", err)
		return
	}
	element = &panelElement{
		goui.ControlElementHelper{
			ElementHelper: goui.ElementHelper{
				ElementLayouter: &panelLayouter{},
			},
			Handle:      handle,
			DestroyFunc: func(ctx *goui.Context, handle native.Handle) error { return goui.OS().Control_Destroy(handle) },
		},
	}
	return
}

func (p *Panel) NumChildren() int {
	return gg.If(p.Widget != nil, 1, 0)
}

func (p *Panel) Child(n int) goui.AbstractWidget {
	return p.Widget
}

func (*Panel) ExclusiveType(marker.TypeWidget)    { /*Nop*/ }
func (*Panel) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

type panelElement struct {
	goui.ControlElementHelper
}

func (elem *panelElement) SetWidget(ctx *goui.Context, widget goui.AbstractWidget) (err error) {
	oldPanel, _ := elem.Widget().(*Panel)
	newPanel := widget.(*Panel)

	if err = elem.ElementHelper.SetWidget(ctx, widget); err != nil {
		return
	}

	if oldPanel != nil {
		if oldPanel.BackgroundColor != newPanel.BackgroundColor {
			if err = goui.OS().Panel_SetBackgroundColor(elem.Handle, newPanel.BackgroundColor); err != nil {
				return
			}
		}
		return
	}
	return goui.OS().Panel_SetBackgroundColor(elem.Handle, newPanel.BackgroundColor)
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

func (l *panelLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
	if err = goui.OS().Control_SetDimensions(l.Element().(*panelElement).Handle, metrics.NewRect(pt, l.layoutSize)); err != nil {
		return
	}
	for child := range l.Children() {
		return child.PositionAt(ctx, metrics.Point{X: 0, Y: 0}) // Position child at (0,0) relative to panel.
	}
	return
}
