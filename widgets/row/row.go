package row

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
	"github.com/mkch/goui/widgets/axes"
)

// Row is a [Container] [Widget] that arranges its children in a horizontal row.
// The height of Row is the maximum height of its children.
// The width of Row is calculated based on its MainAxisSize property:
// - If MainAxisSize is Min, the width of Row is the sum of widths of its children.
// - If MainAxisSize is Max, the width of Row is the maximum width allowed by its parent.
type Row struct {
	ID                 goui.ID
	Widgets            []goui.Widget
	MainAxisSize       axes.Size
	CrossAxisAlignment axes.Alignment
}

func (c *Row) WidgetID() goui.ID {
	return c.ID
}

func (c *Row) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &goui.ElementBase{
		ElementLayouter: &rowLayouter{},
	}, nil
}

func (c *Row) NumChildren() int {
	return len(c.Widgets)
}

func (c *Row) Child(n int) goui.Widget {
	return c.Widgets[n]
}

func (c *Row) Exclusive(goui.Container) { /*Nop*/ }

type rowLayouter struct {
	goui.LayouterBase
	childOffsets []goui.Point
}

func (l *rowLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	l.childOffsets = l.childOffsets[:0]
	var childrenWidth = 0
	size.Height = constraints.MinHeight
	var childrenSizes []goui.Size
	for child := range l.Children() {
		childConstraints := goui.Constraints{
			MinWidth:  0,
			MinHeight: 0,
			MaxHeight: constraints.MaxHeight,
			MaxWidth: gg.IfFunc(constraints.MaxWidth == goui.Infinity,
				func() int { return goui.Infinity },
				func() int { return constraints.MaxWidth - childrenWidth }),
		}
		var childSize goui.Size
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		childrenSizes = append(childrenSizes, childSize)
		if err = layoututil.CheckOverflow(child.Element().Widget(), childSize, childConstraints); err != nil {
			return
		}
		l.childOffsets = append(l.childOffsets, goui.Point{X: childrenWidth, Y: 0})
		childrenWidth += childSize.Width
		// calculate cross axis size
		size.Height = max(size.Height, childSize.Height)
	}
	// determine main axis size
	row := l.Element().Widget().(*Row)
	switch row.MainAxisSize {
	case axes.Min:
		size.Width = childrenWidth
	case axes.Max:
		size.Width = constraints.MaxWidth
	}
	// apply cross axis alignment
	switch row.CrossAxisAlignment {
	case axes.Start:
		// do nothing
	case axes.Center:
		for i := range l.childOffsets {
			l.childOffsets[i].Y = (size.Height - childrenSizes[i].Height) / 2
		}
	case axes.End:
		for i := range l.childOffsets {
			l.childOffsets[i].Y = size.Height - childrenSizes[i].Height
		}
	}
	return
}

func (l *rowLayouter) PositionAt(x, y int) (err error) {
	var i = 0
	for child := range l.Children() {
		if err = child.PositionAt(x+l.childOffsets[i].X, y+l.childOffsets[i].Y); err != nil {
			return
		}
		i++
	}
	return nil
}
