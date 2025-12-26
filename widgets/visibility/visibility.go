package visibility

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/native"
)

// Visibility is a [Container] [Widget] that shows or hides its single child
// based on the Visible field. If not visible, the child is positioned beyond
// the right edge of the window. If MaintainSize is true, the invisible child
// still takes up space in layout.
type Visibility struct {
	ID           goui.ID
	Widget       goui.Widget
	Visible      bool // Whether to show the child widget.
	MaintainSize bool // Whether to maintain the child's size when not visible.
}

func (p *Visibility) WidgetID() goui.ID {
	return p.ID
}

func (p *Visibility) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &goui.ElementBase{
		ElementLayouter: &visibilityLayouter{},
	}, nil
}

func (p *Visibility) NumChildren() int {
	return gg.If(p.Widget != nil, 1, 0)
}

func (p *Visibility) Child(n int) goui.Widget {
	return p.Widget
}

func (p *Visibility) Exclusive(goui.Container) { /*Nop*/ }

type visibilityLayouter struct {
	goui.LayouterBase
	// X offset of child.
	// 0 if visible. Beyond right edge of window if not visible.
	childXOffset int
}

func (l *visibilityLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	visibility := l.Element().Widget().(*Visibility)
	for child := range l.Children() {
		if !visibility.Visible {
			// Set the offset beyond the right edge of the window.
			if _, _, l.childXOffset, _, err = native.WindowClientRect(ctx.NativeWindow()); err != nil {
				return
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

func (l *visibilityLayouter) PositionAt(x, y int) (err error) {
	for child := range l.Children() {
		// See Layout() for the offset logic.
		return child.PositionAt(x+l.childXOffset, y)
	}
	return
}
