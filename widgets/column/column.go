package column

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
)

// Column is a [Container] [Widget] that arranges its children in a vertical column.
type Column struct {
	ID      goui.ID
	Widgets []goui.Widget
}

func (c *Column) WidgetID() goui.ID {
	return c.ID
}

func (c *Column) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &columnElement{}, nil
}

func (c *Column) NumChildren() int {
	return len(c.Widgets)
}

func (c *Column) Child(n int) goui.Widget {
	return c.Widgets[n]
}

type columnElement struct {
	goui.ElementBase
}

func (e *columnElement) ElementLayouter() goui.Layouter {
	return &columnLayouter{}
}

type columnLayouter struct {
	goui.LayouterBase
	childOffsets []goui.Point
}

func (l *columnLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (goui.Size, error) {
	l.Size.Width = constraints.MaxWidth
	l.Size.Height = constraints.MaxHeight
	l.childOffsets = make([]goui.Point, l.NumChildren())
	var childrenHeight = 0
	for i := range l.NumChildren() {
		child := l.Child(i)
		childConstraints := goui.Constraints{
			MinWidth:  constraints.MinWidth,
			MinHeight: constraints.MinHeight,
			MaxWidth:  l.Size.Width,
			MaxHeight: l.Size.Height - childrenHeight,
		}
		childSize, err := child.Layout(ctx, childConstraints)
		if err != nil {
			return goui.Size{}, err
		}
		if err := layoututil.CheckOverflow(child.Element().Widget(), childSize, childConstraints); err != nil {
			return goui.Size{}, err
		}
		l.childOffsets[i] = goui.Point{X: 0, Y: childrenHeight}
		childrenHeight += childSize.Height
	}
	return l.Size, nil
}

func (l *columnLayouter) PositionAt(x, y int) (err error) {
	l.Position = goui.Point{X: x, Y: y}
	for i := range l.NumChildren() {
		l.Child(i).PositionAt(x+l.childOffsets[i].X, y+l.childOffsets[i].Y)
	}
	return nil
}
