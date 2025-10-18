package indicator

import "math"

// BollingerBands represents Bollinger Bands values
type BollingerBands struct {
	Upper  float64 // Upper band
	Middle float64 // Middle band (SMA)
	Lower  float64 // Lower band
	Width  float64 // Band width
	PctB   float64 // %B (position within bands)
}

// CalculateBollingerBands calculates Bollinger Bands
// Middle Band = SMA(price, period)
// Upper Band = Middle Band + (stdDev * multiplier)
// Lower Band = Middle Band - (stdDev * multiplier)
func CalculateBollingerBands(prices []float64, period int, multiplier float64) *BollingerBands {
	if len(prices) < period {
		return &BollingerBands{}
	}

	// Calculate middle band (SMA)
	middle := SMA(prices, period)

	// Calculate standard deviation
	start := len(prices) - period
	variance := 0.0
	for i := start; i < len(prices); i++ {
		diff := prices[i] - middle
		variance += diff * diff
	}
	variance /= float64(period)
	stdDev := math.Sqrt(variance)

	// Calculate upper and lower bands
	upper := middle + (stdDev * multiplier)
	lower := middle - (stdDev * multiplier)

	// Calculate band width
	width := (upper - lower) / middle * 100

	// Calculate %B (position within bands)
	currentPrice := prices[len(prices)-1]
	var pctB float64
	if upper != lower {
		pctB = (currentPrice - lower) / (upper - lower)
	}

	return &BollingerBands{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
		Width:  width,
		PctB:   pctB,
	}
}

// StandardBollingerBands calculates Bollinger Bands with standard parameters (20, 2)
func StandardBollingerBands(prices []float64) *BollingerBands {
	return CalculateBollingerBands(prices, 20, 2.0)
}




