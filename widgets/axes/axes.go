package axes

// AxisSize defines how much space a widget should take in an axis.
type AxisSize int

const (
	// Max means the widget takes all available space in the the axis.
	Max AxisSize = iota
	// Min means the widget takes the minimum space required in the axis.
	Min
)

type CrossAxisSize int

const (
	// CrossAxisSizeMax means the widget takes all available space in the cross axis.
	CrossAxisSizeMax CrossAxisSize = iota
	// CrossAxisSizeMin means the widget takes the minimum space required in the cross axis.
	CrossAxisSizeMin
)
