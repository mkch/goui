// Package rowcol provides utilities to implement Row and Column widgets.
package rowcol

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/internal/debug"
	"github.com/mkch/goui/metrics"
	"github.com/mkch/goui/widgets/axes"
	"github.com/mkch/goui/widgets/expanded"
)

// Layouter is a layouter for Row and Column widgets.
type Layouter struct {
	goui.LayouterHelper
	childrenOffsets []metrics.Size

	// Main returns the main axis value (Width for [Row], Height for [Column]) of the given [Size].
	Main func(*metrics.Size) *metrics.DP
	// Cross returns the cross axis value (Height for [Row], Width for [Column]) of the given [Size].
	Cross func(*metrics.Size) *metrics.DP
	// MaxMain returns the maximum value of the main axis (Width for [Row], Height for [Column]) from the given [Constraints].
	MaxMain func(*metrics.Constraints) *metrics.DP
	// MinMain returns the minimum value of the main axis (Width for [Row], Height for [Column]) from the given [Constraints].
	MinMain func(*metrics.Constraints) *metrics.DP
	// MaxCross returns the maximum value of the cross axis (Height for [Row], Width for [Column]) from the given [Constraints].
	MaxCross func(*metrics.Constraints) *metrics.DP
	// MinCross returns the minimum value of the cross axis (Height for [Row], Width for [Column]) from the given [Constraints].
	MinCross func(*metrics.Constraints) *metrics.DP

	MainAxisSize       func() axes.Size
	CrossAxisAlignment func() axes.Alignment
}

func (l *Layouter) Layout(ctx *goui.Context, constraints metrics.Constraints) (size metrics.Size, err error) {
	l.childrenOffsets = l.childrenOffsets[:0]
	var notExpandableChildrenMain metrics.DP = 0
	*l.Cross(&size) = *l.MinCross(&constraints)
	var childrenSizes []metrics.Size
	var expandedChildren []goui.Layouter
	var expandedChildrenIndexes []int
	for child := range l.Children() {
		if _, ok := child.Element().Widget().(*expanded.Expanded); ok {
			expandedChildren = append(expandedChildren, child)
			expandedChildrenIndexes = append(expandedChildrenIndexes, len(childrenSizes))
			// Placeholder size, will be calculated later.
			childrenSizes = append(childrenSizes, metrics.Size{})
			continue
		}
		var childConstraints metrics.Constraints
		*l.MaxCross(&childConstraints) = *l.MaxCross(&constraints)
		*l.MaxMain(&childConstraints) = gg.IfFunc(*l.MaxMain(&constraints) == metrics.Infinity,
			func() metrics.DP { return metrics.Infinity },
			func() metrics.DP { return *l.MaxMain(&constraints) - notExpandableChildrenMain })
		var childSize metrics.Size
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		childrenSizes = append(childrenSizes, childSize)
		if err = debug.CheckLayoutOverflow(ctx, child.Element().Widget(), childSize, childConstraints); err != nil {
			return
		}
		notExpandableChildrenMain += *l.Main(&childSize)
		// calculate cross axis size
		*l.Cross(&size) = max(*l.Cross(&size), *l.Cross(&childSize))
	}
	// layout expanded children
	if len(expandedChildren) > 0 {
		availableSpace := *l.MaxMain(&constraints) - notExpandableChildrenMain
		var sizes []metrics.Size
		sizes, err = expanded.Layout(ctx, availableSpace, expandedChildren, func(c *metrics.Constraints, mainAxis metrics.DP) {
			*l.MinMain(c) = mainAxis
			*l.MaxMain(c) = mainAxis
			*l.MinCross(c) = 0
			*l.MaxCross(c) = *l.MaxCross(&constraints)
		})
		if err != nil {
			return
		}
		for i, childSize := range sizes {
			childrenSizes[expandedChildrenIndexes[i]] = childSize
			// calculate cross axis size
			*l.Cross(&size) = max(*l.Cross(&size), *l.Cross(&childSize))
		}
	}
	*l.Cross(&size) = max(*l.Cross(&size), *l.MinCross(&constraints))
	// calculate children offsets
	var childMain metrics.DP = 0
	for _, childSize := range childrenSizes {
		var offset metrics.Size
		*l.Main(&offset) = childMain
		l.childrenOffsets = append(l.childrenOffsets, offset)
		childMain += *l.Main(&childSize)
	}
	// determine main axis size
	switch l.MainAxisSize() {
	case axes.Min:
		*l.Main(&size) = childMain
	case axes.Max:
		*l.Main(&size) = *l.MaxMain(&constraints)
	}
	// apply cross axis alignment
	switch l.CrossAxisAlignment() {
	case axes.Start:
		// do nothing
	case axes.Center:
		for i := range l.childrenOffsets {
			*l.Cross(&l.childrenOffsets[i]) = (*l.Cross(&size) - *l.Cross(&childrenSizes[i])) / 2
		}
	case axes.End:
		for i := range l.childrenOffsets {
			*l.Cross(&l.childrenOffsets[i]) = *l.Cross(&size) - *l.Cross(&childrenSizes[i])
		}
	}
	// eliminate floating point error
	size = constraints.Clamp(size)
	return
}

func (l *Layouter) PositionAt(x, y metrics.DP) (err error) {
	var i = 0
	for child := range l.Children() {
		if err = child.PositionAt(x+l.childrenOffsets[i].Width, y+l.childrenOffsets[i].Height); err != nil {
			return
		}
		i++
	}
	return nil
}
