package strategy

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/indicator"
	"github.com/lavumi/crypto-quant/internal/domain"
)

// RSIStrategy implements RSI (Relative Strength Index) strategy
// Buy when RSI < oversold threshold, Sell when RSI > overbought threshold
type RSIStrategy struct {
	period       int
	oversold     float64
	overbought   float64
	positionSize float64

	// Internal state
	prices     []float64
	inPosition bool
}

// NewRSIStrategy creates a new RSI strategy
// period: RSI calculation period (typically 14)
// oversold: Buy signal threshold (typically 30)
// overbought: Sell signal threshold (typically 70)
func NewRSIStrategy(period int, oversold, overbought, positionSize float64) *RSIStrategy {
	return &RSIStrategy{
		period:       period,
		oversold:     oversold,
		overbought:   overbought,
		positionSize: positionSize,
		prices:       make([]float64, 0),
		inPosition:   false,
	}
}

// Name returns the strategy name
func (s *RSIStrategy) Name() string {
	return fmt.Sprintf("RSI_%d_%.0f_%.0f", s.period, s.oversold, s.overbought)
}

// Initialize initializes the strategy
func (s *RSIStrategy) Initialize(ctx context.Context) error {
	s.prices = make([]float64, 0)
	s.inPosition = false
	return nil
}

// OnCandle processes a candle and generates trading signals
func (s *RSIStrategy) OnCandle(ctx context.Context, candle *domain.Candle) (*backtest.Signal, error) {
	// Add new price
	s.prices = append(s.prices, candle.Close)

	// Need at least period+1 candles to calculate RSI
	if len(s.prices) < s.period+1 {
		return nil, nil
	}

	// Calculate RSI
	rsi := indicator.RSI(s.prices, s.period)

	// Generate signals
	// Buy when oversold and not in position
	if rsi < s.oversold && !s.inPosition {
		s.inPosition = true
		return &backtest.Signal{
			Action:   domain.OrderSideBuy,
			Quantity: s.positionSize,
			Price:    0, // Market order
			Reason:   fmt.Sprintf("RSI Oversold: %.2f < %.2f", rsi, s.oversold),
		}, nil
	}

	// Sell when overbought and in position
	if rsi > s.overbought && s.inPosition {
		s.inPosition = false
		return &backtest.Signal{
			Action:   domain.OrderSideSell,
			Quantity: s.positionSize,
			Price:    0, // Market order
			Reason:   fmt.Sprintf("RSI Overbought: %.2f > %.2f", rsi, s.overbought),
		}, nil
	}

	return nil, nil
}




