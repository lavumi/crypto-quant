package wallet

import (
	"fmt"
	"sync"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// Manager manages wallet balances
type Manager struct {
	mu       sync.RWMutex
	balances map[string]*domain.Balance
}

// NewManager creates a new wallet manager
func NewManager(initialBalances map[string]float64) *Manager {
	balances := make(map[string]*domain.Balance)
	for asset, amount := range initialBalances {
		balances[asset] = &domain.Balance{
			Asset:  asset,
			Free:   amount,
			Locked: 0,
			Total:  amount,
		}
	}

	return &Manager{
		balances: balances,
	}
}

// GetBalance returns balance for a specific asset
func (m *Manager) GetBalance(asset string) (*domain.Balance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	balance, ok := m.balances[asset]
	if !ok {
		return nil, fmt.Errorf("balance not found for asset: %s", asset)
	}

	return balance, nil
}

// GetAllBalances returns all balances
func (m *Manager) GetAllBalances() map[string]*domain.Balance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.balances
}

// Lock locks a specific amount of an asset
func (m *Manager) Lock(asset string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance, ok := m.balances[asset]
	if !ok {
		return fmt.Errorf("balance not found for asset: %s", asset)
	}

	if balance.Free < amount {
		return fmt.Errorf("insufficient free balance for %s: have %.8f, need %.8f", asset, balance.Free, amount)
	}

	balance.Free -= amount
	balance.Locked += amount

	return nil
}

// Unlock unlocks a specific amount of an asset
func (m *Manager) Unlock(asset string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance, ok := m.balances[asset]
	if !ok {
		return fmt.Errorf("balance not found for asset: %s", asset)
	}

	if balance.Locked < amount {
		return fmt.Errorf("insufficient locked balance for %s: have %.8f, need %.8f", asset, balance.Locked, amount)
	}

	balance.Locked -= amount
	balance.Free += amount

	return nil
}

// Deduct deducts an amount from locked balance
func (m *Manager) Deduct(asset string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance, ok := m.balances[asset]
	if !ok {
		return fmt.Errorf("balance not found for asset: %s", asset)
	}

	if balance.Locked < amount {
		return fmt.Errorf("insufficient locked balance for %s: have %.8f, need %.8f", asset, balance.Locked, amount)
	}

	balance.Locked -= amount
	balance.Total = balance.Free + balance.Locked

	return nil
}

// Credit adds an amount to free balance
func (m *Manager) Credit(asset string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	balance, ok := m.balances[asset]
	if !ok {
		// Create new balance if it doesn't exist
		m.balances[asset] = &domain.Balance{
			Asset:  asset,
			Free:   amount,
			Locked: 0,
			Total:  amount,
		}
		return nil
	}

	balance.Free += amount
	balance.Total = balance.Free + balance.Locked

	return nil
}

// Transfer transfers balance from one asset to another (for internal accounting)
func (m *Manager) Transfer(fromAsset string, fromAmount float64, toAsset string, toAmount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Deduct from source asset (from locked balance)
	fromBalance, ok := m.balances[fromAsset]
	if !ok {
		return fmt.Errorf("source balance not found for asset: %s", fromAsset)
	}

	if fromBalance.Locked < fromAmount {
		return fmt.Errorf("insufficient locked balance for %s: have %.8f, need %.8f", fromAsset, fromBalance.Locked, fromAmount)
	}

	fromBalance.Locked -= fromAmount
	fromBalance.Total = fromBalance.Free + fromBalance.Locked

	// Credit to destination asset
	toBalance, ok := m.balances[toAsset]
	if !ok {
		m.balances[toAsset] = &domain.Balance{
			Asset:  toAsset,
			Free:   toAmount,
			Locked: 0,
			Total:  toAmount,
		}
	} else {
		toBalance.Free += toAmount
		toBalance.Total = toBalance.Free + toBalance.Locked
	}

	return nil
}

// HasSufficientBalance checks if there's enough free balance
func (m *Manager) HasSufficientBalance(asset string, amount float64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	balance, ok := m.balances[asset]
	if !ok {
		return false
	}

	return balance.Free >= amount
}
