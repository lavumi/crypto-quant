package domain

import "context"

// Strategy defines the interface for trading strategies
type Strategy interface {
	// Initialize is called before backtesting/trading starts
	Initialize(ctx context.Context) error

	// OnCandle is called for each new candle
	OnCandle(ctx context.Context, candle *Candle) (*Signal, error)

	// Name returns the strategy name
	Name() string
}

// Signal represents a trading signal
type Signal struct {
	Action   OrderSide // BUY or SELL
	Quantity float64   // Amount to trade
	Price    float64   // Limit price (0 for market order)
	Reason   string    // Reason for the signal
}
