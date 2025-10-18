package price

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/datasource/exchange"
)

// Service handles market data operations
type Service struct {
	binance *exchange.BinanceExchange
}

// NewService creates a new market service
func NewService(binance *exchange.BinanceExchange) *Service {
	return &Service{
		binance: binance,
	}
}

// GetPrice returns the current price for a symbol
func (s *Service) GetPrice(ctx context.Context, symbol string) (float64, error) {
	price, err := s.binance.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}
	return price, nil
}

// GetMultiplePrices returns prices for multiple symbols
func (s *Service) GetMultiplePrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	prices := make(map[string]float64)

	for _, symbol := range symbols {
		price, err := s.binance.GetCurrentPrice(ctx, symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %w", symbol, err)
		}
		prices[symbol] = price
	}

	return prices, nil
}
