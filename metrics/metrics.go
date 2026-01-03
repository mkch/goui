// Package metrics provides types and functions for device-independent pixel (DP) measurements.
package metrics

import (
	"math"
)

// DP is the measurement of device-independent pixel, based on 160 DPI reference.
type DP float64

// ReferenceDPI is reference DPI for DP unit.
const ReferenceDPI uint = 160

// referenceDPIInv is the inverse (reciprocal) of referenceDPI.
// Multiplying by referenceDPIInv is equivalent to dividing by referenceDPI.
const referenceDPIInv = 1.0 / float64(ReferenceDPI)
const (
	Millimeter = DP(1.0 / 25.4 * float64(ReferenceDPI)) // 1 millimeter in DP
	Inch       = DP(ReferenceDPI)                       // 1 inch in DP
)

// Px converts dp to physical pixel measurement px.
// The dpi parameter is the pixel density (DPI, dots per inch) of the target device
// where px is measured.
func (dp DP) Px(dpi uint) (px int) {
	return round(dp.To(dpi))
}

// To converts dp to a measurement in a device with given dpi (dots per inch, or pixel density).
func (dp DP) To(dpi uint) float64 {
	return float64(dp) * referenceDPIInv * float64(dpi)
}

// Px converts physical pixel measurement px to dp.
// The dpi parameter is the pixel density (DPI, dots per inch) of the source device
// where px is measured.
func Px(px int, dpi uint) (dp DP) {
	return From(float64(px), dpi)
}

// From converts p measured in a device with given dpi (dots per inch, or pixel density) to DP.
func From(p float64, dpi uint) DP {
	return DP(p / float64(dpi) * float64(ReferenceDPI))
}

// round converts a float64 to int.
// -1 is returned if f is NaN, Inf or out of int bounds.
func round(f float64) int {
	f = math.Round(f)

	if math.IsNaN(f) || math.IsInf(f, 0) || f < float64(math.MinInt) || f > float64(math.MaxInt) {
		return -1
	}
	return int(f)
}
