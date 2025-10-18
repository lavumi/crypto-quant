package portfolio

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// Service handles portfolio operations
type Service struct {
	portfolio *Manager
	exchange  domain.Exchange
}

// NewService creates a new portfolio service
func NewService(portfolio *Manager, exchange domain.Exchange) *Service {
	return &Service{
		portfolio: portfolio,
		exchange:  exchange,
	}
}

// GetPosition returns position for a specific symbol
func (s *Service) GetPosition(symbol string) (*domain.Position, error) {
	return s.portfolio.GetPosition(symbol)
}

// GetAllPositions returns all positions
func (s *Service) GetAllPositions() map[string]*domain.Position {
	return s.portfolio.GetAllPositions()
}

// CalculatePnL calculates profit/loss for a symbol
func (s *Service) CalculatePnL(ctx context.Context, symbol string) (float64, float64, error) {
	position, err := s.portfolio.GetPosition(symbol)
	if err != nil {
		return 0, 0, err
	}

	currentPrice, err := s.exchange.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get current price: %w", err)
	}

	unrealizedPnL := (currentPrice - position.AvgEntryPrice) * position.Quantity
	realizedPnL := position.RealizedPnL

	return unrealizedPnL, realizedPnL, nil
}

// GetTotalPnL returns total PnL across all positions
func (s *Service) GetTotalPnL() (float64, float64) {
	return s.portfolio.CalculateTotalPnL()
}
