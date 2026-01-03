package metrics

import (
	"math"
	"testing"
)

func TestFromAndTo(t *testing.T) {
	tests := []struct {
		name     string
		dpi      uint
		original float64
	}{
		{"96 DPI standard", 96, 10},
		{"160 DPI reference", 160, 10},
		{"192 DPI high density", 192, 10},
		{"zero value", 96, 0},
		{"fractional value", 96, 10.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// From converts device measurement to DP
			dp := From(tt.original, tt.dpi)
			// DP.To converts DP back to device measurement
			result := dp.To(tt.dpi)

			// Should be round-trip equivalent
			if math.Abs(result-tt.original) > 1e-10 {
				t.Errorf("From-To round trip failed: original=%v, result=%v, diff=%v",
					tt.original, result, result-tt.original)
			}
		})
	}
}

func TestToPxAndToDP(t *testing.T) {
	tests := []struct {
		name     string
		dpi      uint
		original int
	}{
		{"96 DPI standard", 96, 100},
		{"160 DPI reference", 160, 100},
		{"192 DPI high density", 192, 100},
		{"zero value", 96, 0},
		{"large value", 96, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToDP converts physical pixels to DP
			dp := Px(tt.original, tt.dpi)
			// DP.Px converts DP back to physical pixels
			result := dp.Px(tt.dpi)

			// Should be approximately equal (may differ due to rounding)
			if result != tt.original {
				// Allow 1 pixel difference due to rounding
				if abs(result-tt.original) > 1 {
					t.Errorf("ToDP-ToPx round trip failed: original=%d, result=%d, diff=%d",
						tt.original, result, result-tt.original)
				}
			}
		})
	}
}

func TestToWithDifferentDPI(t *testing.T) {
	// Test conversion between different DPI values
	// 10 DP should equal 16 pixels at 96 DPI (10 * 96 / 160)
	// and 19.2 pixels at 192 DPI (10 * 192 / 160)

	tests := []struct {
		dp       DP
		dpi      uint
		expected float64
	}{
		{10, 96, 6},     // 10 * 96 / 160 = 6
		{10, 160, 10},   // 10 * 160 / 160 = 10
		{10, 192, 12},   // 10 * 192 / 160 = 12
		{100, 96, 60},   // 100 * 96 / 160 = 60
		{160, 160, 160}, // 160 * 160 / 160 = 160
	}

	for _, tt := range tests {
		result := tt.dp.To(tt.dpi)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("To(%v, %d): expected %v, got %v",
				tt.dp, tt.dpi, tt.expected, result)
		}
	}
}

func TestFromWithDifferentDPI(t *testing.T) {
	// Test conversion from different DPI values to DP
	// 6 pixels at 96 DPI should equal 10 DP (6 * 160 / 96)
	// 12 pixels at 192 DPI should equal 10 DP (12 * 160 / 192)

	tests := []struct {
		p        float64
		dpi      uint
		expected float64
	}{
		{6, 96, 10},     // 6 * 160 / 96 = 10
		{10, 160, 10},   // 10 * 160 / 160 = 10
		{12, 192, 10},   // 12 * 160 / 192 = 10
		{60, 96, 100},   // 60 * 160 / 96 = 100
		{160, 160, 160}, // 160 * 160 / 160 = 160
	}

	for _, tt := range tests {
		result := From(tt.p, tt.dpi)
		if math.Abs(float64(result)-tt.expected) > 1e-10 {
			t.Errorf("From(%v, %d): expected %v, got %v",
				tt.p, tt.dpi, tt.expected, result)
		}
	}
}

func TestToPx(t *testing.T) {
	tests := []struct {
		name     string
		dp       DP
		dpi      uint
		expected int
	}{
		{"10 DP at 96 DPI", 10, 96, 6},    // 10 * 96 / 160 = 6
		{"10 DP at 160 DPI", 10, 160, 10}, // 10 * 160 / 160 = 10
		{"10 DP at 192 DPI", 10, 192, 12}, // 10 * 192 / 160 = 12
		{"0 DP", 0, 96, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dp.Px(tt.dpi)
			if result != tt.expected {
				t.Errorf("ToPx(%v, %d): expected %d, got %d",
					tt.dp, tt.dpi, tt.expected, result)
			}
		})
	}
}

func TestRoundWithDebugMode(t *testing.T) {
	// Test normal rounding
	tests := []struct {
		input    float64
		expected int
	}{
		{3.2, 3},
		{3.5, 4},
		{3.7, 4},
		{-3.5, -4},
		{0, 0},
	}

	for _, tt := range tests {
		result := round(tt.input)
		if result != tt.expected {
			t.Errorf("round(%v): expected %d, got %d",
				tt.input, tt.expected, result)
		}
	}
}

func TestRoundWithInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"NaN", math.NaN()},
		{"Positive Inf", math.Inf(1)},
		{"Negative Inf", math.Inf(-1)},
		{"Out of bounds positive", float64(math.MaxInt) * 2},
		{"Out of bounds negative", float64(math.MinInt) * 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round(tt.input)
			if result != -1 {
				t.Errorf("round(%v): expected -1, got %d", tt.input, result)
			}
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
