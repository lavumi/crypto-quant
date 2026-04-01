package backtest

import (
	"fmt"
	"math"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// OrderIntent captures a trade before it is applied to portfolio state.
type OrderIntent struct {
	Timestamp time.Time
	Side      domain.OrderSide
	Price     float64
	Quantity  float64
	Reason    string
}

// PortfolioState is the current account snapshot used by execution models.
type PortfolioState struct {
	Balance  float64
	Position float64
	Equity   float64
}

// FeeModel calculates transaction costs for an order.
type FeeModel interface {
	CalculateFee(intent OrderIntent) float64
}

// SlippageModel adjusts the execution price for an order.
type SlippageModel interface {
	Apply(intent OrderIntent) float64
}

// RiskManager can reject an order before execution.
type RiskManager interface {
	Validate(intent OrderIntent, state PortfolioState) error
}

// FixedRateFeeModel applies the same fee rate to every trade.
type FixedRateFeeModel struct {
	Rate float64
}

// NewFixedRateFeeModel creates a fee model from a fixed rate.
func NewFixedRateFeeModel(rate float64) FixedRateFeeModel {
	return FixedRateFeeModel{Rate: rate}
}

// CalculateFee returns the fee charged for the order notional.
func (m FixedRateFeeModel) CalculateFee(intent OrderIntent) float64 {
	return intent.Price * intent.Quantity * m.Rate
}

// NoSlippageModel leaves the execution price unchanged.
type NoSlippageModel struct{}

// Apply returns the original order price.
func (NoSlippageModel) Apply(intent OrderIntent) float64 {
	return intent.Price
}

// BpsSlippageModel applies fixed adverse slippage in basis points.
type BpsSlippageModel struct {
	BasisPoints float64
}

// Apply adjusts the order price against the trader.
func (m BpsSlippageModel) Apply(intent OrderIntent) float64 {
	if intent.Price <= 0 || m.BasisPoints == 0 {
		return intent.Price
	}

	multiplier := m.BasisPoints / 10000.0
	switch intent.Side {
	case domain.OrderSideBuy:
		return intent.Price * (1 + multiplier)
	case domain.OrderSideSell:
		return intent.Price * math.Max(0, 1-multiplier)
	default:
		return intent.Price
	}
}

// NoopRiskManager accepts every order.
type NoopRiskManager struct{}

// Validate always accepts the order.
func (NoopRiskManager) Validate(intent OrderIntent, state PortfolioState) error {
	return nil
}

// MaxExposureRiskManager blocks buys that would exceed a target exposure.
type MaxExposureRiskManager struct {
	MaxExposure float64
}

// Validate rejects buy orders that would take exposure above MaxExposure.
func (m MaxExposureRiskManager) Validate(intent OrderIntent, state PortfolioState) error {
	if m.MaxExposure <= 0 || intent.Side != domain.OrderSideBuy {
		return nil
	}
	if state.Equity <= 0 {
		return fmt.Errorf("cannot validate exposure with non-positive equity")
	}

	nextExposure := (state.Position*intent.Price + intent.Quantity*intent.Price) / state.Equity
	if nextExposure > m.MaxExposure {
		return fmt.Errorf("order exceeds max exposure %.2f%%", m.MaxExposure*100)
	}
	return nil
}
