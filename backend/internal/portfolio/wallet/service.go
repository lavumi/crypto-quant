package wallet

import (
	"github.com/lavumi/crypto-quant/internal/domain"
)

// Service handles wallet operations
type Service struct {
	wallet *Manager
}

// NewService creates a new wallet service
func NewService(wallet *Manager) *Service {
	return &Service{
		wallet: wallet,
	}
}

// GetBalance returns balance for a specific asset
func (s *Service) GetBalance(asset string) (*domain.Balance, error) {
	return s.wallet.GetBalance(asset)
}

// GetAllBalances returns all balances
func (s *Service) GetAllBalances() map[string]*domain.Balance {
	return s.wallet.GetAllBalances()
}
