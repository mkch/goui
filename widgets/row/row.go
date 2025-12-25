package row

import (
	"github.com/mkch/goui"
	"github.com/mkch/goui/widgets/axes"
	"github.com/mkch/goui/widgets/internal/rowcol"
)

// Row is a [Container] [Widget] that arranges its children in a horizontal row.
// The height of Row is the maximum height of its children.
// The width of Row is calculated based on its MainAxisSize property:
// - If MainAxisSize is Min, the width of Row is the sum of widths of its children.
// - If MainAxisSize is Max, the width of Row is the maximum width allowed by its parent.
type Row struct {
	ID                 goui.ID
	Widgets            []goui.Widget
	MainAxisSize       axes.Size
	CrossAxisAlignment axes.Alignment
}

func (row *Row) WidgetID() goui.ID {
	return row.ID
}

func (row *Row) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &goui.ElementBase{
		ElementLayouter: &rowcol.Layouter{
			Main:               func(s *goui.Size) *int { return &s.Width },
			Cross:              func(s *goui.Size) *int { return &s.Height },
			MaxMain:            func(c *goui.Constraints) *int { return &c.MaxWidth },
			MinMain:            func(c *goui.Constraints) *int { return &c.MinWidth },
			MaxCross:           func(c *goui.Constraints) *int { return &c.MaxHeight },
			MinCross:           func(c *goui.Constraints) *int { return &c.MinHeight },
			MainAxisSize:       func() axes.Size { return row.MainAxisSize },
			CrossAxisAlignment: func() axes.Alignment { return row.CrossAxisAlignment },
		},
	}, nil
}

func (row *Row) NumChildren() int {
	return len(row.Widgets)
}

func (row *Row) Child(n int) goui.Widget {
	return row.Widgets[n]
}

func (row *Row) Exclusive(goui.Container) { /*Nop*/ }
