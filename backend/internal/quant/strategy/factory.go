package strategy

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/lavumi/crypto-quant/internal/quant/backtest"
)

const (
	NameMACross     = "ma_cross"
	NameRSI         = "rsi"
	NameBBRSI       = "bb_rsi"
	NameDCA         = "dca"
	NameGoldenRSIBB = "golden_rsi_bb"
)

var supportedStrategies = []string{
	NameMACross,
	NameRSI,
	NameBBRSI,
	NameDCA,
	NameGoldenRSIBB,
}

// Config captures the normalized parameter set for strategy construction.
type Config struct {
	Name string `json:"name"`

	FastPeriod int `json:"fast_period,omitempty"`
	SlowPeriod int `json:"slow_period,omitempty"`

	RSIPeriod     int     `json:"rsi_period,omitempty"`
	RSIOversold   float64 `json:"rsi_oversold,omitempty"`
	RSIOverbought float64 `json:"rsi_overbought,omitempty"`

	BBPeriod     int     `json:"bb_period,omitempty"`
	BBMultiplier float64 `json:"bb_multiplier,omitempty"`

	DCAPeriod     string  `json:"dca_period,omitempty"`
	DCAAmountUSDT float64 `json:"dca_amount_usdt,omitempty"`

	GoldenFastPeriod      int     `json:"golden_fast_period,omitempty"`
	GoldenSlowPeriod      int     `json:"golden_slow_period,omitempty"`
	GoldenRSIPeriod       int     `json:"golden_rsi_period,omitempty"`
	GoldenRSILowerBound   float64 `json:"golden_rsi_lower_bound,omitempty"`
	GoldenRSIUpperBound   float64 `json:"golden_rsi_upper_bound,omitempty"`
	GoldenBBPeriod        int     `json:"golden_bb_period,omitempty"`
	GoldenBBMultiplier    float64 `json:"golden_bb_multiplier,omitempty"`
	GoldenVolumeThreshold float64 `json:"golden_volume_threshold,omitempty"`
	GoldenTakeProfitPct   float64 `json:"golden_take_profit_pct,omitempty"`
	GoldenStopLossPct     float64 `json:"golden_stop_loss_pct,omitempty"`

	PositionSize float64 `json:"position_size,omitempty"`
}

// Normalize applies defaults so callers can serialize a stable config.
func (c Config) Normalize() Config {
	normalized := c
	if normalized.Name == "" {
		normalized.Name = NameGoldenRSIBB
	}

	switch normalized.Name {
	case NameMACross:
		if normalized.FastPeriod == 0 {
			normalized.FastPeriod = 10
		}
		if normalized.SlowPeriod == 0 {
			normalized.SlowPeriod = 30
		}

	case NameRSI:
		if normalized.RSIPeriod == 0 {
			normalized.RSIPeriod = 14
		}
		if normalized.RSIOversold == 0 {
			normalized.RSIOversold = 30
		}
		if normalized.RSIOverbought == 0 {
			normalized.RSIOverbought = 70
		}
		if normalized.PositionSize == 0 {
			normalized.PositionSize = 0.01
		}

	case NameBBRSI:
		if normalized.BBPeriod == 0 {
			normalized.BBPeriod = 20
		}
		if normalized.BBMultiplier == 0 {
			normalized.BBMultiplier = 2.0
		}
		if normalized.RSIPeriod == 0 {
			normalized.RSIPeriod = 14
		}
		if normalized.RSIOversold == 0 {
			normalized.RSIOversold = 30
		}
		if normalized.RSIOverbought == 0 {
			normalized.RSIOverbought = 70
		}
		if normalized.PositionSize == 0 {
			normalized.PositionSize = 0.01
		}

	case NameDCA:
		if normalized.DCAPeriod == "" {
			normalized.DCAPeriod = "24h"
		}
		if normalized.DCAAmountUSDT == 0 {
			normalized.DCAAmountUSDT = 100
		}

	case NameGoldenRSIBB:
		if normalized.GoldenFastPeriod == 0 {
			normalized.GoldenFastPeriod = 5
		}
		if normalized.GoldenSlowPeriod == 0 {
			normalized.GoldenSlowPeriod = 20
		}
		if normalized.GoldenRSIPeriod == 0 {
			normalized.GoldenRSIPeriod = 14
		}
		if normalized.GoldenRSILowerBound == 0 {
			normalized.GoldenRSILowerBound = 40
		}
		if normalized.GoldenRSIUpperBound == 0 {
			normalized.GoldenRSIUpperBound = 70
		}
		if normalized.GoldenBBPeriod == 0 {
			normalized.GoldenBBPeriod = 20
		}
		if normalized.GoldenBBMultiplier == 0 {
			normalized.GoldenBBMultiplier = 2.0
		}
		if normalized.GoldenVolumeThreshold == 0 {
			normalized.GoldenVolumeThreshold = 1.3
		}
		if normalized.GoldenTakeProfitPct == 0 {
			normalized.GoldenTakeProfitPct = 0.06
		}
		if normalized.GoldenStopLossPct == 0 {
			normalized.GoldenStopLossPct = 0.03
		}
		if normalized.PositionSize == 0 {
			normalized.PositionSize = 1.0
		}
	}

	return normalized
}

// Build constructs a strategy and its normalized config JSON.
func Build(cfg Config) (backtest.Strategy, string, error) {
	normalized := cfg.Normalize()
	if !slices.Contains(supportedStrategies, normalized.Name) {
		return nil, "", fmt.Errorf("unsupported strategy %q", normalized.Name)
	}

	var strat backtest.Strategy

	switch normalized.Name {
	case NameMACross:
		strat = NewMACrossStrategy(normalized.FastPeriod, normalized.SlowPeriod)

	case NameRSI:
		strat = NewRSIStrategy(
			normalized.RSIPeriod,
			normalized.RSIOversold,
			normalized.RSIOverbought,
			normalized.PositionSize,
		)

	case NameBBRSI:
		strat = NewBBRSIStrategy(
			normalized.BBPeriod,
			normalized.BBMultiplier,
			normalized.RSIPeriod,
			normalized.RSIOversold,
			normalized.RSIOverbought,
			normalized.PositionSize,
		)

	case NameDCA:
		period, err := time.ParseDuration(normalized.DCAPeriod)
		if err != nil {
			return nil, "", fmt.Errorf("invalid dca_period %q: %w", normalized.DCAPeriod, err)
		}
		strat = NewDCAStrategy(period, normalized.DCAAmountUSDT)

	case NameGoldenRSIBB:
		strat = NewCustomGoldenRSIBBStrategy(
			normalized.GoldenFastPeriod,
			normalized.GoldenSlowPeriod,
			normalized.GoldenRSIPeriod,
			normalized.GoldenBBPeriod,
			normalized.GoldenRSILowerBound,
			normalized.GoldenRSIUpperBound,
			normalized.GoldenBBMultiplier,
			normalized.GoldenVolumeThreshold,
			normalized.GoldenTakeProfitPct,
			normalized.GoldenStopLossPct,
			normalized.PositionSize,
		)
	}

	configJSON, err := json.Marshal(normalized)
	if err != nil {
		return nil, "", fmt.Errorf("marshal strategy config: %w", err)
	}

	return strat, string(configJSON), nil
}

// SupportedStrategies returns the set of strategy names exposed to CLI and optimization code.
func SupportedStrategies() []string {
	return append([]string(nil), supportedStrategies...)
}
