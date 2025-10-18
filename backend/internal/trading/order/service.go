package order

import (
	"context"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/domain"
	"github.com/lavumi/crypto-quant/internal/trading/portfolio"
	"github.com/lavumi/crypto-quant/internal/trading/wallet"
)

// Service handles order operations
type Service struct {
	wallet    *wallet.Manager
	portfolio *portfolio.Manager
	exchange  domain.Exchange
}

// NewService creates a new order service
func NewService(wallet *wallet.Manager, portfolio *portfolio.Manager, exchange domain.Exchange) *Service {
	return &Service{
		wallet:    wallet,
		portfolio: portfolio,
		exchange:  exchange,
	}
}

// PlaceOrder places a new order (spot trading)
func (s *Service) PlaceOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	// Get quote asset (assume USDT for simplicity)
	quoteAsset := "USDT"
	baseAsset := order.Symbol[:len(order.Symbol)-len(quoteAsset)]

	// Validate and lock balance
	if err := s.validateAndLockOrder(ctx, order, baseAsset, quoteAsset); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Place order on exchange
	executedOrder, err := s.exchange.PlaceOrder(ctx, order)
	if err != nil {
		// Unlock balance on failure
		s.unlockOrderBalance(order, baseAsset, quoteAsset)
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	// Update wallet and portfolio if order is filled
	if executedOrder.Status == domain.OrderStatusFilled {
		if err := s.settleOrder(executedOrder, baseAsset, quoteAsset); err != nil {
			return nil, fmt.Errorf("failed to settle order: %w", err)
		}
	}

	return executedOrder, nil
}

// validateAndLockOrder validates order and locks required balance
func (s *Service) validateAndLockOrder(ctx context.Context, order *domain.Order, baseAsset, quoteAsset string) error {
	if order.Side == domain.OrderSideBuy {
		// Buying: need quote asset (USDT)
		price := order.Price
		if order.Type == domain.OrderTypeMarket {
			currentPrice, err := s.exchange.GetCurrentPrice(ctx, order.Symbol)
			if err != nil {
				return fmt.Errorf("failed to get current price: %w", err)
			}
			price = currentPrice
			order.Price = price
		}

		required := price * order.Quantity * 1.001 // Add 0.1% for fees

		if !s.wallet.HasSufficientBalance(quoteAsset, required) {
			balance, _ := s.wallet.GetBalance(quoteAsset)
			free := 0.0
			if balance != nil {
				free = balance.Free
			}
			return fmt.Errorf("insufficient %s: have %.2f, need %.2f", quoteAsset, free, required)
		}

		return s.wallet.Lock(quoteAsset, required)
	} else {
		// Selling: need base asset (BTC, ETH, etc.)
		if !s.wallet.HasSufficientBalance(baseAsset, order.Quantity) {
			balance, _ := s.wallet.GetBalance(baseAsset)
			free := 0.0
			if balance != nil {
				free = balance.Free
			}
			return fmt.Errorf("insufficient %s: have %.8f, need %.8f", baseAsset, free, order.Quantity)
		}

		return s.wallet.Lock(baseAsset, order.Quantity)
	}
}

// unlockOrderBalance unlocks balance if order fails
func (s *Service) unlockOrderBalance(order *domain.Order, baseAsset, quoteAsset string) {
	if order.Side == domain.OrderSideBuy {
		amount := order.Price * order.Quantity * 1.001
		s.wallet.Unlock(quoteAsset, amount)
	} else {
		s.wallet.Unlock(baseAsset, order.Quantity)
	}
}

// settleOrder settles a filled order by updating wallet and portfolio
func (s *Service) settleOrder(order *domain.Order, baseAsset, quoteAsset string) error {
	fee := order.AvgPrice * order.FilledQty * 0.001 // 0.1% fee
	quoteAmount := order.AvgPrice * order.FilledQty

	if order.Side == domain.OrderSideBuy {
		// Transfer: locked USDT -> BTC
		lockedAmount := order.Price * order.Quantity * 1.001
		if err := s.wallet.Transfer(quoteAsset, lockedAmount, baseAsset, order.FilledQty); err != nil {
			return err
		}

		// Deduct fee from quote asset
		if err := s.wallet.Credit(quoteAsset, -(fee)); err != nil {
			return err
		}

		// Update position (for tracking)
		s.portfolio.UpdatePosition(order.Symbol, order.FilledQty, order.AvgPrice)
	} else {
		// Transfer: locked BTC -> USDT
		if err := s.wallet.Transfer(baseAsset, order.FilledQty, quoteAsset, quoteAmount-fee); err != nil {
			return err
		}

		// Update position (for tracking)
		s.portfolio.UpdatePosition(order.Symbol, -order.FilledQty, order.AvgPrice)
	}

	return nil
}

// CancelOrder cancels an existing order
func (s *Service) CancelOrder(ctx context.Context, orderID string) error {
	if err := s.exchange.CancelOrder(ctx, orderID); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	return nil
}

// GetOrder retrieves order details
func (s *Service) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	order, err := s.exchange.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}
