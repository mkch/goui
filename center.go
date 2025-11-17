package goui

// Center is a [Container] [Widget] that centers its single child within itself.
type Center struct {
	ID     ID
	Widget Widget
	// Width scaling factor. If not 0, the desired with of Center is calculated as
	// child's width multiplied by WidthFactor.
	// A 0 WidthFactor means to take all available width from parent.
	WidthFactor int
	// Height scaling factor. If not 0, the desired height of Center is calculated as
	// child's height multiplied by HeightFactor.
	// A 0 HeightFactor means to take all available height from parent.
	HeightFactor int
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

	var childSize Size
	var err error

	if l.numChildren() > 0 {
		childSize, err = l.child(0).Layout(ctx, constraints)
		if err != nil {
			return Size{}, err
		}
	}

	if err := checkOverflow(l.child(0).element().widget(), childSize, constraints); err != nil {
		return Size{}, err
	}

	center := l.element().(*centerElement).widget().(*Center)
	if center.WidthFactor == 0 {
		l.size.Width = constraints.MaxWidth
	} else {
		l.size.Width = clampInt(childSize.Width*center.WidthFactor, constraints.MinWidth, constraints.MaxWidth)
	}
	if center.HeightFactor == 0 {
		l.size.Height = constraints.MaxHeight
	} else {
		l.size.Height = clampInt(childSize.Height*center.HeightFactor, constraints.MinHeight, constraints.MaxHeight)
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

func (l *centerLayouter) Replayer() func(ctx *Context) error {
	if l.lastConstraints == nil {
		return nil
	}
	center := l.element().(*centerElement).widget().(*Center)
	if center.WidthFactor == 0 || center.HeightFactor == 0 {
		return nil
	}
	return func(ctx *Context) error {
		if _, err := l.Layout(ctx, *l.lastConstraints); err != nil {
			return err
		}
		return l.PositionAt(l.position.X, l.position.Y)
	}
}
