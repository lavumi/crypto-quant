package portfolio

import (
	"fmt"
	"sync"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// Manager manages trading positions (for futures/margin trading)
type Manager struct {
	mu        sync.RWMutex
	positions map[string]*domain.Position
}

// NewManager creates a new portfolio manager
func NewManager() *Manager {
	return &Manager{
		positions: make(map[string]*domain.Position),
	}
}

// GetPosition returns position for a specific symbol
func (m *Manager) GetPosition(symbol string) (*domain.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pos, ok := m.positions[symbol]
	if !ok || pos.Quantity == 0 {
		return nil, fmt.Errorf("no position for symbol: %s", symbol)
	}

	return pos, nil
}

// GetAllPositions returns all positions
func (m *Manager) GetAllPositions() map[string]*domain.Position {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.positions
}

// UpdatePosition updates position after trade
func (m *Manager) UpdatePosition(symbol string, quantity, price float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pos, ok := m.positions[symbol]
	if !ok {
		pos = &domain.Position{
			Symbol:        symbol,
			Quantity:      0,
			AvgEntryPrice: 0,
			CurrentPrice:  price,
			UnrealizedPnL: 0,
			RealizedPnL:   0,
			UpdatedAt:     time.Now(),
		}
		m.positions[symbol] = pos
	}

	// Calculate new position
	oldQty := pos.Quantity
	newQty := oldQty + quantity

	if oldQty == 0 {
		// Opening new position
		pos.AvgEntryPrice = price
		pos.Quantity = newQty
	} else if (oldQty > 0 && quantity > 0) || (oldQty < 0 && quantity < 0) {
		// Adding to existing position
		totalCost := (pos.AvgEntryPrice * oldQty) + (price * quantity)
		pos.AvgEntryPrice = totalCost / newQty
		pos.Quantity = newQty
	} else {
		// Reducing or closing position
		if newQty == 0 {
			// Closing position - realize PnL
			realizedPnL := (price - pos.AvgEntryPrice) * (-quantity)
			pos.RealizedPnL += realizedPnL
			pos.Quantity = 0
			pos.UnrealizedPnL = 0
		} else if (oldQty > 0 && newQty > 0) || (oldQty < 0 && newQty < 0) {
			// Partial close
			closedQty := -quantity
			realizedPnL := (price - pos.AvgEntryPrice) * closedQty
			pos.RealizedPnL += realizedPnL
			pos.Quantity = newQty
		} else {
			// Reversing position
			closedQty := oldQty
			realizedPnL := (price - pos.AvgEntryPrice) * closedQty
			pos.RealizedPnL += realizedPnL
			pos.AvgEntryPrice = price
			pos.Quantity = newQty
		}
	}

	pos.CurrentPrice = price
	pos.UpdatedAt = time.Now()
}

// UpdatePrices updates current prices and unrealized PnL for all positions
func (m *Manager) UpdatePrices(prices map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for symbol, pos := range m.positions {
		if pos.Quantity == 0 {
			continue
		}

		if price, ok := prices[symbol]; ok {
			pos.CurrentPrice = price
			pos.UnrealizedPnL = (price - pos.AvgEntryPrice) * pos.Quantity
			pos.UpdatedAt = time.Now()
		}
	}
}

// CalculateTotalPnL calculates total profit/loss across all positions
func (m *Manager) CalculateTotalPnL() (unrealized, realized float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, pos := range m.positions {
		unrealized += pos.UnrealizedPnL
		realized += pos.RealizedPnL
	}

	return unrealized, realized
}
