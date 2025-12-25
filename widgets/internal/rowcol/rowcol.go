// Package rowcol provides utilities to implement Row and Column widgets.
package rowcol

import (
	"github.com/mkch/gg"
	"github.com/mkch/goui"
	"github.com/mkch/goui/layoututil"
	"github.com/mkch/goui/widgets/axes"
	"github.com/mkch/goui/widgets/expanded"
)

// Layouter is a layouter for Row and Column widgets.
type Layouter struct {
	goui.LayouterBase
	childrenOffsets []goui.Size

	// Main returns the main axis value (Width for [Row], Height for [Column]) of the given [Size].
	Main func(*goui.Size) *int
	// Cross returns the cross axis value (Height for [Row], Width for [Column]) of the given [Size].
	Cross func(*goui.Size) *int
	// MaxMain returns the maximum value of the main axis (Width for [Row], Height for [Column]) from the given [Constraints].
	MaxMain func(*goui.Constraints) *int
	// MinMain returns the minimum value of the main axis (Width for [Row], Height for [Column]) from the given [Constraints].
	MinMain func(*goui.Constraints) *int
	// MaxCross returns the maximum value of the cross axis (Height for [Row], Width for [Column]) from the given [Constraints].
	MaxCross func(*goui.Constraints) *int
	// MinCross returns the minimum value of the cross axis (Height for [Row], Width for [Column]) from the given [Constraints].
	MinCross func(*goui.Constraints) *int

	MainAxisSize       func() axes.Size
	CrossAxisAlignment func() axes.Alignment
}

func (l *Layouter) Layout(ctx *goui.Context, constraints goui.Constraints) (size goui.Size, err error) {
	l.childrenOffsets = l.childrenOffsets[:0]
	var notExpandableChildrenMain = 0
	*l.Cross(&size) = *l.MinCross(&constraints)
	var childrenSizes []goui.Size
	var expandedChildren []goui.Layouter
	var expandedChildrenIndexes []int
	for child := range l.Children() {
		if _, ok := child.Element().Widget().(*expanded.Expanded); ok {
			expandedChildren = append(expandedChildren, child)
			expandedChildrenIndexes = append(expandedChildrenIndexes, len(childrenSizes))
			// Placeholder size, will be calculated later.
			childrenSizes = append(childrenSizes, goui.Size{})
			continue
		}
		var childConstraints goui.Constraints
		*l.MaxCross(&childConstraints) = *l.MaxCross(&constraints)
		*l.MaxMain(&childConstraints) = gg.IfFunc(*l.MaxMain(&constraints) == goui.Infinity,
			func() int { return goui.Infinity },
			func() int { return *l.MaxMain(&constraints) - notExpandableChildrenMain })
		var childSize goui.Size
		childSize, err = child.Layout(ctx, childConstraints)
		if err != nil {
			return
		}
		childrenSizes = append(childrenSizes, childSize)
		if err = layoututil.CheckOverflow(child.Element().Widget(), childSize, childConstraints); err != nil {
			return
		}
		notExpandableChildrenMain += *l.Main(&childSize)
		// calculate cross axis size
		*l.Cross(&size) = max(*l.Cross(&size), *l.Cross(&childSize))
	}
	// layout expanded children
	if len(expandedChildren) > 0 {
		availableSpace := *l.MaxMain(&constraints) - notExpandableChildrenMain
		var sizes []goui.Size
		sizes, err = expanded.Layout(availableSpace, expandedChildren, func(c *goui.Constraints, mainAxis int) {
			*l.MinMain(c) = mainAxis
			*l.MaxMain(c) = mainAxis
			*l.MinCross(c) = 0
			*l.MaxCross(c) = *l.MaxCross(&constraints)
		})
		if err != nil {
			return
		}
		var expandedTotalSize goui.Size // The overall size of all expanded children
		for i, childSize := range sizes {
			*l.Main(&expandedTotalSize) += *l.Main(&childSize)                                     // sum main axis sizes
			*l.Cross(&expandedTotalSize) = max(*l.Cross(&expandedTotalSize), *l.Cross(&childSize)) // max cross axis size
			childrenSizes[expandedChildrenIndexes[i]] = childSize
			// calculate cross axis size
			*l.Cross(&size) = max(*l.Cross(&size), *l.Cross(&childSize))
		}
		var expandedTotalConstraints goui.Constraints
		*l.MinMain(&expandedTotalConstraints) = availableSpace
		*l.MaxMain(&expandedTotalConstraints) = availableSpace
		*l.MinCross(&expandedTotalConstraints) = 0
		*l.MaxCross(&expandedTotalConstraints) = *l.MaxCross(&constraints)
		if err = layoututil.CheckOverflow(nil, expandedTotalSize, expandedTotalConstraints); err != nil {
			return
		}
	}
	*l.Cross(&size) = max(*l.Cross(&size), *l.MinCross(&constraints))
	// calculate children offsets
	var childMain = 0
	for _, childSize := range childrenSizes {
		var offset goui.Size
		*l.Main(&offset) = childMain
		l.childrenOffsets = append(l.childrenOffsets, offset)
		childMain += *l.Main(&childSize)
	}
	// determine main axis size
	switch l.MainAxisSize() {
	case axes.Min:
		*l.Main(&size) = max(childMain, *l.MinMain(&constraints))
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
	return
}

func (l *Layouter) PositionAt(x, y int) (err error) {
	var i = 0
	for child := range l.Children() {
		if err = child.PositionAt(x+l.childrenOffsets[i].Width, y+l.childrenOffsets[i].Height); err != nil {
			return
		}
		i++
	}
	return nil
}
