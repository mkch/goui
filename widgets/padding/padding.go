package padding

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
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
	return 1
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

	for child := range l.Children() {
		childMaxWidth := layoututil.Clamp(constraints.MaxWidth-padding.Left-padding.Right, constraints.MinWidth, constraints.MaxWidth)
		childMaxHeight := layoututil.Clamp(constraints.MaxHeight-padding.Top-padding.Bottom, constraints.MinHeight, constraints.MaxHeight)
		childConstraints := goui.Constraints{
			MinWidth:  constraints.MinWidth,
			MaxWidth:  childMaxWidth,
			MinHeight: constraints.MinHeight,
			MaxHeight: childMaxHeight,
		}
		var childSize goui.Size
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		if err = layoututil.CheckOverflow(child.Element().Widget(), childSize, childConstraints); err != nil {
			return
		}

		size = goui.Size{
			Width:  layoututil.Clamp(childSize.Width+padding.Left+padding.Right, constraints.MinWidth, constraints.MaxWidth),
			Height: layoututil.Clamp(childSize.Height+padding.Top+padding.Bottom, constraints.MinHeight, constraints.MaxHeight),
		}
		return // only one child
	}
	return goui.Size{
		Width:  padding.Left + padding.Right,
		Height: padding.Top + padding.Bottom,
	}, nil
}

func (l *paddingLayouter) PositionAt(x, y int) (err error) {
	for child := range l.Children() {
		padding := l.Element().Widget().(*Padding)
		return child.PositionAt(x+padding.Left, y+padding.Top)
	}
	return
}
