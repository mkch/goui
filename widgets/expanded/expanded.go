package expanded

import (
	"github.com/mkch/gg"
	"github.com/mkch/gg/slices2"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
)

// Expanded is a widget that expands to fill the available space in the parent container.
// If more than one Expanded widget is present in a parent, the available space is divided
// among them according to their Flex factor.
type Expanded struct {
	ID     goui.ID
	Widget goui.Widget
	Flex   int // The Flex factor to use for this Expanded widget.
}

func (p *Expanded) WidgetID() goui.ID {
	return p.ID
}

func (p *Expanded) CreateElement(ctx *goui.Context) (goui.Element, error) {
	return &goui.ElementBase{
		ElementLayouter: &expandedLayouter{},
	}, nil
}

func (p *Expanded) NumChildren() int {
	return gg.If(p.Widget != nil, 1, 0)
}

func (p *Expanded) Child(n int) goui.Widget {
	return p.Widget
}

func (p *Expanded) Exclusive(goui.Container) { /*Nop*/ }

type expandedLayouter struct {
	goui.LayouterBase
}

func (l *expandedLayouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	for child := range l.Children() {
		size, err = child.Layout(ctx, constraints)
		if err != nil {
			return
		}
		if err = debug.CheckLayoutOverflow(ctx, child.Element().Widget(), size, constraints); err != nil {
			return
		}
		return // Only one child
	}
	return
}

func (l *expandedLayouter) PositionAt(x, y int) (err error) {
	for child := range l.Children() {
		return child.PositionAt(x, y)
	}
	return
}

// Layout layouts the given Expanded widgets within the available space.
func Layout(ctx *goui.Context, availableSpace int, expandedLayouters []goui.Layouter, setConstraints func(c *goui.Constraints, crossAxis int)) (sizes []goui.Size, err error) {
	widgets := slices2.Map(expandedLayouters, func(l goui.Layouter) *Expanded {
		return l.Element().Widget().(*Expanded)
	})
	totalFlex := slices2.Reduce(widgets, func(acc int, cur *Expanded, i int) int {
		flex := max(0, cur.Flex)
		return acc + flex
	}, 0)

	// Fast path: If totalFlex is zero, layout each Expanded with zero constraints.
	if totalFlex == 0 {
		for _, l := range expandedLayouters {
			if _, err = l.Layout(nil, goui.Constraints{ /*zero*/ }); err != nil {
				return
			}
		}
		return
	}

	// Call Layout on each Expanded widget with calculated tight constraints.
	remainingSpace := availableSpace
	for i, l := range expandedLayouters {
		var constraints goui.Constraints
		var size int
		if i == len(widgets)-1 {
			// Give all remaining space to the last Expanded to avoid rounding errors.
			size = remainingSpace
		} else if remainingSpace > 0 {
			// Ensure that the total allocated space does not exceed availableSpace
			// due to rounding errors.
			if widgets[i].Flex > 0 {
				size = int(float32(max(0, widgets[i].Flex)) / float32(totalFlex) * float32(availableSpace))
				remainingSpace -= size
			}
		}
		setConstraints(&constraints, size)
		var layoutSize goui.Size
		layoutSize, err = l.Layout(ctx, constraints)
		if err != nil {
			return
		}
		if err = debug.CheckLayoutOverflow(ctx, l.Element().Widget(), layoutSize, constraints); err != nil {
			return
		}
		sizes = append(sizes, layoutSize)
	}
	return
}
