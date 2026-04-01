package market

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// PriceService handles market price operations
type PriceService struct {
	exchange domain.Exchange
}

// NewPriceService creates a new price service
func NewPriceService(exchange domain.Exchange) *PriceService {
	return &PriceService{
		exchange: exchange,
	}
}

// GetPrice returns the current price for a symbol
func (s *PriceService) GetPrice(ctx context.Context, symbol string) (float64, error) {
	price, err := s.exchange.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}
	return price, nil
}

// GetMultiplePrices returns prices for multiple symbols
func (s *PriceService) GetMultiplePrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	prices := make(map[string]float64)

	for _, symbol := range symbols {
		price, err := s.exchange.GetCurrentPrice(ctx, symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %w", symbol, err)
		}
		prices[symbol] = price
	}

	return prices, nil
}
