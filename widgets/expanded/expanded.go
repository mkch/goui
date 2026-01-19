package expanded

import (
	"github.com/mkch/gg"
	"github.com/mkch/gg/slices2"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
	"github.com/mkch/goui/marker"
	"github.com/mkch/goui/metrics"
)

// Expanded is a widget that expands to fill the available space in the parent container.
// Expanded passes a tight constraint to its child widget to occupy all the allocated space.
// If more than one Expanded widget is present in a parent, the available space is divided
// among them according to their Flex factor.
type Expanded struct {
	ID     goui.ID
	Widget goui.Widget
	// The Flex factor to use for this Expanded widget.
	// Negative value is treated as zero.
	Flex float64
}

func (p *Expanded) WidgetID() goui.ID {
	return p.ID
}

func (p *Expanded) CreateElement(ctx *goui.Context, parent goui.Element) (goui.Element, error) {
	return &goui.ElementHelper{
		ElementLayouter: &expandedLayouter{},
	}, nil
}

func (p *Expanded) NumChildren() int {
	return gg.If(p.Widget != nil, 1, 0)
}

func (p *Expanded) Child(n int) goui.AbstractWidget {
	return p.Widget
}

func (*Expanded) ExclusiveType(marker.TypeWidget)    { /*Nop*/ }
func (*Expanded) ExclusiveKind(marker.KindContainer) { /*Nop*/ }

type expandedLayouter struct {
	goui.LayouterHelper
}

func (l *expandedLayouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	for child := range l.Children() {
		// Tight max constraint
		size, err = child.Layout(ctx, constraints.TightMax())
		if err != nil {
			return
		}
		err = debug.CheckLayoutOverflow(ctx, child.Element().Widget().(goui.Widget), size, constraints)
		return // Only one child
	}
	return constraints.MaxSize(), nil
}

func (l *expandedLayouter) PositionAt(ctx *goui.Context, pt metrics.Point) (err error) {
	for child := range l.Children() {
		return child.PositionAt(ctx, pt)
	}
	return
}

// Layout layouts the given Expanded widgets within the available space.
// setConstraints is a function set a constraints use provided cross axis size and other
// caller-calculated values.
func Layout(ctx *goui.Context, availableSpace metrics.DP, expandedLayouters []goui.Layouter, setConstraints func(c *metrics.Constraints, crossAxis metrics.DP)) (sizes []metrics.Size, err error) {
	if len(expandedLayouters) == 1 {
		// Only one Expanded, occupy all available space.
		l := expandedLayouters[0]
		var constraints metrics.Constraints
		setConstraints(&constraints, availableSpace)
		var size metrics.Size
		if size, err = l.Layout(ctx, constraints); err != nil {
			return
		}
		sizes = append(sizes, size)
		err = debug.CheckLayoutOverflow(ctx, l.Element().Widget().(goui.Widget), size, constraints)
		return
	}
	widgets := slices2.Map(expandedLayouters, func(l goui.Layouter) *Expanded {
		return l.Element().Widget().(*Expanded)
	})
	totalFlex := slices2.Reduce(widgets, func(acc float64, cur *Expanded, i int) float64 {
		flex := max(0, cur.Flex)
		return acc + flex
	}, 0)

	// Fast path: If totalFlex is zero, layout each Expanded with zero constraints.
	if totalFlex == 0 {
		for _, l := range expandedLayouters {
			if _, err = l.Layout(ctx, metrics.Constraints{ /*zero*/ }); err != nil {
				return
			}
		}
		return
	}

	totalFlexInv := 1 / totalFlex

	// Call Layout on each Expanded widget with calculated tight constraints.
	remainingSpace := availableSpace
	for i, l := range expandedLayouters {
		var constraints metrics.Constraints
		var size metrics.DP
		if i == len(widgets)-1 {
			// Give all remaining space to the last Expanded to avoid rounding errors.
			size = remainingSpace
		} else if remainingSpace > 0 {
			// Ensure that the total allocated space does not exceed availableSpace
			// due to rounding errors.
			if widgets[i].Flex > 0 {
				size = metrics.DP(max(0, widgets[i].Flex) * totalFlexInv * float64(availableSpace))
				remainingSpace -= size
			}
		}
		setConstraints(&constraints, size)
		var layoutSize metrics.Size
		layoutSize, err = l.Layout(ctx, constraints)
		if err != nil {
			return
		}
		if err = debug.CheckLayoutOverflow(ctx, l.Element().Widget().(goui.Widget), layoutSize, constraints); err != nil {
			return
		}
		sizes = append(sizes, layoutSize)
	}
	return
}
