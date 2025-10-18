package strategy

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/indicator"
	"github.com/lavumi/crypto-quant/internal/domain"
)

// BBRSIStrategy combines Bollinger Bands and RSI
// Buy when price touches lower BB and RSI is oversold
// Sell when price touches upper BB and RSI is overbought
type BBRSIStrategy struct {
	bbPeriod      int
	bbMultiplier  float64
	rsiPeriod     int
	rsiOversold   float64
	rsiOverbought float64
	positionSize  float64

	// Internal state
	prices     []float64
	inPosition bool
}

// NewBBRSIStrategy creates a new Bollinger Bands + RSI strategy
func NewBBRSIStrategy(bbPeriod int, bbMultiplier float64, rsiPeriod int, rsiOversold, rsiOverbought, positionSize float64) *BBRSIStrategy {
	return &BBRSIStrategy{
		bbPeriod:      bbPeriod,
		bbMultiplier:  bbMultiplier,
		rsiPeriod:     rsiPeriod,
		rsiOversold:   rsiOversold,
		rsiOverbought: rsiOverbought,
		positionSize:  positionSize,
		prices:        make([]float64, 0),
		inPosition:    false,
	}
}

// Name returns the strategy name
func (s *BBRSIStrategy) Name() string {
	return fmt.Sprintf("BB_RSI_%d_%.1f_%d", s.bbPeriod, s.bbMultiplier, s.rsiPeriod)
}

// Initialize initializes the strategy
func (s *BBRSIStrategy) Initialize(ctx context.Context) error {
	s.prices = make([]float64, 0)
	s.inPosition = false
	return nil
}

// OnCandle processes a candle and generates trading signals
func (s *BBRSIStrategy) OnCandle(ctx context.Context, candle *domain.Candle) (*backtest.Signal, error) {
	// Add new price
	s.prices = append(s.prices, candle.Close)

	// Need sufficient data
	minPeriod := s.bbPeriod
	if s.rsiPeriod > minPeriod {
		minPeriod = s.rsiPeriod
	}

	if len(s.prices) < minPeriod+1 {
		return nil, nil
	}

	// Calculate indicators
	bb := indicator.CalculateBollingerBands(s.prices, s.bbPeriod, s.bbMultiplier)
	rsi := indicator.RSI(s.prices, s.rsiPeriod)
	currentPrice := candle.Close

	// Buy signal: price near/below lower BB and RSI oversold
	if currentPrice <= bb.Lower*1.01 && rsi < s.rsiOversold && !s.inPosition {
		s.inPosition = true
		return &backtest.Signal{
			Action:   domain.OrderSideBuy,
			Quantity: s.positionSize,
			Price:    0,
			Reason:   fmt.Sprintf("BB Lower (%.2f) + RSI Oversold (%.2f)", bb.Lower, rsi),
		}, nil
	}

	// Sell signal: price near/above upper BB and RSI overbought
	if currentPrice >= bb.Upper*0.99 && rsi > s.rsiOverbought && s.inPosition {
		s.inPosition = false
		return &backtest.Signal{
			Action:   domain.OrderSideSell,
			Quantity: s.positionSize,
			Price:    0,
			Reason:   fmt.Sprintf("BB Upper (%.2f) + RSI Overbought (%.2f)", bb.Upper, rsi),
		}, nil
	}

	return nil, nil
}




