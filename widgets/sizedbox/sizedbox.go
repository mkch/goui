package sizedbox

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
)

// SizedBox is a widget that imposes fixed width and height constraints on its child widget.
type SizedBox struct {
	ID     goui.ID
	Widget goui.Widget
	Width  int // Desired width.
	Height int // Desired height.
}

func (s *SizedBox) WidgetID() goui.ID {
	return s.ID
}

func (s *SizedBox) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &sizedBoxElement{}, nil
}

func (s *SizedBox) NumChildren() int {
	return 1
}

func (s *SizedBox) Child(n int) goui.Widget {
	return s.Widget
}

type sizedBoxElement struct {
	goui.ElementBase
	layouter sizedBoxLayouter
}

func (e *sizedBoxElement) ElementLayouter() goui.Layouter {
	return &e.layouter
}

type sizedBoxLayouter struct {
	goui.LayouterBase
	lastConstraints *goui.Constraints
}

func (l *sizedBoxLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (goui.Size, error) {
	l.lastConstraints = &constraints
	sizedBox := l.Element().(*sizedBoxElement).Widget().(*SizedBox)
	l.Size = goui.Size{
		Width:  layoututil.Clamp(sizedBox.Width, constraints.MinWidth, constraints.MaxWidth),
		Height: layoututil.Clamp(sizedBox.Height, constraints.MinHeight, constraints.MaxHeight),
	}
	if l.NumChildren() == 0 {
		return l.Size, nil
	}
	childConstraints := goui.Constraints{
		MinWidth:  l.Size.Width,
		MinHeight: l.Size.Height,
		MaxWidth:  l.Size.Width,
		MaxHeight: l.Size.Height,
	}
	childSize, err := l.Child(0).Layout(ctx, childConstraints)
	if err != nil {
		return goui.Size{}, err
	}
	if err = layoututil.CheckOverflow(l.Child(0).Element().Widget(), childSize, childConstraints); err != nil {
		return goui.Size{}, err
	}
	return l.Size, nil
}

func (l *sizedBoxLayouter) PositionAt(x, y int) (err error) {
	l.Position = goui.Point{X: x, Y: y}
	if l.NumChildren() == 0 {
		return nil
	}
	return l.Child(0).PositionAt(x, y)
}

func (l *sizedBoxLayouter) Replayer() func(ctx *goui.Context) error {
	if l.lastConstraints == nil {
		return nil
	}
	return func(ctx *goui.Context) error {
		_, err := l.Layout(ctx, *l.lastConstraints)
		if err != nil {
			return err
		}
		return l.PositionAt(l.Position.X, l.Position.Y)
	}
}
