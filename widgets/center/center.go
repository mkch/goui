package center

import (
	"slices"

	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
	"github.com/mkch/goui/layoututil"
)

// Center is a [Container] [Widget] that centers its single child within itself.
type Center struct {
	ID     goui.ID
	Widget goui.Widget
	// Width scaling factor. If not 0, the desired with of Center is calculated as
	// child's width multiplied by WidthFactor%(i.e, 120 means 120%).
	// A 0 WidthFactor means to take all available width from parent.
	// A non-zero WidthFactor must be greater than 100, or it panics.
	WidthFactor int
	// Height scaling factor. If not 0, the desired height of Center is calculated as
	// child's height multiplied by HeightFactor%(i.e, 120 means 120%).
	// A 0 HeightFactor means to take all available height from parent.
	// A non-zero HeightFactor must be greater than 100, or it panics.
	HeightFactor int
}

func (c *Center) WidgetID() goui.ID {
	return c.ID
}

func (c *Center) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &centerElement{
		ElementBase: goui.ElementBase{
			ElementLayouter: &centerLayouter{},
		},
	}, nil
}

func (c *Center) NumChildren() int {
	return gg.If(c.Widget != nil, 1, 0)
}

func (c *Center) Child(n int) goui.Widget {
	if n != 0 {
		panic("index out of range")
	}
	return c.Widget
}

func (c *Center) Exclusive(goui.Container) { /*Nop*/ }

type centerElement struct {
	goui.ElementBase
}

func (e *centerElement) SetWidget(ctx *goui.Context, widget goui.Widget) {
	center := widget.(*Center)
	if center.WidthFactor < 100 && center.WidthFactor != 0 {
		panic("Center.WidthFactor must be either 0 or greater than 100")
	}
	if center.HeightFactor < 100 && center.HeightFactor != 0 {
		panic("Center.HeightFactor must be either 0 or greater than 100")
	}
	e.ElementBase.SetWidget(ctx, widget)
}

type centerLayouter struct {
	goui.LayouterBase
	lastConstraints *goui.Constraints // For replaying
	childOffset     goui.Point
	pos             goui.Point
}

func (l *centerLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	l.lastConstraints = &constraints

	for child := range l.Children() {
		var childSize goui.Size
		childSize, err = child.Layout(ctx, constraints)
		if err != nil {
			return goui.Size{}, err
		}

		if err = debug.CheckLayoutOverflow(ctx, child.Element().Widget(), childSize, constraints); err != nil {
			return
		}

		center := l.Element().(*centerElement).Widget().(*Center)
		if center.WidthFactor == 0 {
			size.Width = constraints.MaxWidth
		} else {
			size.Width = layoututil.Clamp(childSize.Width*center.WidthFactor/100, constraints.MinWidth, constraints.MaxWidth)
		}
		if center.HeightFactor == 0 {
			size.Height = constraints.MaxHeight
		} else {
			size.Height = layoututil.Clamp(childSize.Height*center.HeightFactor/100, constraints.MinHeight, constraints.MaxHeight)
		}

		l.childOffset.X = (size.Width - childSize.Width) / 2
		l.childOffset.Y = (size.Height - childSize.Height) / 2
		return
	}
	return goui.Size{Width: constraints.MinWidth, Height: constraints.MinHeight}, nil
}

func (l *centerLayouter) PositionAt(x, y int) (err error) {
	l.pos = goui.Point{X: x, Y: y}
	children := slices.Collect(l.Children())
	if children == nil {
		return nil
	}
	return children[0].PositionAt(x+l.childOffset.X, y+l.childOffset.Y)
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
		return l.PositionAt(l.pos.X, l.pos.Y)
	}
}
