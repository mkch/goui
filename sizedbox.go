package goui

// SizedBox is a widget that imposes fixed width and height constraints on its child widget.
type SizedBox struct {
	ID     ID
	Widget Widget
	Width  int // Desired width.
	Height int // Desired height.
}

func (s *SizedBox) WidgetID() ID {
	return s.ID
}

func (s *SizedBox) CreateElement(ctx *Context) (Element, error) {
	return &sizedBoxElement{}, nil
}

func (s *SizedBox) NumChildren() int {
	return 1
}

func (s *SizedBox) Child(n int) Widget {
	return s.Widget
}

type sizedBoxElement struct {
	element
	layouter sizedBoxLayouter
}

func (e *sizedBoxElement) Layouter() Layouter {
	return &e.layouter
}

type sizedBoxLayouter struct {
	LayouterBase
}

func (l *sizedBoxLayouter) Layout(ctx *Context, constraints Constraints) (Size, error) {
	l.LayouterBase.Layout(ctx, constraints)

	sizedBox := l.element().(*sizedBoxElement).widget().(*SizedBox)
	l.size = Size{
		Width:  clampInt(sizedBox.Width, constraints.MinWidth, constraints.MaxWidth),
		Height: clampInt(sizedBox.Height, constraints.MinHeight, constraints.MaxHeight),
	}
	if l.numChildren() == 0 {
		return l.size, nil
	}
	childConstraints := Constraints{
		MinWidth:  l.size.Width,
		MinHeight: l.size.Height,
		MaxWidth:  l.size.Width,
		MaxHeight: l.size.Height,
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
	return l.size, nil
}

func (l *sizedBoxLayouter) PositionAt(x, y int) (err error) {
	l.LayouterBase.PositionAt(x, y)
	if l.numChildren() == 0 {
		return nil
	}
	return l.child(0).PositionAt(x, y)
}

func (l *sizedBoxLayouter) ChildrenIndependent() bool {
	return true
}
