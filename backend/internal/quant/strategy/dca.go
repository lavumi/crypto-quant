package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
)

// DCAStrategy implements Dollar Cost Averaging strategy
// Buys a fixed USDT amount at regular intervals regardless of price
type DCAStrategy struct {
	period     time.Duration // Purchase interval (e.g., 24h for daily)
	amountUSDT float64       // Amount to purchase in USDT each time

	// Internal state
	lastPurchaseTime time.Time
	isInitialized    bool
}

// NewDCAStrategy creates a new DCA strategy
// period: Time interval between purchases (e.g., 24*time.Hour for daily)
// amountUSDT: Fixed amount in USDT to purchase each interval
func NewDCAStrategy(period time.Duration, amountUSDT float64) *DCAStrategy {
	return &DCAStrategy{
		period:     period,
		amountUSDT: amountUSDT,
	}
}

// Name returns the strategy name
func (s *DCAStrategy) Name() string {
	// Convert period to human readable format
	hours := s.period.Hours()
	var periodStr string

	if hours < 1 {
		minutes := s.period.Minutes()
		periodStr = fmt.Sprintf("%.0fm", minutes)
	} else if hours < 24 {
		periodStr = fmt.Sprintf("%.0fh", hours)
	} else {
		days := hours / 24
		periodStr = fmt.Sprintf("%.0fd", days)
	}

	return fmt.Sprintf("DCA_%s_%.0fUSDT", periodStr, s.amountUSDT)
}

// Initialize initializes the strategy
func (s *DCAStrategy) Initialize(ctx context.Context) error {
	s.lastPurchaseTime = time.Time{}
	s.isInitialized = false
	return nil
}

// OnCandle processes a candle and generates trading signals
func (s *DCAStrategy) OnCandle(ctx context.Context, candle *domain.Candle) (*backtest.Signal, error) {
	// First candle - initialize and make first purchase
	if !s.isInitialized {
		s.isInitialized = true
		s.lastPurchaseTime = candle.OpenTime

		// Calculate quantity based on USDT amount and current price
		quantity := s.amountUSDT / candle.Close

		return &backtest.Signal{
			Action:   domain.OrderSideBuy,
			Quantity: quantity,
			Price:    0, // Market order
			Reason:   fmt.Sprintf("DCA Initial Purchase: %.2f USDT @ %.2f", s.amountUSDT, candle.Close),
		}, nil
	}

	// Check if it's time for next purchase
	timeSinceLastPurchase := candle.OpenTime.Sub(s.lastPurchaseTime)

	if timeSinceLastPurchase >= s.period {
		s.lastPurchaseTime = candle.OpenTime

		// Calculate quantity based on USDT amount and current price
		quantity := s.amountUSDT / candle.Close

		return &backtest.Signal{
			Action:   domain.OrderSideBuy,
			Quantity: quantity,
			Price:    0, // Market order
			Reason:   fmt.Sprintf("DCA Regular Purchase: %.2f USDT @ %.2f", s.amountUSDT, candle.Close),
		}, nil
	}

	// No signal
	return nil, nil
}
