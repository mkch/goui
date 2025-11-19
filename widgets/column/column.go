package column

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
	"github.com/mkch/goui/widgets/axes"
)

// Column is a [Container] [Widget] that arranges its children in a vertical column.
// The width of Column is the maximum width of its children.
// The height of Column is calculated based on its MainAxisSize property:
// - If MainAxisSize is Min, the height of Column is the sum of heights of its children.
// - If MainAxisSize is Max, the height of Column is the maximum height allowed by its parent.
type Column struct {
	ID           goui.ID
	Widgets      []goui.Widget
	MainAxisSize axes.MainAxisSize
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

func (l *columnLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	l.childOffsets = l.childOffsets[:0]
	var childrenHeight = 0
	size.Width = constraints.MinWidth
	for i := range l.NumChildren() {
		child := l.Child(i)
		childConstraints := goui.Constraints{
			MinWidth:  0,
			MinHeight: 0,
			MaxWidth:  constraints.MaxWidth,
			MaxHeight: gg.If(constraints.MaxHeight == goui.Infinity, goui.Infinity, constraints.MaxHeight-childrenHeight),
		}
		var childSize goui.Size
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		if err := layoututil.CheckOverflow(child.Element().Widget(), childSize, childConstraints); err != nil {
			return goui.Size{}, err
		}
		l.childOffsets = append(l.childOffsets, goui.Point{X: 0, Y: childrenHeight})
		childrenHeight += childSize.Height
		size.Width = max(size.Width, childSize.Width)
	}
	switch l.Element().(*columnElement).Widget().(*Column).MainAxisSize {
	case axes.MainAxisSizeMin:
		size.Height = childrenHeight
	case axes.MainAxisSizeMax:
		size.Height = constraints.MaxHeight
	}
	return
}

func (l *columnLayouter) PositionAt(x, y int) (err error) {
	for i := range l.NumChildren() {
		l.Child(i).PositionAt(x+l.childOffsets[i].X, y+l.childOffsets[i].Y)
	}
	return nil
}
