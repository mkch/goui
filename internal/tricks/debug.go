// Package tricks includes internal tricks used by goui for debugging purposes.
// The content of this package is here to avoid import cycles.
package tricks

// Debug must have the same field layout of goui.Debug.
type Debug struct {
	Layout *bool
}

func (debug *Debug) LayoutDebugEnabled() bool {
	return debug != nil && debug.Layout != nil && *debug.Layout
}

func (debug *Debug) Clone() (result *Debug) {
	if debug == nil {
		return nil
	}
	result = &Debug{}
	if debug.Layout != nil {
		layoutDebug := *debug.Layout
		result.Layout = &layoutDebug
	}
	return
}
