package goui

// Center is a [Container] [Widget] that centers its single child within itself.
type Center struct {
	ID     ID
	Widget Widget
}

func (c *Center) WidgetID() ID {
	return c.ID
}

func (c *Center) CreateElement(ctx *Context) (Element, error) {
	return &centerElement{}, nil
}

func (c *Center) NumChildren() int {
	return 1
}

func (c *Center) Child(n int) Widget {
	if n != 0 {
		panic("index out of range")
	}
	return c.Widget
}

type centerElement struct {
	element
}

func (e *centerElement) Layouter() Layouter {
	return &centerLayouter{}
}

type centerLayouter struct {
	LayouterBase
	size        Size
	childOffset Point
}

func (l *centerLayouter) Layout(ctx *Context, constraints Constraints) Size {
	l.size.Width = constraints.MaxWidth
	l.size.Height = constraints.MaxHeight
	childSize := l.child(0).Layout(ctx, Constraints{
		MinWidth: 0, MinHeight: 0,
		MaxWidth: l.size.Width, MaxHeight: l.size.Height})
	l.childOffset.X = (l.size.Width - childSize.Width) / 2
	l.childOffset.Y = (l.size.Height - childSize.Height) / 2
	return l.size
}

func (l *centerLayouter) Apply(x, y int) error {
	return l.child(0).Apply(x+l.childOffset.X, y+l.childOffset.Y)
}
