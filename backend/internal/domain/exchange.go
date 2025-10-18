package domain

import "context"

// Exchange defines the interface for interacting with exchanges
type Exchange interface {
	// PlaceOrder places a new order
	PlaceOrder(ctx context.Context, order *Order) (*Order, error)

	// CancelOrder cancels an existing order
	CancelOrder(ctx context.Context, orderID string) error

	// GetOrder retrieves order details
	GetOrder(ctx context.Context, orderID string) (*Order, error)

	// GetCurrentPrice gets the current price for a symbol
	GetCurrentPrice(ctx context.Context, symbol string) (float64, error)

	// GetCandles retrieves historical candlestick data
	GetCandles(ctx context.Context, symbol, interval string, limit int) ([]*Candle, error)

	// Close closes the exchange connection
	Close() error
}
