package axes

// MainAxisSize defines how much space a widget should take in the main axis.
type MainAxisSize int

const (
	// MainAxisSizeMax means the widget takes all available space in the main axis.
	MainAxisSizeMax MainAxisSize = iota
	// MainAxisSizeMin means the widget takes the minimum space required in the main axis.
	MainAxisSizeMin
)

type CrossAxisSize int

const (
	// CrossAxisSizeMax means the widget takes all available space in the cross axis.
	CrossAxisSizeMax CrossAxisSize = iota
	// CrossAxisSizeMin means the widget takes the minimum space required in the cross axis.
	CrossAxisSizeMin
)
