package domain

import "time"

// Position represents a trading position
type Position struct {
	Symbol        string    `json:"symbol"`
	Quantity      float64   `json:"quantity"` // Positive for long, negative for short
	AvgEntryPrice float64   `json:"avg_entry_price"`
	CurrentPrice  float64   `json:"current_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	RealizedPnL   float64   `json:"realized_pnl"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Balance represents account balance for an asset
type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`   // Available balance
	Locked float64 `json:"locked"` // Locked in orders
	Total  float64 `json:"total"`  // Free + Locked
}

// Portfolio represents the entire portfolio state
type Portfolio struct {
	Balances  map[string]*Balance  `json:"balances"`
	Positions map[string]*Position `json:"positions"`
	TotalPnL  float64              `json:"total_pnl"`
	UpdatedAt time.Time            `json:"updated_at"`
}
