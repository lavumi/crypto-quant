package indicator

// MACD represents MACD indicator values
type MACD struct {
	MACD      float64 // MACD line
	Signal    float64 // Signal line
	Histogram float64 // MACD Histogram
}

// CalculateMACD calculates MACD (Moving Average Convergence Divergence)
// MACD = 12-period EMA - 26-period EMA
// Signal = 9-period EMA of MACD
// Histogram = MACD - Signal
func CalculateMACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) *MACD {
	if len(prices) < slowPeriod {
		return &MACD{}
	}

	// Calculate fast and slow EMA
	fastEMA := EMA(prices, fastPeriod)
	slowEMA := EMA(prices, slowPeriod)

	// Calculate MACD line
	macdLine := fastEMA - slowEMA

	// Calculate MACD values for signal line
	macdValues := make([]float64, 0)
	for i := slowPeriod; i <= len(prices); i++ {
		fast := EMA(prices[:i], fastPeriod)
		slow := EMA(prices[:i], slowPeriod)
		macdValues = append(macdValues, fast-slow)
	}

	// Calculate signal line (EMA of MACD)
	var signalLine float64
	if len(macdValues) >= signalPeriod {
		signalLine = EMA(macdValues, signalPeriod)
	} else {
		signalLine = macdLine
	}

	// Calculate histogram
	histogram := macdLine - signalLine

	return &MACD{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: histogram,
	}
}

// StandardMACD calculates MACD with standard parameters (12, 26, 9)
func StandardMACD(prices []float64) *MACD {
	return CalculateMACD(prices, 12, 26, 9)
}




