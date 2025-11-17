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
	lastConstraints *Constraints
	childOffset     Point
}

func (l *centerLayouter) Layout(ctx *Context, constraints Constraints) (Size, error) {
	l.lastConstraints = &constraints
	l.size = Size{constraints.MaxWidth, constraints.MaxHeight}
	if l.numChildren() == 0 {
		return l.size, nil
	}
	childConstraints := Constraints{
		MinWidth: 0, MinHeight: 0,
		MaxWidth: l.size.Width, MaxHeight: l.size.Height}
	childSize, err := l.child(0).Layout(ctx, childConstraints)
	if err != nil {
		return Size{}, err
	}
	if childSize.Width > childConstraints.MaxWidth || childSize.Height > childConstraints.MaxHeight {
		return Size{}, &OverflowParentError{
			Widget:      l.child(0).element().widget(),
			Size:        childSize,
			Constraints: childConstraints,
		}
	}
	l.childOffset.X = (l.size.Width - childSize.Width) / 2
	l.childOffset.Y = (l.size.Height - childSize.Height) / 2
	return l.size, nil
}

func (l *centerLayouter) PositionAt(x, y int) (err error) {
	l.position = Point{x, y}
	if l.numChildren() == 0 {
		return nil
	}
	return l.child(0).PositionAt(x+l.childOffset.X, y+l.childOffset.Y)
}
