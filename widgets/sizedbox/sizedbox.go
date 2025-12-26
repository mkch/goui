package sizedbox

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
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
	return &goui.ElementBase{
		ElementLayouter: &sizedBoxLayouter{},
	}, nil
}

func (s *SizedBox) NumChildren() int {
	return gg.If(s.Widget != nil, 1, 0)
}

func (s *SizedBox) Child(n int) goui.Widget {
	return s.Widget
}

func (s *SizedBox) Exclusive(goui.Container) { /*Nop*/ }

type sizedBoxLayouter struct {
	goui.LayouterBase
	lastConstraints *goui.Constraints
	pos             goui.Point
}

func (l *sizedBoxLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	l.lastConstraints = &constraints
	sizedBox := l.Element().Widget().(*SizedBox)
	size = constraints.Clamp(goui.Size{Width: sizedBox.Width, Height: sizedBox.Height})
	for child := range l.Children() {
		childConstraints := goui.Constraints{
			MinWidth:  size.Width,
			MinHeight: size.Height,
			MaxWidth:  size.Width,
			MaxHeight: size.Height,
		}
		var childSize goui.Size
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		err = debug.CheckLayoutOverflow(ctx, child.Element().Widget(), childSize, childConstraints)
		return // Only one child
	}
	return
}

func (l *sizedBoxLayouter) PositionAt(x, y int) (err error) {
	l.pos = goui.Point{X: x, Y: y}
	for child := range l.Children() {
		return child.PositionAt(x, y)
	}
	return
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
		return l.PositionAt(l.pos.X, l.pos.Y)
	}
}
