package metrics

import (
	"math"
)

// DP is the measurement of device-independent pixel, based on 160 DPI reference.
type DP float64

// refDPI is reference DPI for DP unit.
const refDPI uint = 160

// refDPIInv is the inverse (reciprocal) of refDPI.
// Multiplying by refDPIInv is equivalent to dividing by refDPI.
const refDPIInv = 1.0 / float64(refDPI)

// ToPx converts dp to physical pixel measurement px.
// The dpi parameter is the pixel density (DPI, dots per inch) of the target device
// where px is measured.
func ToPx(dp DP, dpi uint) (px int) {
	return round(To(dp, dpi))
}

// ToDP converts physical pixel measurement px to dp.
// The dpi parameter is the pixel density (DPI, dots per inch) of the source device
// where px is measured.
func ToDP(px int, dpi uint) (dp DP) {
	return From(float64(px), dpi)
}

// From converts p measured in a device with given dpi (dots per inch, or pixel density) to DP.
func From(p float64, dpi uint) DP {
	return DP(p / float64(dpi) * float64(refDPI))
}

// To converts dp to a measurement in a device with given dpi (dots per inch, or pixel density).
func To(dp DP, dpi uint) float64 {
	return float64(dp) * refDPIInv * float64(dpi)
}

// debug indicates whether debug mode is enabled.
// Set by link_link_setDebug function which is linked to goui.metricsSetDebug.
var debug bool

// round converts a float64 to int.
// In debug mode, it panics if f is NaN, Inf or out of int bounds.
func round(f float64) int {
	if debug {
		if math.IsNaN(f) || math.IsInf(f, 0) {
			panic("attempt to convert NaN or Inf to int")
		}
	}

	rounded := math.Round(f)

	if debug {
		if rounded < float64(math.MinInt) || rounded > float64(math.MaxInt) {
			panic("attempt to convert out-of-bounds float to int")
		}
	}
	return int(rounded)
}
