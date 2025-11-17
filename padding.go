package goui

type Padding struct {
	ID                       ID
	Widget                   Widget
	Left, Top, Right, Bottom int
}

func (p *Padding) WidgetID() ID {
	return p.ID
}

func (p *Padding) CreateElement(ctx *Context) (Element, error) {
	return &paddingElement{}, nil
}

func (p *Padding) NumChildren() int {
	return 1
}

func (p *Padding) Child(n int) Widget {
	return p.Widget
}

type paddingElement struct {
	element
	layouter paddingLayouter
}

func (e *paddingElement) Layouter() Layouter {
	return &e.layouter
}

type paddingLayouter struct {
	LayouterBase
}

func (l *paddingLayouter) Layout(ctx *Context, constraints Constraints) (Size, error) {
	padding := l.element().(*paddingElement).widget().(*Padding)
	if l.numChildren() == 0 {
		return Size{
			Width:  padding.Left + padding.Right,
			Height: padding.Top + padding.Bottom,
		}, nil
	}

	childMaxWidth := clampInt(constraints.MaxWidth-padding.Left-padding.Right, constraints.MinWidth, constraints.MaxWidth)
	childMaxHeight := clampInt(constraints.MaxHeight-padding.Top-padding.Bottom, constraints.MinHeight, constraints.MaxHeight)
	childConstraints := Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  childMaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: childMaxHeight,
	}
	childSize, err := l.child(0).Layout(ctx, childConstraints)
	if err != nil {
		return Size{}, err
	}
	if err = checkOverflow(l.child(0).element().widget(), childSize, childConstraints); err != nil {
		return Size{}, err
	}

	l.size = Size{
		Width:  clampInt(childSize.Width+padding.Left+padding.Right, constraints.MinWidth, constraints.MaxWidth),
		Height: clampInt(childSize.Height+padding.Top+padding.Bottom, constraints.MinHeight, constraints.MaxHeight),
	}
	return l.size, err
}

func (l *paddingLayouter) PositionAt(x, y int) (err error) {
	l.position = Point{x, y}
	if l.numChildren() == 0 {
		return
	}
	padding := l.element().(*paddingElement).widget().(*Padding)
	return l.child(0).PositionAt(l.position.X+padding.Left, l.position.Y+padding.Top)
}
