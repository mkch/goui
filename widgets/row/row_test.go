package row

import (
	"testing"

	"github.com/mkch/goui"
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
	return constraints.Clamp(l.IntrinsicSize), nil
}

func (l *mockLayouter) PositionAt(x, y int) error {
	l.Position = goui.Point{X: x, Y: y}
	return nil
}

func Test_RowSize(t *testing.T) {
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

	column := &Row{
		Widgets:      []goui.Widget{widget1, widget2},
		MainAxisSize: axes.Min,
	}
	_, layouter, err := widgetstest.BuildElementTree(ctx, column, nil)
	if err != nil {
		t.Fatalf("BuildElementTree error: %v", err)
	}
	size, err := layouter.Layout(ctx, goui.Constraints{
		MinWidth: 150, MinHeight: 40,
		MaxWidth: 300, MaxHeight: 200,
	})
	if err != nil {
		t.Fatalf("Layout error: %v", err)
	}
	if size.Width != 300 || size.Height != 50 {
		t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
	}

	column = &Row{
		Widgets:      []goui.Widget{widget1, widget2},
		MainAxisSize: axes.Max,
	}
	_, layouter, err = widgetstest.BuildElementTree(ctx, column, nil)
	if err != nil {
		t.Fatalf("BuildElementTree error: %v", err)
	}
	size, err = layouter.Layout(ctx, goui.Constraints{
		MinWidth: 150, MinHeight: 40,
		MaxWidth: 300, MaxHeight: 200,
	})
	if err != nil {
		t.Fatalf("Layout error: %v", err)
	}
	if size.Width != 300 || size.Height != 50 {
		t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
	}
}

func Test_RowAlign(t *testing.T) {
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

	column := &Row{
		Widgets:      []goui.Widget{widget1, widget2},
		MainAxisSize: axes.Min,
	}
	_, layouter, err := widgetstest.BuildElementTree(ctx, column, nil)
	if err != nil {
		t.Fatalf("BuildElementTree error: %v", err)
	}
	size, err := layouter.Layout(ctx, goui.Constraints{
		MinWidth: 150, MinHeight: 40,
		MaxWidth: 300, MaxHeight: 200,
	})
	if err != nil {
		t.Fatalf("Layout error: %v", err)
	}
	if size.Width != 300 || size.Height != 50 {
		t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
	}
	if err = layouter.PositionAt(0, 0); err != nil {
		t.Fatalf("PositionAt error: %v", err)
	}
	if y := widget1.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
		t.Fatalf("Unexpected widget1 Y position: got %d, want 0", y)
	}
	if y := widget2.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
		t.Fatalf("Unexpected widget2 Y position: got %d, want 0", y)
	}

	column = &Row{
		Widgets:            []goui.Widget{widget1, widget2},
		MainAxisSize:       axes.Max,
		CrossAxisAlignment: axes.Center,
	}
	_, layouter, err = widgetstest.BuildElementTree(ctx, column, nil)
	if err != nil {
		t.Fatalf("BuildElementTree error: %v", err)
	}
	size, err = layouter.Layout(ctx, goui.Constraints{
		MinWidth: 150, MinHeight: 40,
		MaxWidth: 300, MaxHeight: 200,
	})
	if err != nil {
		t.Fatalf("Layout error: %v", err)
	}
	if size.Width != 300 || size.Height != 50 {
		t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
	}
	if err = layouter.PositionAt(0, 0); err != nil {
		t.Fatalf("PositionAt error: %v", err)
	}
	if y := widget1.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
		t.Fatalf("Unexpected widget1 Y position: got %d, want 0", y)
	}
	if y := widget2.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 10 {
		t.Fatalf("Unexpected widget2 Y position: got %d, want 10", y)
	}

	column.CrossAxisAlignment = axes.End
	if _, layouter, err = widgetstest.BuildElementTree(ctx, column, nil); err != nil {
		t.Fatalf("BuildElementTree error: %v", err)
	}
	if _, err = layouter.Layout(ctx, goui.Constraints{
		MinWidth: 150, MinHeight: 40,
		MaxWidth: 300, MaxHeight: 200,
	}); err != nil {
		t.Fatalf("Layout error: %v", err)
	}
	if err = layouter.PositionAt(0, 0); err != nil {
		t.Fatalf("PositionAt error: %v", err)
	}
	if y := widget1.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
		t.Fatalf("Unexpected widget1 Y position: got %d, want 0", y)
	}
	if y := widget2.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 20 {
		t.Fatalf("Unexpected widget2 Y position: got %d, want 20", y)
	}
}
