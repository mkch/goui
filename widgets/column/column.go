package column

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets/axes"
	"github.com/mkch/goui/widgets/internal/rowcol"
)

// Column is a [Container] [Widget] that arranges its children in a vertical column.
// The width of Column is the maximum width of its children.
// The height of Column is calculated based on its MainAxisSize property:
// - If MainAxisSize is Min, the height of Column is the sum of heights of its children.
// - If MainAxisSize is Max, the height of Column is the maximum height allowed by its parent.
type Column struct {
	ID                 goui.ID
	Widgets            []goui.Widget
	MainAxisSize       axes.Size
	CrossAxisAlignment axes.Alignment
}

func (c *Column) WidgetID() goui.ID {
	return c.ID
}

func (c *Column) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &goui.ElementBase{
		ElementLayouter: &rowcol.Layouter{
			Main:               func(s *goui.Size) *int { return &s.Height },
			Cross:              func(s *goui.Size) *int { return &s.Width },
			MaxMain:            func(c *goui.Constraints) *int { return &c.MaxHeight },
			MinMain:            func(c *goui.Constraints) *int { return &c.MinHeight },
			MaxCross:           func(c *goui.Constraints) *int { return &c.MaxWidth },
			MinCross:           func(c *goui.Constraints) *int { return &c.MinWidth },
			MainAxisSize:       func() axes.Size { return c.MainAxisSize },
			CrossAxisAlignment: func() axes.Alignment { return c.CrossAxisAlignment },
		},
	}, nil
}

func (c *Column) NumChildren() int {
	return len(c.Widgets)
}

func (c *Column) Child(n int) goui.Widget {
	return c.Widgets[n]
}

func (c *Column) Exclusive(goui.Container) { /*Nop*/ }
