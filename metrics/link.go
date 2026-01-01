package metrics

import (
	_ "unsafe" // for go:linkname
)

//go:linkname link_setDebug github.com/mkch/goui.metricsSetDebug
func link_setDebug(value bool) {
	debug = value
}
