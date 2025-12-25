package column

import (
	"testing"

	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
	"github.com/mkch/goui/widgets/axes"
	"github.com/mkch/goui/widgets/widgetstest"
)

type mockWidget struct {
	ID      goui.ID
	Element mockElement
}

func (w *mockWidget) WidgetID() goui.ID {
	return w.ID
}

func (w *mockWidget) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &w.Element, nil
}

type mockElement struct {
	goui.ElementBase
}

type mockLayouter struct {
	goui.LayouterBase
	IntrinsicSize goui.Size
	Position      goui.Point
}

func (l *mockLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	return layoututil.ClampSize(l.IntrinsicSize, constraints), nil
}

func (l *mockLayouter) PositionAt(x, y int) error {
	l.Position = goui.Point{X: x, Y: y}
	return nil
}

func Test_Column(t *testing.T) {
	ctx := widgetstest.NewContext()
	widget1 := &mockWidget{
		ID: goui.ValueID("widget1"),
		Element: mockElement{
			ElementBase: goui.ElementBase{
				ElementLayouter: &mockLayouter{
					IntrinsicSize: goui.Size{Width: 100, Height: 50},
				},
			},
		},
	}

	widget2 := &mockWidget{
		ID: goui.ValueID("widget2"),
		Element: mockElement{
			ElementBase: goui.ElementBase{
				ElementLayouter: &mockLayouter{
					IntrinsicSize: goui.Size{Width: 200, Height: 30},
				},
			},
		},
	}

	column := &Column{
		Widgets:      []goui.Widget{widget1, widget2},
		MainAxisSize: axes.Min,
	}
	_, layouter, err := widgetstest.BuildElementTree(ctx, column, nil)
	if err != nil {
		t.Fatalf("CreateElement error: %v", err)
	}
	size, err := layouter.Layout(ctx, goui.Constraints{
		MinWidth: 150, MinHeight: 40,
		MaxWidth: 300, MaxHeight: 200,
	})
	if err != nil {
		t.Fatalf("Layout error: %v", err)
	}
	if size.Width != 200 || size.Height != 80 {
		t.Fatalf("Unexpected size: got %v, want Width=200 Height=80", size)
	}

	column = &Column{
		Widgets:      []goui.Widget{widget1, widget2},
		MainAxisSize: axes.Max,
	}
	_, layouter, err = widgetstest.BuildElementTree(ctx, column, nil)
	if err != nil {
		t.Fatalf("CreateElement error: %v", err)
	}
	size, err = layouter.Layout(ctx, goui.Constraints{
		MinWidth: 150, MinHeight: 40,
		MaxWidth: 300, MaxHeight: 200,
	})
	if err != nil {
		t.Fatalf("Layout error: %v", err)
	}
	if size.Width != 200 || size.Height != 200 {
		t.Fatalf("Unexpected size: got %v, want Width=200 Height=80", size)
	}
}
