package axes

// Size defines how much space a widget should take in an axis.
type Size int

const (
	// Max means the widget takes all available space in the the axis.
	Max Size = iota
	// Min means the widget takes the minimum space required in the axis.
	Min
)

type Alignment int

const (
	// Start means the widget is aligned to the start of the axis.
	Start Alignment = iota
	// Center means the widget is centered in the axis.
	Center
	// End means the widget is aligned to the end of the axis.
	End
)
