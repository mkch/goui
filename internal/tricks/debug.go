// Package tricks includes internal tricks used by goui for debugging purposes.
// The content of this package is here to avoid import cycles.
package tricks

// Debug must have the same field layout of goui.Debug.
type Debug struct {
	LayoutOutline bool
}

func (debug *Debug) LayoutOutlineEnabled() bool {
	return debug != nil && debug.LayoutOutline
}

func (debug *Debug) Clone() (result *Debug) {
	if debug == nil {
		return nil
	}
	result = &Debug{}
	*result = *debug
	return
}
