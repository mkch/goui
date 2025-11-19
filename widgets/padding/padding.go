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
	return &paddingElement{}, nil
}

func (p *Padding) NumChildren() int {
	return 1
}

func (p *Padding) Child(n int) goui.Widget {
	return p.Widget
}

type paddingElement struct {
	goui.ElementBase
	layouter paddingLayouter
}

func (e *paddingElement) ElementLayouter() goui.Layouter {
	return &e.layouter
}

type paddingLayouter struct {
	goui.LayouterBase
}

func (l *paddingLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	padding := l.Element().(*paddingElement).Widget().(*Padding)
	if l.NumChildren() == 0 {
		return goui.Size{
			Width:  padding.Left + padding.Right,
			Height: padding.Top + padding.Bottom,
		}, nil
	}

	childMaxWidth := layoututil.Clamp(constraints.MaxWidth-padding.Left-padding.Right, constraints.MinWidth, constraints.MaxWidth)
	childMaxHeight := layoututil.Clamp(constraints.MaxHeight-padding.Top-padding.Bottom, constraints.MinHeight, constraints.MaxHeight)
	childConstraints := goui.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  childMaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: childMaxHeight,
	}
	var childSize goui.Size
	childSize, err = l.Child(0).Layout(ctx, childConstraints)
	if err != nil {
		return
	}
	if err = layoututil.CheckOverflow(l.Child(0).Element().Widget(), childSize, childConstraints); err != nil {
		return
	}

	size = goui.Size{
		Width:  layoututil.Clamp(childSize.Width+padding.Left+padding.Right, constraints.MinWidth, constraints.MaxWidth),
		Height: layoututil.Clamp(childSize.Height+padding.Top+padding.Bottom, constraints.MinHeight, constraints.MaxHeight),
	}
	return
}

func (l *paddingLayouter) PositionAt(x, y int) (err error) {
	if l.NumChildren() == 0 {
		return
	}
	padding := l.Element().(*paddingElement).Widget().(*Padding)
	return l.Child(0).PositionAt(x+padding.Left, y+padding.Top)
}
