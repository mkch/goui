package expanded_test

import (
	"testing"

	"github.com/mkch/goui"
	"github.com/mkch/goui/gouitest"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/widgets/expanded"
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

func (*mockWidget) ExclusiveType(marker.TypeWidget) {}

type mockElement struct {
	goui.ElementHelper
}

type mockLayouter struct {
	goui.LayouterHelper
	IntrinsicSize metrics.Size
}

func (l *mockLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (metrics.Size, error) {
	return constraints.Clamp(l.IntrinsicSize), nil
}

func (l *mockLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) error {
	return nil
}

func newMockWidget(id string, size metrics.Size) *mockWidget {
	return &mockWidget{
		ID: goui.ValueID(id),
		Element: mockElement{
			ElementHelper: goui.ElementHelper{
				ElementLayouter: &mockLayouter{IntrinsicSize: size},
			},
		},
	}
}

// testBuildConstraints treats allocated as tight width, cross axis max=200.
func testBuildConstraints(allocated metrics.DP) metrics.Constraints {
	return metrics.Constraints{
		MinWidth: allocated, MaxWidth: allocated,
		MinHeight: 0, MaxHeight: 200,
	}
}

func buildExpandedLayouter(t *testing.T, ctx *goui.Context, exp *expanded.Expanded) goui.Layouter {
	t.Helper()
	_, layouter, err := gouitest.BuildElementTree(ctx, exp, nil)
	if err != nil {
		t.Fatalf("BuildElementTree error: %v", err)
	}
	return layouter
}

// Test_Layout_Single verifies the single-Expanded fast path:
// the sole Expanded occupies the entire available space.
func Test_Layout_Single(t *testing.T) {
	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

		exp := &expanded.Expanded{Widget: newMockWidget("child", metrics.Size{Width: 50, Height: 50})}
		l := buildExpandedLayouter(t, ctx, exp)

		sizes, err := expanded.Layout(ctx, 500, []goui.Layouter{l}, testBuildConstraints)
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if len(sizes) != 1 {
			t.Fatalf("expected 1 size, got %d", len(sizes))
		}
		if sizes[0].Width != 500 {
			t.Fatalf("expected width=500, got %v", sizes[0].Width)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{}})
}

// Test_Layout_FlexAllocation verifies that multiple Expanded widgets are allocated
// space proportional to their Flex factors.
func Test_Layout_FlexAllocation(t *testing.T) {
	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

		l1 := buildExpandedLayouter(t, ctx, &expanded.Expanded{
			ID:     goui.ValueID("exp1"),
			Widget: newMockWidget("child1", metrics.Size{Width: 10, Height: 10}),
			Flex:   1,
		})
		l2 := buildExpandedLayouter(t, ctx, &expanded.Expanded{
			ID:     goui.ValueID("exp2"),
			Widget: newMockWidget("child2", metrics.Size{Width: 10, Height: 10}),
			Flex:   2,
		})

		// availableSpace=300, Flex 1:2 -> expect [100, 200]
		sizes, err := expanded.Layout(ctx, 300, []goui.Layouter{l1, l2}, testBuildConstraints)
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if len(sizes) != 2 {
			t.Fatalf("expected 2 sizes, got %d", len(sizes))
		}
		if sizes[0].Width != 100 {
			t.Fatalf("expected sizes[0].Width=100, got %v", sizes[0].Width)
		}
		if sizes[1].Width != 200 {
			t.Fatalf("expected sizes[1].Width=200, got %v", sizes[1].Width)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{}})
}

// Test_Layout_RoundingError verifies that the last Expanded absorbs rounding errors
func Test_Layout_RoundingError(t *testing.T) {
	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

		l1 := buildExpandedLayouter(t, ctx, &expanded.Expanded{
			ID:     goui.ValueID("exp1"),
			Widget: newMockWidget("child1", metrics.Size{}),
			Flex:   1,
		})
		l2 := buildExpandedLayouter(t, ctx, &expanded.Expanded{
			ID:     goui.ValueID("exp2"),
			Widget: newMockWidget("child2", metrics.Size{}),
			Flex:   2,
		})
		l3 := buildExpandedLayouter(t, ctx, &expanded.Expanded{
			ID:     goui.ValueID("exp3"),
			Widget: newMockWidget("child3", metrics.Size{}),
			Flex:   3,
		})

		// availableSpace=100, Flex 1:2:3 -> exp1≈16.666, exp2≈33.333, exp3 gets remainder
		const availableSpace metrics.DP = 100
		sizes, err := expanded.Layout(ctx, availableSpace, []goui.Layouter{l1, l2, l3}, testBuildConstraints)
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if len(sizes) != 3 {
			t.Fatalf("expected 3 sizes, got %d", len(sizes))
		}
		total := sizes[0].Width + sizes[1].Width + sizes[2].Width
		if total != availableSpace {
			t.Fatalf("expected total width≈%v, got %v (sizes[0]=%v, sizes[1]=%v, sizes[2]=%v)",
				availableSpace, total, sizes[0].Width, sizes[1].Width, sizes[2].Width)
		}
	}, &goui.AppConfig{Debug: &goui.Debug{}})
}

// Test_Layout_ZeroFlex verifies that when all Flex factors are zero (or negative),
// each Expanded is laid out with zero constraints and no sizes are returned.
func Test_Layout_ZeroFlex(t *testing.T) {
	gouitest.RunContext(func(ctx *goui.Context) {
		defer ctx.App().Exit(0)

		l1 := buildExpandedLayouter(t, ctx, &expanded.Expanded{ID: goui.ValueID("exp1"), Flex: 0})
		l2 := buildExpandedLayouter(t, ctx, &expanded.Expanded{ID: goui.ValueID("exp2"), Flex: -1})

		sizes, err := expanded.Layout(ctx, 300, []goui.Layouter{l1, l2}, testBuildConstraints)
		if err != nil {
			t.Fatalf("Layout error: %v", err)
		}
		if len(sizes) != 0 {
			t.Fatalf("expected no sizes when totalFlex==0, got %d", len(sizes))
		}
	}, &goui.AppConfig{Debug: &goui.Debug{}})
}
