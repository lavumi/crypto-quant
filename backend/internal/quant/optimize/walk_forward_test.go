package optimize

import (
	"context"
	"testing"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
)

func TestBuildWindowsGeneratesRollingSplits(t *testing.T) {
	candles := makeCandles(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 12, 24*time.Hour)

	windows, err := BuildWindows(candles, WalkForwardSpec{
		TrainDuration: 3 * 24 * time.Hour,
		TestDuration:  2 * 24 * time.Hour,
		StepDuration:  2 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("BuildWindows returned error: %v", err)
	}

	if len(windows) != 4 {
		t.Fatalf("expected 4 windows, got %d", len(windows))
	}
	if !windows[0].TrainStart.Equal(candles[0].OpenTime) {
		t.Fatalf("unexpected first train start: %v", windows[0].TrainStart)
	}
	if !windows[0].TestStart.Equal(windows[0].TrainEnd) {
		t.Fatalf("expected test to begin when train ends")
	}
}

func TestRunWalkForwardProducesTrainingAndOutOfSampleRuns(t *testing.T) {
	candles := makeTrendingCandles(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 80, 24*time.Hour)

	summary, err := RunWalkForward(context.Background(), candles, RunnerConfig{
		InitialBalance: 10000,
		Commission:     0.001,
		Symbol:         "BTCUSDT",
		Interval:       "1d",
	}, WalkForwardSpec{
		SweepSpec: SweepSpec{
			BaseConfig: strategy.Config{Name: strategy.NameMACross},
			Parameters: map[string][]string{
				"fast": []string{"3", "5"},
				"slow": []string{"8", "10"},
			},
			SortBy: "sharpe",
		},
		TrainDuration: 20 * 24 * time.Hour,
		TestDuration:  10 * 24 * time.Hour,
		StepDuration:  10 * 24 * time.Hour,
		MaxWindows:    2,
	})
	if err != nil {
		t.Fatalf("RunWalkForward returned error: %v", err)
	}

	if summary.Completed != 2 {
		t.Fatalf("expected 2 completed windows, got %d", summary.Completed)
	}
	for _, window := range summary.Windows {
		if len(window.Training) == 0 {
			t.Fatalf("expected training results for window %+v", window.Window)
		}
		if window.Selected == nil {
			t.Fatalf("expected selected candidate for window %+v", window.Window)
		}
		if window.OutOfSample == nil {
			t.Fatalf("expected out-of-sample run for window %+v", window.Window)
		}
	}
}

func makeCandles(start time.Time, count int, step time.Duration) []*domain.Candle {
	candles := make([]*domain.Candle, 0, count)
	for i := range count {
		ts := start.Add(time.Duration(i) * step)
		candles = append(candles, &domain.Candle{
			OpenTime: ts,
			Close:    100 + float64(i),
			Volume:   1000,
		})
	}
	return candles
}

func makeTrendingCandles(start time.Time, count int, step time.Duration) []*domain.Candle {
	candles := make([]*domain.Candle, 0, count)
	price := 100.0
	for i := range count {
		if i%7 == 0 {
			price -= 2
		} else {
			price += 3
		}
		ts := start.Add(time.Duration(i) * step)
		candles = append(candles, &domain.Candle{
			OpenTime: ts,
			Close:    price,
			Volume:   1000 + float64(i),
		})
	}
	return candles
}
