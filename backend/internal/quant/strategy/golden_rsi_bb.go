package strategy

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/domain"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/indicator"
)

// GoldenRSIBBStrategy implements a complex strategy combining:
// - Golden Cross (MA5 > MA20)
// - RSI filtering (40-70 range)
// - Bollinger Bands middle line confirmation
// - Volume spike detection (1.3x average)
//
// Entry conditions (ALL must be true):
// 1. Golden Cross: MA5 > MA20
// 2. RSI between 40 and 70
// 3. Price > Bollinger Middle Band (SMA20)
// 4. Volume >= 1.3x average volume
//
// Exit conditions (ANY triggers exit):
// 1. Take profit: +6%
// 2. Stop loss: -3%
// 3. Death Cross: MA5 < MA20
type GoldenRSIBBStrategy struct {
	fastPeriod      int     // MA fast period (default: 5)
	slowPeriod      int     // MA slow period (default: 20)
	rsiPeriod       int     // RSI period (default: 14)
	rsiLowerBound   float64 // RSI lower bound (default: 40)
	rsiUpperBound   float64 // RSI upper bound (default: 70)
	bbPeriod        int     // Bollinger Bands period (default: 20)
	bbMultiplier    float64 // Bollinger Bands multiplier (default: 2.0)
	volumeThreshold float64 // Volume multiplier (default: 1.3)
	takeProfitPct   float64 // Take profit percentage (default: 0.06 = 6%)
	stopLossPct     float64 // Stop loss percentage (default: 0.03 = 3%)
	positionSize    float64 // Position size

	// Internal state
	prices       []float64
	volumes      []float64
	timestamps   []string
	fastMA       []float64
	slowMA       []float64
	rsiValues    []float64
	inPosition   bool
	entryPrice   float64
	lastCross    string // "golden" or "death" or ""
}

// NewGoldenRSIBBStrategy creates a new strategy with default parameters
func NewGoldenRSIBBStrategy(positionSize float64) *GoldenRSIBBStrategy {
	return &GoldenRSIBBStrategy{
		fastPeriod:      5,
		slowPeriod:      20,
		rsiPeriod:       14,
		rsiLowerBound:   40,
		rsiUpperBound:   70,
		bbPeriod:        20,
		bbMultiplier:    2.0,
		volumeThreshold: 1.3,
		takeProfitPct:   0.06, // 6%
		stopLossPct:     0.03, // 3%
		positionSize:    positionSize,
		prices:          make([]float64, 0),
		volumes:         make([]float64, 0),
		timestamps:      make([]string, 0),
		fastMA:          make([]float64, 0),
		slowMA:          make([]float64, 0),
		rsiValues:       make([]float64, 0),
		inPosition:      false,
		entryPrice:      0,
		lastCross:       "",
	}
}

// NewCustomGoldenRSIBBStrategy creates a new strategy with custom parameters
func NewCustomGoldenRSIBBStrategy(
	fastPeriod, slowPeriod, rsiPeriod, bbPeriod int,
	rsiLowerBound, rsiUpperBound, bbMultiplier, volumeThreshold, takeProfitPct, stopLossPct, positionSize float64,
) *GoldenRSIBBStrategy {
	return &GoldenRSIBBStrategy{
		fastPeriod:      fastPeriod,
		slowPeriod:      slowPeriod,
		rsiPeriod:       rsiPeriod,
		rsiLowerBound:   rsiLowerBound,
		rsiUpperBound:   rsiUpperBound,
		bbPeriod:        bbPeriod,
		bbMultiplier:    bbMultiplier,
		volumeThreshold: volumeThreshold,
		takeProfitPct:   takeProfitPct,
		stopLossPct:     stopLossPct,
		positionSize:    positionSize,
		prices:          make([]float64, 0),
		volumes:         make([]float64, 0),
		timestamps:      make([]string, 0),
		fastMA:          make([]float64, 0),
		slowMA:          make([]float64, 0),
		rsiValues:       make([]float64, 0),
		inPosition:      false,
		entryPrice:      0,
		lastCross:       "",
	}
}

// Name returns the strategy name
func (s *GoldenRSIBBStrategy) Name() string {
	return fmt.Sprintf("GoldenRSIBB_MA%d_%d_RSI%d_BB%d_Vol%.1fx",
		s.fastPeriod, s.slowPeriod, s.rsiPeriod, s.bbPeriod, s.volumeThreshold)
}

// Initialize initializes the strategy
func (s *GoldenRSIBBStrategy) Initialize(ctx context.Context) error {
	s.prices = make([]float64, 0)
	s.volumes = make([]float64, 0)
	s.timestamps = make([]string, 0)
	s.fastMA = make([]float64, 0)
	s.slowMA = make([]float64, 0)
	s.rsiValues = make([]float64, 0)
	s.inPosition = false
	s.entryPrice = 0
	s.lastCross = ""
	return nil
}

// OnCandle processes a candle and generates trading signals
func (s *GoldenRSIBBStrategy) OnCandle(ctx context.Context, candle *domain.Candle) (*backtest.Signal, error) {
	// Add new price and volume
	s.prices = append(s.prices, candle.Close)
	s.volumes = append(s.volumes, candle.Volume)

	// Store timestamp in RFC3339 format
	timestamp := candle.OpenTime.Format("2006-01-02T15:04:05Z07:00")
	s.timestamps = append(s.timestamps, timestamp)

	// Need at least slowPeriod candles to calculate all indicators
	if len(s.prices) < s.slowPeriod {
		return nil, nil
	}

	// Calculate all indicators
	fastMA := indicator.SMA(s.prices, s.fastPeriod)
	slowMA := indicator.SMA(s.prices, s.slowPeriod)
	bb := indicator.CalculateBollingerBands(s.prices, s.bbPeriod, s.bbMultiplier)
	
	// Store MA values for tracking
	s.fastMA = append(s.fastMA, fastMA)
	s.slowMA = append(s.slowMA, slowMA)

	// Calculate RSI (need rsiPeriod+1 candles)
	var rsi float64
	if len(s.prices) >= s.rsiPeriod+1 {
		rsi = indicator.RSI(s.prices, s.rsiPeriod)
		s.rsiValues = append(s.rsiValues, rsi)
	} else {
		return nil, nil
	}

	// Current price
	currentPrice := candle.Close

	// Calculate average volume (use last 20 periods)
	avgVolume := s.calculateAverageVolume(20)

	// Check if we're in a position
	if s.inPosition {
		// Calculate profit/loss percentage
		profitPct := (currentPrice - s.entryPrice) / s.entryPrice

		// Exit condition 1: Take profit (+6%)
		if profitPct >= s.takeProfitPct {
			s.inPosition = false
			return &backtest.Signal{
				Action:   domain.OrderSideSell,
				Quantity: s.positionSize,
				Price:    0, // Market order
				Reason: fmt.Sprintf("Take Profit: +%.2f%% (entry: %.2f, current: %.2f)",
					profitPct*100, s.entryPrice, currentPrice),
			}, nil
		}

		// Exit condition 2: Stop loss (-3%)
		if profitPct <= -s.stopLossPct {
			s.inPosition = false
			return &backtest.Signal{
				Action:   domain.OrderSideSell,
				Quantity: s.positionSize,
				Price:    0, // Market order
				Reason: fmt.Sprintf("Stop Loss: %.2f%% (entry: %.2f, current: %.2f)",
					profitPct*100, s.entryPrice, currentPrice),
			}, nil
		}

		// Exit condition 3: Death Cross (MA5 < MA20)
		if len(s.fastMA) >= 2 && len(s.slowMA) >= 2 {
			prevFast := s.fastMA[len(s.fastMA)-2]
			currFast := s.fastMA[len(s.fastMA)-1]
			prevSlow := s.slowMA[len(s.slowMA)-2]
			currSlow := s.slowMA[len(s.slowMA)-1]

			// Detect death cross
			if prevFast >= prevSlow && currFast < currSlow {
				s.inPosition = false
				s.lastCross = "death"
				return &backtest.Signal{
					Action:   domain.OrderSideSell,
					Quantity: s.positionSize,
					Price:    0, // Market order
					Reason: fmt.Sprintf("Death Cross Exit: MA%d(%.2f) < MA%d(%.2f), P/L: %.2f%%",
						s.fastPeriod, currFast, s.slowPeriod, currSlow, profitPct*100),
				}, nil
			}
		}
	} else {
		// Not in position - check entry conditions
		// Need at least 2 MA values to detect crossover
		if len(s.fastMA) < 2 || len(s.slowMA) < 2 {
			return nil, nil
		}

		prevFast := s.fastMA[len(s.fastMA)-2]
		currFast := s.fastMA[len(s.fastMA)-1]
		prevSlow := s.slowMA[len(s.slowMA)-2]
		currSlow := s.slowMA[len(s.slowMA)-1]

		// Detect golden cross
		isGoldenCross := prevFast <= prevSlow && currFast > currSlow
		if !isGoldenCross {
			// Even if not crossing, check if still in golden state
			isGoldenCross = currFast > currSlow && s.lastCross == "golden"
		}

		if isGoldenCross && s.lastCross != "golden" {
			s.lastCross = "golden"
		}

		// Entry condition 1: Golden Cross (MA5 > MA20)
		if currFast <= currSlow {
			return nil, nil
		}

		// Entry condition 2: RSI between 40 and 70
		if rsi < s.rsiLowerBound || rsi > s.rsiUpperBound {
			return nil, nil
		}

		// Entry condition 3: Price > Bollinger Middle Band (SMA20)
		if currentPrice <= bb.Middle {
			return nil, nil
		}

		// Entry condition 4: Volume >= 1.3x average
		if candle.Volume < avgVolume*s.volumeThreshold {
			return nil, nil
		}

		// All conditions met - enter position
		s.inPosition = true
		s.entryPrice = currentPrice

		return &backtest.Signal{
			Action:   domain.OrderSideBuy,
			Quantity: s.positionSize,
			Price:    0, // Market order
			Reason: fmt.Sprintf("Golden Entry: MA5(%.2f)>MA20(%.2f), RSI(%.1f), Price(%.2f)>BB.Mid(%.2f), Vol(%.0f)>Avg(%.0f)x%.1f",
				currFast, currSlow, rsi, currentPrice, bb.Middle, candle.Volume, avgVolume, s.volumeThreshold),
		}, nil
	}

	return nil, nil
}

// calculateAverageVolume calculates the average volume for the last N periods
func (s *GoldenRSIBBStrategy) calculateAverageVolume(period int) float64 {
	if len(s.volumes) < period {
		period = len(s.volumes)
	}

	if period == 0 {
		return 0
	}

	start := len(s.volumes) - period
	sum := 0.0
	for i := start; i < len(s.volumes); i++ {
		sum += s.volumes[i]
	}

	return sum / float64(period)
}

// GetIndicatorData returns indicator data for charting
func (s *GoldenRSIBBStrategy) GetIndicatorData() map[string]interface{} {
	return map[string]interface{}{
		"prices":     s.prices,
		"timestamps": s.timestamps,
		"fastMA":     s.fastMA,
		"slowMA":     s.slowMA,
		"rsiValues":  s.rsiValues,
		"volumes":    s.volumes,
	}
}

