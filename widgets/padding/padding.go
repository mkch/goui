package padding

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
)

type Padding struct {
	ID                       goui.ID
	Widget                   goui.Widget
	Left, Top, Right, Bottom int
}

func (p *Padding) WidgetID() goui.ID {
	return p.ID
}

func (p *Padding) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &goui.ElementBase{
		ElementLayouter: &paddingLayouter{},
	}, nil
}

func (p *Padding) NumChildren() int {
	return gg.If(p.Widget != nil, 1, 0)
}

func (p *Padding) Child(n int) goui.Widget {
	return p.Widget
}

func (p *Padding) Exclusive(goui.Container) { /*Nop*/ }

type paddingLayouter struct {
	goui.LayouterBase
}

func (l *paddingLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	padding := l.Element().Widget().(*Padding)

	var childSize goui.Size
	for child := range l.Children() {
		childMaxWidth := constraints.ClampWidth(constraints.MaxWidth - padding.Left - padding.Right)
		childMaxHeight := constraints.ClampHeight(constraints.MaxHeight - padding.Top - padding.Bottom)
		childConstraints := goui.Constraints{
			MinWidth:  constraints.MinWidth,
			MaxWidth:  childMaxWidth,
			MinHeight: constraints.MinHeight,
			MaxHeight: childMaxHeight,
		}
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		if err = debug.CheckLayoutOverflow(ctx, child.Element().Widget(), childSize, childConstraints); err != nil {
			return
		}
		break // only one child
	}
	size = constraints.Clamp(goui.Size{
		Width:  childSize.Width + padding.Left + padding.Right,
		Height: childSize.Height + padding.Top + padding.Bottom})
	return
}

func (l *paddingLayouter) PositionAt(x, y int) (err error) {
	padding := l.Element().Widget().(*Padding)
	for child := range l.Children() {
		return child.PositionAt(x+padding.Left, y+padding.Top)
	}
	return
}
