package sizedbox

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
)

// SizedBox is a widget that imposes fixed width and height constraints on its child widget.
type SizedBox struct {
	ID     goui.ID
	Widget goui.Widget
	Width  metrics.DP // Desired width.
	Height metrics.DP // Desired height.
}

func (s *SizedBox) WidgetID() goui.ID {
	return s.ID
}

func (s *SizedBox) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementHelper{
		ElementLayouter: &sizedBoxLayouter{},
	}, nil
}

func (s *SizedBox) NumChildren() int {
	return gg.If(s.Widget != nil, 1, 0)
}

func (s *SizedBox) Child(n int) goui.AbstractWidget {
	return s.Widget
}

func (*SizedBox) ExclusiveType(marker.TypeWidget)    { /*Nop*/ }
func (*SizedBox) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

type sizedBoxLayouter struct {
	goui.LayouterHelper
	lastConstraints *metrics.Constraints
	pos             metrics.Point
}

func (l *sizedBoxLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	l.lastConstraints = &constraints
	sizedBox := l.Element().Widget().(*SizedBox)
	size = constraints.Clamp(metrics.Size{Width: sizedBox.Width, Height: sizedBox.Height})
	for child := range l.Children() {
		childConstraints := metrics.Constraints{
			MinWidth:  size.Width,
			MinHeight: size.Height,
			MaxWidth:  size.Width,
			MaxHeight: size.Height,
		}
		var childSize metrics.Size
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		err = debug.CheckLayoutOverflow(ctx, child.Element().Widget(), childSize, childConstraints)
		return // Only one child
	}
	return
}

func (l *sizedBoxLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
	l.pos = pt
	for child := range l.Children() {
		return child.PositionAt(ctx, pt)
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
		return l.PositionAt(ctx, l.pos)
	}
}
