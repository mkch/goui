package visibility

import (
	"github.com/mkch/gg"
	"github.com/mkch/gg/errortrace"
	"github.com/mkch/goui"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/native"
)

// Visibility is a [Container] [Widget] that shows or hides its single child
// based on the Visible field. If not visible, the child is positioned beyond
// the right edge of the window. If MaintainSize is true, the invisible child
// still takes up space in layout.
// Menu and menu items are not allowed to be placed inside Visibility.
type Visibility struct {
	ID           goui.ID
	Widget       goui.Widget
	Visible      bool // Whether to show the child widget.
	MaintainSize bool // Whether to maintain the child's size when not visible.
}

func (p *Visibility) WidgetID() goui.ID {
	return p.ID
}

func (p *Visibility) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementHelper{
		ElementLayouter: &visibilityLayouter{},
	}, nil
}

func (p *Visibility) NumChildren() int {
	return gg.If(p.Widget != nil, 1, 0)
}

func (p *Visibility) Child(n int) goui.Component {
	return p.Widget
}

func (Visibility) ExclusiveType(marker.TypeWidget)     { /*Nop*/ }
func (*Visibility) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

type visibilityLayouter struct {
	goui.LayouterHelper
	// X offset of child.
	// 0 if visible. Beyond right edge of window if not visible.
	childXOffset metrics.DP
}

func (l *visibilityLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	elem := l.Element()
	visibility := elem.Widget().(*Visibility)
	for child := range l.Children() {
		if !visibility.Visible {
			// Set the offset beyond the right edge of the window.
			var parentHandle native.Handle
			parentHandle, err = goui.WidgetNativeParent(ctx, elem.Parent())
			if err != nil {
				err = errortrace.ErrorfStack("Layout Visibility failed: %w", err)
				return
			}
			var rect metrics.Rect
			if rect, err = goui.OS().Window_ClientRect(parentHandle); err != nil {
				err = errortrace.ErrorfStack("Layout Visibility failed: %w", err)
				return
			} else {
				l.childXOffset = rect.Width()
			}
			if visibility.MaintainSize {
				// Use the child's size.
				size, err = child.Layout(ctx, constraints)
			} else {
				// Use the minimum size.
				size = constraints.MinSize()
			}
			return
		}
		// Visible
		l.childXOffset = 0                    // Normal position
		return child.Layout(ctx, constraints) // Normal layout
	}
	return constraints.MinSize(), nil
}

func (l *visibilityLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
	for child := range l.Children() {
		// See Layout() for the offset logic.
		return child.PositionAt(ctx, metrics.Point{
			X: pt.X + l.childXOffset,
			Y: pt.Y})
	}
	return
}
