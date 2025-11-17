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
	childOffsets []Point
}

func (l *columnLayouter) Layout(ctx *Context, constraints Constraints) (Size, error) {
	l.size.Width = constraints.MaxWidth
	l.size.Height = constraints.MaxHeight
	l.childOffsets = make([]Point, l.numChildren())
	var childrenHeight = 0
	for i, child := range l.children {
		childConstraints := Constraints{
			MinWidth:  constraints.MinWidth,
			MinHeight: constraints.MinHeight,
			MaxWidth:  l.size.Width,
			MaxHeight: l.size.Height - childrenHeight,
		}
		childSize, err := child.Layout(ctx, childConstraints)
		if err != nil {
			return Size{}, err
		}
		if err := checkOverflow(child.element().widget(), childSize, childConstraints); err != nil {
			return Size{}, err
		}
		l.childOffsets[i] = Point{X: 0, Y: childrenHeight}
		childrenHeight += childSize.Height
	}
	return l.size, nil
}

func (l *columnLayouter) PositionAt(x, y int) (err error) {
	l.position = Point{x, y}
	for i, child := range l.children {
		child.PositionAt(x+l.childOffsets[i].X, y+l.childOffsets[i].Y)
	}
	return nil
}
