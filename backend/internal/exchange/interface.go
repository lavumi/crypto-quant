package exchange

import (
	"context"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// HistoricalDataProvider defines the interface for collecting historical data
type HistoricalDataProvider interface {
	// CollectHistorical collects historical candle data for a symbol
	CollectHistorical(ctx context.Context, symbol, interval string, startTime, endTime time.Time) error
}

// StreamingProvider defines the interface for real-time data streaming
type StreamingProvider interface {
	// StreamKlines streams real-time kline/candle data
	StreamKlines(ctx context.Context, symbol, interval string, callback func(*domain.Candle)) error

	// SubscribePrice subscribes to price updates via WebSocket
	SubscribePrice(ctx context.Context, symbol string) (<-chan float64, error)
}

// Exchange defines the complete exchange interface
// This extends the basic domain.Exchange with additional capabilities
type Exchange interface {
	domain.Exchange
	HistoricalDataProvider
	StreamingProvider
}
