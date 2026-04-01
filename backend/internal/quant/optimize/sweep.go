package optimize

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/domain"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
)

// SweepSpec describes a parameter sweep over a base strategy config.
type SweepSpec struct {
	BaseConfig strategy.Config
	Parameters map[string][]string
	SortBy     string
	Top        int
}

// RunnerConfig carries execution settings shared across all candidates.
type RunnerConfig struct {
	InitialBalance float64
	Commission     float64
	Persist        bool
	DB             *database.DB
	Symbol         string
	Interval       string
}

// SweepResult stores the result of one parameter combination.
type SweepResult struct {
	Config     strategy.Config
	ConfigJSON string
	Result     *backtest.Result
	Err        error
}

// SweepSummary provides a quick aggregate view of a sweep run.
type SweepSummary struct {
	TotalCandidates int
	SuccessfulRuns  int
	FailedRuns      int
	SortBy          string
	Best            *SweepResult
	Median          *SweepResult
	Worst           *SweepResult
}

// Expand produces all candidate configs from the cartesian product of Parameters.
func Expand(spec SweepSpec) ([]strategy.Config, error) {
	base := spec.BaseConfig.Normalize()
	if len(spec.Parameters) == 0 {
		return []strategy.Config{base}, nil
	}

	keys := make([]string, 0, len(spec.Parameters))
	for key := range spec.Parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	configs := []strategy.Config{base}
	for _, key := range keys {
		values := spec.Parameters[key]
		if len(values) == 0 {
			return nil, fmt.Errorf("parameter %q has no values", key)
		}

		next := make([]strategy.Config, 0, len(configs)*len(values))
		for _, cfg := range configs {
			for _, value := range values {
				candidate := cfg
				if err := applyParam(&candidate, key, value); err != nil {
					return nil, err
				}
				next = append(next, candidate)
			}
		}
		configs = next
	}

	return configs, nil
}

// RunSweep executes every config candidate against the same candle set.
func RunSweep(ctx context.Context, candles []*domain.Candle, runnerCfg RunnerConfig, spec SweepSpec) ([]SweepResult, error) {
	configs, err := Expand(spec)
	if err != nil {
		return nil, err
	}

	results := make([]SweepResult, 0, len(configs))
	for _, cfg := range configs {
		strat, configJSON, buildErr := strategy.Build(cfg)
		if buildErr != nil {
			results = append(results, SweepResult{Config: cfg, Err: buildErr})
			continue
		}

		engine := backtest.NewEngine(&backtest.Config{
			InitialBalance: runnerCfg.InitialBalance,
			Commission:     runnerCfg.Commission,
			Strategy:       strat,
			Persist:        runnerCfg.Persist,
			DB:             runnerCfg.DB,
			Symbol:         runnerCfg.Symbol,
			Interval:       runnerCfg.Interval,
			ConfigJSON:     configJSON,
		})

		result, runErr := engine.Run(ctx, candles)
		results = append(results, SweepResult{
			Config:     cfg.Normalize(),
			ConfigJSON: configJSON,
			Result:     result,
			Err:        runErr,
		})
	}

	sortResults(results, spec.SortBy)
	return results, nil
}

// Summarize builds an aggregate view over the sorted sweep results.
func Summarize(results []SweepResult, sortBy string) SweepSummary {
	successful := make([]SweepResult, 0, len(results))
	failed := 0
	for _, item := range results {
		if item.Err != nil || item.Result == nil {
			failed++
			continue
		}
		successful = append(successful, item)
	}

	summary := SweepSummary{
		TotalCandidates: len(results),
		SuccessfulRuns:  len(successful),
		FailedRuns:      failed,
		SortBy:          normalizeMetric(sortBy),
	}

	if len(successful) == 0 {
		return summary
	}

	best := successful[0]
	median := successful[len(successful)/2]
	worst := successful[len(successful)-1]
	summary.Best = &best
	summary.Median = &median
	summary.Worst = &worst
	return summary
}

func applyParam(cfg *strategy.Config, key, raw string) error {
	switch strings.ToLower(key) {
	case "fast", "fast_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.FastPeriod = v
	case "slow", "slow_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.SlowPeriod = v
	case "rsi", "rsi_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.RSIPeriod = v
	case "rsi_lower", "rsi_oversold":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.RSIOversold = v
	case "rsi_upper", "rsi_overbought":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.RSIOverbought = v
	case "bb", "bb_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.BBPeriod = v
	case "bb_mult", "bb_multiplier":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.BBMultiplier = v
	case "dca_period":
		cfg.DCAPeriod = raw
	case "dca_amount", "dca_amount_usdt":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.DCAAmountUSDT = v
	case "golden_fast", "golden_fast_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenFastPeriod = v
	case "golden_slow", "golden_slow_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenSlowPeriod = v
	case "golden_rsi", "golden_rsi_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenRSIPeriod = v
	case "golden_rsi_lower", "golden_rsi_lower_bound":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenRSILowerBound = v
	case "golden_rsi_upper", "golden_rsi_upper_bound":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenRSIUpperBound = v
	case "golden_bb", "golden_bb_period":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenBBPeriod = v
	case "golden_bb_mult", "golden_bb_multiplier":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenBBMultiplier = v
	case "vol_threshold", "golden_volume_threshold":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenVolumeThreshold = v
	case "tp", "golden_take_profit_pct":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenTakeProfitPct = v
	case "sl", "golden_stop_loss_pct":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.GoldenStopLossPct = v
	case "position", "position_size":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parse %s=%q: %w", key, raw, err)
		}
		cfg.PositionSize = v
	default:
		return fmt.Errorf("unsupported sweep parameter %q", key)
	}
	return nil
}

func sortResults(results []SweepResult, sortBy string) {
	metric := normalizeMetric(sortBy)

	score := func(item SweepResult) float64 {
		if item.Err != nil || item.Result == nil {
			return -1e18
		}
		switch metric {
		case "return", "total_return":
			return item.Result.TotalReturn
		case "calmar":
			return item.Result.CalmarRatio
		case "profit_factor":
			return item.Result.ProfitFactor
		case "win_rate":
			return item.Result.WinRate
		case "mdd", "max_drawdown_pct":
			return -item.Result.MaxDrawdownPct
		default:
			return item.Result.SharpeRatio
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return score(results[i]) > score(results[j])
	})
}

func normalizeMetric(sortBy string) string {
	metric := strings.ToLower(strings.TrimSpace(sortBy))
	if metric == "" {
		return "sharpe"
	}
	return metric
}
