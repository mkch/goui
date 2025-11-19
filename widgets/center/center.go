package center

import (
	"github.com/mkch/goui"
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
	return &centerElement{}, nil
}

func (c *Center) NumChildren() int {
	return 1
}

func (c *Center) Child(n int) goui.Widget {
	if n != 0 {
		panic("index out of range")
	}
	return c.Widget
}

type centerElement struct {
	goui.ElementBase
}

func (e *centerElement) SetWidget(widget goui.Widget) {
	center := widget.(*Center)
	if center.WidthFactor < 100 && center.WidthFactor != 0 {
		panic("Center.WidthFactor must be either 0 or greater than 100")
	}
	if center.HeightFactor < 100 && center.HeightFactor != 0 {
		panic("Center.HeightFactor must be either 0 or greater than 100")
	}
	e.ElementBase.SetWidget(widget)
}

func (e *centerElement) ElementLayouter() goui.Layouter {
	return &centerLayouter{}
}

type centerLayouter struct {
	goui.LayouterBase
	lastConstraints *goui.Constraints
	childOffset     goui.Point
}

func (l *centerLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (goui.Size, error) {
	l.lastConstraints = &constraints

	var childSize goui.Size
	var err error

	if l.NumChildren() > 0 {
		childSize, err = l.Child(0).Layout(ctx, constraints)
		if err != nil {
			return goui.Size{}, err
		}
	}

	if err := layoututil.CheckOverflow(l.Child(0).Element().Widget(), childSize, constraints); err != nil {
		return goui.Size{}, err
	}

	center := l.Element().(*centerElement).Widget().(*Center)
	if center.WidthFactor == 0 {
		l.Size.Width = constraints.MaxWidth
	} else {
		l.Size.Width = layoututil.Clamp(childSize.Width*center.WidthFactor/100, constraints.MinWidth, constraints.MaxWidth)
	}
	if center.HeightFactor == 0 {
		l.Size.Height = constraints.MaxHeight
	} else {
		l.Size.Height = layoututil.Clamp(childSize.Height*center.HeightFactor/100, constraints.MinHeight, constraints.MaxHeight)
	}

	l.childOffset.X = (l.Size.Width - childSize.Width) / 2
	l.childOffset.Y = (l.Size.Height - childSize.Height) / 2
	return l.Size, nil
}

func (l *centerLayouter) PositionAt(x, y int) (err error) {
	l.Position = goui.Point{X: x, Y: y}
	if l.NumChildren() == 0 {
		return nil
	}
	return l.Child(0).PositionAt(x+l.childOffset.X, y+l.childOffset.Y)
}

func (l *centerLayouter) Replayer() func(ctx *goui.Context) error {
	if l.lastConstraints == nil {
		return nil
	}
	center := l.Element().(*centerElement).Widget().(*Center)
	if center.WidthFactor == 0 || center.HeightFactor == 0 {
		return nil
	}
	return func(ctx *goui.Context) error {
		if _, err := l.Layout(ctx, *l.lastConstraints); err != nil {
			return err
		}
		return l.PositionAt(l.Position.X, l.Position.Y)
	}
}
