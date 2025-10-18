package strategy

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/domain"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
)

// MACrossStrategy implements a Moving Average Crossover strategy
// Buy when fast MA crosses above slow MA, sell when it crosses below
type MACrossStrategy struct {
	fastPeriod int
	slowPeriod int

	// Internal state
	fastMA     []float64
	slowMA     []float64
	prices     []float64
	timestamps []string // Store timestamps for indicator data
	lastCross  string   // "golden" or "death"
}

// NewMACrossStrategy creates a new MA crossover strategy
func NewMACrossStrategy(fastPeriod, slowPeriod int) *MACrossStrategy {
	return &MACrossStrategy{
		fastPeriod: fastPeriod,
		slowPeriod: slowPeriod,
		prices:     make([]float64, 0),
		fastMA:     make([]float64, 0),
		slowMA:     make([]float64, 0),
	}
}

// Name returns the strategy name
func (s *MACrossStrategy) Name() string {
	return fmt.Sprintf("MA_Cross_%d_%d", s.fastPeriod, s.slowPeriod)
}

// Initialize initializes the strategy
func (s *MACrossStrategy) Initialize(ctx context.Context) error {
	s.prices = make([]float64, 0)
	s.fastMA = make([]float64, 0)
	s.slowMA = make([]float64, 0)
	s.timestamps = make([]string, 0)
	s.lastCross = ""
	return nil
}

// OnCandle processes a candle and generates trading signals
func (s *MACrossStrategy) OnCandle(ctx context.Context, candle *domain.Candle) (*backtest.Signal, error) {
	// Add new price
	s.prices = append(s.prices, candle.Close)

	// Store timestamp in RFC3339 format
	timestamp := candle.OpenTime.Format("2006-01-02T15:04:05Z07:00")
	s.timestamps = append(s.timestamps, timestamp)

	// Need at least slowPeriod candles to calculate MAs
	if len(s.prices) < s.slowPeriod {
		return nil, nil
	}

	// Calculate moving averages
	fastMA := s.calculateSMA(s.prices, s.fastPeriod)
	slowMA := s.calculateSMA(s.prices, s.slowPeriod)

	s.fastMA = append(s.fastMA, fastMA)
	s.slowMA = append(s.slowMA, slowMA)

	// Need at least 2 MA values to detect crossover
	if len(s.fastMA) < 2 {
		return nil, nil
	}

	// Get current and previous MA values
	prevFast := s.fastMA[len(s.fastMA)-2]
	currFast := s.fastMA[len(s.fastMA)-1]
	prevSlow := s.slowMA[len(s.slowMA)-2]
	currSlow := s.slowMA[len(s.slowMA)-1]

	// Detect crossover
	// Golden cross: fast MA crosses above slow MA (bullish signal)
	if prevFast <= prevSlow && currFast > currSlow && s.lastCross != "golden" {
		s.lastCross = "golden"
		return &backtest.Signal{
			Action:   domain.OrderSideBuy,
			Quantity: 0.01, // Fixed position size for simplicity
			Price:    0,    // Market order
			Reason:   fmt.Sprintf("Golden Cross: Fast MA (%.2f) > Slow MA (%.2f)", currFast, currSlow),
		}, nil
	}

	// Death cross: fast MA crosses below slow MA (bearish signal)
	if prevFast >= prevSlow && currFast < currSlow && s.lastCross != "death" {
		s.lastCross = "death"
		return &backtest.Signal{
			Action:   domain.OrderSideSell,
			Quantity: 0.01, // Fixed position size
			Price:    0,    // Market order
			Reason:   fmt.Sprintf("Death Cross: Fast MA (%.2f) < Slow MA (%.2f)", currFast, currSlow),
		}, nil
	}

	return nil, nil
}

// calculateSMA calculates Simple Moving Average
func (s *MACrossStrategy) calculateSMA(prices []float64, period int) float64 {
	// Use indicator package for consistency
	return calculateSMA(prices, period)
}

// calculateSMA is a helper function for SMA calculation
func calculateSMA(prices []float64, period int) float64 {
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

// GetIndicatorData returns indicator data for charting
func (s *MACrossStrategy) GetIndicatorData() (prices []float64, timestamps []string, fastMA []float64, slowMA []float64) {
	return s.prices, s.timestamps, s.fastMA, s.slowMA
}
