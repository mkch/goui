package goui

// Column is a [Container] [Widget] that arranges its children in a vertical column.
type Column struct {
	ID      ID
	Widgets []Widget
}

func (c *Column) WidgetID() ID {
	return c.ID
}

func (c *Column) CreateElement(ctx *Context) (Element, error) {
	return &columnElement{}, nil
}

func (c *Column) NumChildren() int {
	return len(c.Widgets)
}

func (c *Column) Child(n int) Widget {
	return c.Widgets[n]
}

type columnElement struct {
	element
}

func (e *columnElement) Layouter() Layouter {
	return &columnLayouter{}
}

type columnLayouter struct {
	LayouterBase
	size         Size
	childOffsets []Point
}

func (l *columnLayouter) Layout(ctx *Context, constraints Constraints) Size {
	l.size.Width = constraints.MaxWidth
	l.size.Height = constraints.MaxHeight
	l.childOffsets = make([]Point, l.numChildren())
	var y = 20
	for i, child := range l.children {
		childSize := child.Layout(ctx, Constraints{
			MinWidth:  0,
			MinHeight: 0,
			MaxWidth:  l.size.Width,
			MaxHeight: l.size.Height,
		})
		l.childOffsets[i] = Point{X: (l.size.Width - childSize.Width) / 2, Y: y}
		y += childSize.Height + 20
	}
	return l.size
}

func (l *columnLayouter) Apply(x, y int) error {
	for i, child := range l.children {
		child.Apply(l.childOffsets[i].X, l.childOffsets[i].Y)
	}
	return nil
}
