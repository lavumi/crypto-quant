package indicator

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	start := len(prices) - period
	sum := 0.0
	for i := start; i < len(prices); i++ {
		sum += prices[i]
	}

	return sum / float64(period)
}

// EMA calculates Exponential Moving Average
func EMA(prices []float64, period int) float64 {
	if len(prices) == 0 {
		return 0
	}

	if len(prices) < period {
		// Not enough data, return SMA
		return SMA(prices, len(prices))
	}

	// Calculate multiplier
	multiplier := 2.0 / float64(period+1)

	// Start with SMA
	ema := SMA(prices[:period], period)

	// Calculate EMA for remaining prices
	for i := period; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// VWMA calculates Volume Weighted Moving Average
func VWMA(prices []float64, volumes []float64, period int) float64 {
	if len(prices) < period || len(volumes) < period {
		return 0
	}

	start := len(prices) - period
	volumeSum := 0.0
	priceVolumeSum := 0.0

	for i := start; i < len(prices); i++ {
		priceVolumeSum += prices[i] * volumes[i]
		volumeSum += volumes[i]
	}

	if volumeSum == 0 {
		return 0
	}

	return priceVolumeSum / volumeSum
}



