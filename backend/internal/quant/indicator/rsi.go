package indicator

import "math"

// RSI calculates Relative Strength Index
// RSI = 100 - (100 / (1 + RS))
// where RS = Average Gain / Average Loss
func RSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0 // Neutral RSI
	}

	// Calculate price changes
	gains := make([]float64, 0)
	losses := make([]float64, 0)

	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, math.Abs(change))
		}
	}

	if len(gains) < period {
		return 50.0
	}

	// Calculate initial average gain and loss (SMA)
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate smoothed averages (EMA-like)
	for i := period; i < len(gains); i++ {
		avgGain = ((avgGain * float64(period-1)) + gains[i]) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + losses[i]) / float64(period)
	}

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// StochasticRSI calculates Stochastic RSI
// Stoch RSI = (RSI - Lowest RSI) / (Highest RSI - Lowest RSI)
func StochasticRSI(prices []float64, rsiPeriod, stochPeriod int) float64 {
	if len(prices) < rsiPeriod+stochPeriod {
		return 50.0
	}

	// Calculate RSI values
	rsiValues := make([]float64, 0)
	for i := rsiPeriod; i <= len(prices); i++ {
		rsi := RSI(prices[:i], rsiPeriod)
		rsiValues = append(rsiValues, rsi)
	}

	if len(rsiValues) < stochPeriod {
		return 50.0
	}

	// Get last N RSI values
	recentRSI := rsiValues[len(rsiValues)-stochPeriod:]

	// Find highest and lowest RSI
	highestRSI := recentRSI[0]
	lowestRSI := recentRSI[0]
	for _, rsi := range recentRSI {
		if rsi > highestRSI {
			highestRSI = rsi
		}
		if rsi < lowestRSI {
			lowestRSI = rsi
		}
	}

	// Calculate Stochastic RSI
	currentRSI := rsiValues[len(rsiValues)-1]
	if highestRSI == lowestRSI {
		return 50.0
	}

	stochRSI := (currentRSI - lowestRSI) / (highestRSI - lowestRSI) * 100

	return stochRSI
}




