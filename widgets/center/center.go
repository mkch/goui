package center

import (
	"errors"
	"slices"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
)

// Center is a [Container] [Widget] that centers its single child within itself.
type Center struct {
	ID     goui.ID
	Widget goui.Widget
	// Width scaling factor. If not 0, the desired with of Center is calculated as
	// child's width multiplied by WidthFactor.
	// A 0 WidthFactor means to take all available width from parent.
	// A non-zero WidthFactor must be greater than 100, or it panics.
	WidthFactor float64
	// Height scaling factor. If not 0, the desired height of Center is calculated as
	// child's height multiplied by HeightFactor.
	// A 0 HeightFactor means to take all available height from parent.
	// A non-zero HeightFactor must be greater than 100, or it panics.
	HeightFactor float64
}

func (c *Center) WidgetID() goui.ID {
	return c.ID
}

func (c *Center) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &centerElement{
		ElementBase: goui.ElementBase{
			ElementLayouter: &centerLayouter{},
		},
	}, nil
}

func (c *Center) NumChildren() int {
	return gg.If(c.Widget != nil, 1, 0)
}

func (c *Center) Child(n int) goui.WidgetBase {
	if c.Widget == nil || n != 0 {
		panic("index out of range")
	}
	return c.Widget
}

func (*Center) ExclusiveType(marker.TypeWidget)    { /*Nop*/ }
func (*Center) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

type centerElement struct {
	goui.ElementBase
}

func (e *centerElement) SetWidget(ctx *goui.Context, widget goui.WidgetBase) (err error) {
	center := widget.(*Center)
	if center.WidthFactor < 0 {
		return errors.New("Center.WidthFactor must be greater than or equal to 0")
	}
	if center.HeightFactor < 0 {
		return errors.New("Center.HeightFactor must be greater than or equal to 0")
	}
	return e.ElementBase.SetWidget(ctx, widget)
}

type centerLayouter struct {
	goui.LayouterHelper
	lastConstraints *metrics.Constraints // For replaying
	childOffset     metrics.Point
	pos             metrics.Point
}

func (l *centerLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	l.lastConstraints = &constraints

	center := l.Element().Widget().(*Center)
	for child := range l.Children() {
		var childSize metrics.Size
		childSize, err = child.Layout(ctx, constraints)
		if err != nil {
			return metrics.Size{}, err
		}

		if err = debug.CheckLayoutOverflow(ctx, child.Element().Widget().(goui.Widget), childSize, constraints); err != nil {
			return
		}
		if center.WidthFactor == 0 {
			size.Width = constraints.MaxWidth
		} else {
			size.Width = constraints.ClampWidth(childSize.Width * metrics.DP(center.WidthFactor))
		}
		if center.HeightFactor == 0 {
			size.Height = constraints.MaxHeight
		} else {
			size.Height = constraints.ClampHeight(childSize.Height * metrics.DP(center.HeightFactor))
		}

		l.childOffset.X = (size.Width - childSize.Width) / 2
		l.childOffset.Y = (size.Height - childSize.Height) / 2
		return
	}
	// No children
	size.Width = gg.If(center.WidthFactor == 0, constraints.MaxWidth, constraints.MinHeight)
	size.Height = gg.If(center.HeightFactor == 0, constraints.MaxHeight, constraints.MinHeight)
	return
}

func (l *centerLayouter) PositionAt(pt metrics.Point) (err error) {
	l.pos = pt
	children := slices.Collect(l.Children())
	if children == nil {
		return nil
	}
	return children[0].PositionAt(metrics.Point{X: pt.X + l.childOffset.X, Y: pt.Y + l.childOffset.Y})
}

func (l *centerLayouter) Replayer() func(ctx *goui.Context) error {
	if l.lastConstraints == nil {
		// No previous layout info.
		return nil
	}
	center := l.Element().(*centerElement).Widget().(*Center)
	if center.WidthFactor != 0 || center.HeightFactor != 0 {
		// Cannot replay if size depends on child size.
		return nil
	}
	return func(ctx *goui.Context) error {
		if _, err := l.Layout(ctx, *l.lastConstraints); err != nil {
			return err
		}
		return l.PositionAt(l.pos)
	}
}
