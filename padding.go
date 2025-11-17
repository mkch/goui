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

	childMaxWidth := max(constraints.MaxWidth-padding.Left-padding.Right, 0)
	childMaxHeight := max(constraints.MaxHeight-padding.Top-padding.Bottom, 0)
	childConstraints := Constraints{
		MaxWidth:  childMaxWidth,
		MaxHeight: childMaxHeight,
	}
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
