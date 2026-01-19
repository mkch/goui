package row

import (
	"testing"

	"github.com/mkch/goui"
	"github.com/mkch/goui/gouitest"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/widgets/axes"
)

type mockWidget struct {
	ID      goui.ID
	Element mockElement
}

func (w *mockWidget) WidgetID() goui.ID {
	return w.ID
}

func (w *mockWidget) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &w.Element, nil
}

func (*mockWidget) ExclusiveType(marker.TypeWidget) { /*Nop*/ }

type mockElement struct {
	goui.ElementHelper
}

type mockLayouter struct {
	goui.LayouterHelper
	IntrinsicSize metrics.Size
	Position      metrics.Point
}

func (l *mockLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	return constraints.Clamp(l.IntrinsicSize), nil
}

func (l *mockLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) error {
	l.Position = pt
	return nil
}

func Test_RowSize(t *testing.T) {
	gouitest.RunContext(func(ctx *goui.Context) {
		widget1 := &mockWidget{
			ID: goui.ValueID("widget1"),
			Element: mockElement{
				ElementHelper: goui.ElementHelper{
					ElementLayouter: &mockLayouter{
						IntrinsicSize: metrics.Size{Width: 100, Height: 50},
					},
				},
			},
		}

		widget2 := &mockWidget{
			ID: goui.ValueID("widget2"),
			Element: mockElement{
				ElementHelper: goui.ElementHelper{
					ElementLayouter: &mockLayouter{
						IntrinsicSize: metrics.Size{Width: 200, Height: 30},
					},
				},
			},
		}

		row := &Row{
			Widgets:      []goui.Widget{widget1, widget2},
			MainAxisSize: axes.Min,
		}
		_, layouter, err := gouitest.BuildElementTree(ctx, row, nil)
		if err != nil {
			t.Fatalf("BuildElementTree error: %v", err)
		}
		size, err := layouter.Layout(ctx, metrics.Constraints{
			MinWidth: 150, MinHeight: 40,
			MaxWidth: 300, MaxHeight: 200,
		})
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if size.Width != 300 || size.Height != 50 {
			t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
		}

		row = &Row{
			Widgets:      []goui.Widget{widget1, widget2},
			MainAxisSize: axes.Max,
		}
		_, layouter, err = gouitest.BuildElementTree(ctx, row, nil)
		if err != nil {
			t.Fatalf("BuildElementTree error: %v", err)
		}
		size, err = layouter.Layout(ctx, metrics.Constraints{
			MinWidth: 150, MinHeight: 40,
			MaxWidth: 300, MaxHeight: 200,
		})
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if size.Width != 300 || size.Height != 50 {
			t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{}})
}

func Test_RowAlign(t *testing.T) {
	gouitest.RunContext(func(ctx *goui.Context) {
		widget1 := &mockWidget{
			ID: goui.ValueID("widget1"),
			Element: mockElement{
				ElementHelper: goui.ElementHelper{
					ElementLayouter: &mockLayouter{
						IntrinsicSize: metrics.Size{Width: 100, Height: 50},
					},
				},
			},
		}

		widget2 := &mockWidget{
			ID: goui.ValueID("widget2"),
			Element: mockElement{
				ElementHelper: goui.ElementHelper{
					ElementLayouter: &mockLayouter{
						IntrinsicSize: metrics.Size{Width: 200, Height: 30},
					},
				},
			},
		}

		column := &Row{
			Widgets:      []goui.Widget{widget1, widget2},
			MainAxisSize: axes.Min,
		}
		_, layouter, err := gouitest.BuildElementTree(ctx, column, nil)
		if err != nil {
			t.Fatalf("BuildElementTree error: %v", err)
		}
		size, err := layouter.Layout(ctx, metrics.Constraints{
			MinWidth: 150, MinHeight: 40,
			MaxWidth: 300, MaxHeight: 200,
		})
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if size.Width != 300 || size.Height != 50 {
			t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
		}
		if err = layouter.PositionAt(ctx, metrics.Point{X: 0, Y: 0}); err != nil {
			t.Fatalf("PositionAt error: %v", err)
		}
		if y := widget1.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
			t.Fatalf("Unexpected widget1 Y position: got %v, want 0", y)
		}
		if y := widget2.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
			t.Fatalf("Unexpected widget2 Y position: got %v, want 0", y)
		}

		column = &Row{
			Widgets:            []goui.Widget{widget1, widget2},
			MainAxisSize:       axes.Max,
			CrossAxisAlignment: axes.Center,
		}
		_, layouter, err = gouitest.BuildElementTree(ctx, column, nil)
		if err != nil {
			t.Fatalf("BuildElementTree error: %v", err)
		}
		size, err = layouter.Layout(ctx, metrics.Constraints{
			MinWidth: 150, MinHeight: 40,
			MaxWidth: 300, MaxHeight: 200,
		})
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if size.Width != 300 || size.Height != 50 {
			t.Fatalf("Unexpected size: got %v, want Width=300 Height=50", size)
		}
		if err = layouter.PositionAt(ctx, metrics.Point{X: 0, Y: 0}); err != nil {
			t.Fatalf("PositionAt error: %v", err)
		}
		if y := widget1.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
			t.Fatalf("Unexpected widget1 Y position: got %v, want 0", y)
		}
		if y := widget2.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 10 {
			t.Fatalf("Unexpected widget2 Y position: got %v, want 10", y)
		}

		column.CrossAxisAlignment = axes.End
		if _, layouter, err = gouitest.BuildElementTree(ctx, column, nil); err != nil {
			t.Fatalf("BuildElementTree error: %v", err)
		}
		if _, err = layouter.Layout(ctx, metrics.Constraints{
			MinWidth: 150, MinHeight: 40,
			MaxWidth: 300, MaxHeight: 200,
		}); err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if err = layouter.PositionAt(ctx, metrics.Point{X: 0, Y: 0}); err != nil {
			t.Fatalf("PositionAt error: %v", err)
		}
		if y := widget1.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 0 {
			t.Fatalf("Unexpected widget1 Y position: got %v, want 0", y)
		}
		if y := widget2.Element.ElementLayouter.(*mockLayouter).Position.Y; y != 20 {
			t.Fatalf("Unexpected widget2 Y position: got %v, want 20", y)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{}})
}
