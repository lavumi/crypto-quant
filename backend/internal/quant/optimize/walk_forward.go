package optimize

import (
	"context"
	"fmt"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// Window defines one train/test split for walk-forward validation.
type Window struct {
	TrainStart time.Time
	TrainEnd   time.Time
	TestStart  time.Time
	TestEnd    time.Time
}

// WalkForwardSpec defines how windows are generated and evaluated.
type WalkForwardSpec struct {
	SweepSpec       SweepSpec
	TrainDuration   time.Duration
	TestDuration    time.Duration
	StepDuration    time.Duration
	MaxWindows      int
	SelectionMetric string
}

// WalkForwardResult captures one train/test cycle.
type WalkForwardResult struct {
	Window      Window
	Training    []SweepResult
	Selected    *SweepResult
	OutOfSample *SweepResult
}

// WalkForwardSummary aggregates multiple windows.
type WalkForwardSummary struct {
	Windows         []WalkForwardResult
	Completed       int
	SelectionMetric string
}

// BuildWindows generates rolling train/test windows over the candle span.
func BuildWindows(candles []*domain.Candle, spec WalkForwardSpec) ([]Window, error) {
	if len(candles) == 0 {
		return nil, fmt.Errorf("cannot build walk-forward windows without candles")
	}
	if spec.TrainDuration <= 0 || spec.TestDuration <= 0 {
		return nil, fmt.Errorf("train and test durations must be positive")
	}

	step := spec.StepDuration
	if step <= 0 {
		step = spec.TestDuration
	}

	start := candles[0].OpenTime
	end := candles[len(candles)-1].OpenTime
	windows := make([]Window, 0)

	for trainStart := start; ; trainStart = trainStart.Add(step) {
		trainEnd := trainStart.Add(spec.TrainDuration)
		testStart := trainEnd
		testEnd := testStart.Add(spec.TestDuration)

		if !testEnd.Before(end) && !testEnd.Equal(end) {
			break
		}

		windows = append(windows, Window{
			TrainStart: trainStart,
			TrainEnd:   trainEnd,
			TestStart:  testStart,
			TestEnd:    testEnd,
		})

		if spec.MaxWindows > 0 && len(windows) >= spec.MaxWindows {
			break
		}
	}

	if len(windows) == 0 {
		return nil, fmt.Errorf("no valid walk-forward windows fit within the candle range")
	}

	return windows, nil
}

// RunWalkForward performs training sweeps followed by out-of-sample evaluation.
func RunWalkForward(ctx context.Context, candles []*domain.Candle, runnerCfg RunnerConfig, spec WalkForwardSpec) (*WalkForwardSummary, error) {
	windows, err := BuildWindows(candles, spec)
	if err != nil {
		return nil, err
	}

	selectionMetric := spec.SelectionMetric
	if selectionMetric == "" {
		selectionMetric = spec.SweepSpec.SortBy
	}
	if selectionMetric == "" {
		selectionMetric = "sharpe"
	}

	summary := &WalkForwardSummary{
		Windows:         make([]WalkForwardResult, 0, len(windows)),
		SelectionMetric: selectionMetric,
	}

	for _, window := range windows {
		trainCandles := filterCandles(candles, window.TrainStart, window.TrainEnd)
		testCandles := filterCandles(candles, window.TestStart, window.TestEnd)
		if len(trainCandles) == 0 || len(testCandles) == 0 {
			continue
		}

		training, err := RunSweep(ctx, trainCandles, RunnerConfig{
			InitialBalance: runnerCfg.InitialBalance,
			Commission:     runnerCfg.Commission,
			Persist:        false,
			Symbol:         runnerCfg.Symbol,
			Interval:       runnerCfg.Interval,
		}, SweepSpec{
			BaseConfig: spec.SweepSpec.BaseConfig,
			Parameters: spec.SweepSpec.Parameters,
			SortBy:     selectionMetric,
		})
		if err != nil {
			return nil, fmt.Errorf("training sweep failed for window %+v: %w", window, err)
		}

		selected := firstSuccessful(training)
		result := WalkForwardResult{
			Window:   window,
			Training: training,
			Selected: selected,
		}

		if selected != nil {
			testRun, err := RunSweep(ctx, testCandles, RunnerConfig{
				InitialBalance: runnerCfg.InitialBalance,
				Commission:     runnerCfg.Commission,
				Persist:        false,
				Symbol:         runnerCfg.Symbol,
				Interval:       runnerCfg.Interval,
			}, SweepSpec{
				BaseConfig: selected.Config,
				SortBy:     selectionMetric,
			})
			if err != nil {
				return nil, fmt.Errorf("test run failed for window %+v: %w", window, err)
			}
			result.OutOfSample = firstSuccessful(testRun)
		}

		summary.Windows = append(summary.Windows, result)
	}

	summary.Completed = len(summary.Windows)
	return summary, nil
}

func filterCandles(candles []*domain.Candle, start, end time.Time) []*domain.Candle {
	filtered := make([]*domain.Candle, 0)
	for _, candle := range candles {
		if (candle.OpenTime.Equal(start) || candle.OpenTime.After(start)) &&
			(candle.OpenTime.Before(end) || candle.OpenTime.Equal(end)) {
			filtered = append(filtered, candle)
		}
	}
	return filtered
}

func firstSuccessful(results []SweepResult) *SweepResult {
	for _, item := range results {
		if item.Err == nil && item.Result != nil {
			result := item
			return &result
		}
	}
	return nil
}
