package optimize

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/lavumi/crypto-quant/internal/datasource/database"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
)

func TestPersistSweepStoresJobAndResults(t *testing.T) {
	db := openTestDB(t)

	results := []SweepResult{
		{
			Config:     strategy.Config{Name: strategy.NameMACross, FastPeriod: 5, SlowPeriod: 20},
			ConfigJSON: `{"name":"ma_cross","fast_period":5,"slow_period":20}`,
			Result: &backtest.Result{
				TotalReturn:    0.12,
				SharpeRatio:    1.4,
				CalmarRatio:    0.9,
				MaxDrawdownPct: 0.08,
				ProfitFactor:   1.3,
				WinRate:        0.55,
				TotalTrades:    18,
			},
		},
	}

	id, err := PersistSweep(context.Background(), db, RunnerConfig{
		Symbol:   "BTCUSDT",
		Interval: "1h",
	}, SweepSpec{
		BaseConfig: strategy.Config{Name: strategy.NameMACross},
		Parameters: map[string][]string{"fast": []string{"5"}},
		SortBy:     "sharpe",
	}, results)
	if err != nil {
		t.Fatalf("PersistSweep returned error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive sweep id, got %d", id)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM backtest_sweep_results WHERE sweep_id = ?`, id); err != nil {
		t.Fatalf("failed to count sweep results: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 persisted sweep result, got %d", count)
	}
}

func TestPersistWalkForwardStoresRunAndWindows(t *testing.T) {
	db := openTestDB(t)

	summary := &WalkForwardSummary{
		Completed:       1,
		SelectionMetric: "sharpe",
		Windows: []WalkForwardResult{
			{
				Window: Window{
					TrainStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					TrainEnd:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
					TestStart:  time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
					TestEnd:    time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
				},
				Selected: &SweepResult{
					ConfigJSON: `{"name":"ma_cross","fast_period":5,"slow_period":20}`,
					Result: &backtest.Result{
						SharpeRatio: 1.1,
					},
				},
				OutOfSample: &SweepResult{
					Result: &backtest.Result{
						TotalReturn:    0.07,
						SharpeRatio:    0.6,
						MaxDrawdownPct: 0.05,
					},
				},
			},
		},
	}

	id, err := PersistWalkForward(context.Background(), db, RunnerConfig{
		Symbol:   "BTCUSDT",
		Interval: "1h",
	}, WalkForwardSpec{
		SweepSpec: SweepSpec{
			BaseConfig: strategy.Config{Name: strategy.NameMACross},
			Parameters: map[string][]string{"fast": []string{"5"}, "slow": []string{"20"}},
		},
		TrainDuration:   180 * 24 * time.Hour,
		TestDuration:    60 * 24 * time.Hour,
		StepDuration:    60 * 24 * time.Hour,
		SelectionMetric: "sharpe",
	}, summary)
	if err != nil {
		t.Fatalf("PersistWalkForward returned error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive walk-forward id, got %d", id)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM walk_forward_windows WHERE walk_forward_id = ?`, id); err != nil {
		t.Fatalf("failed to count walk-forward windows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 persisted walk-forward window, got %d", count)
	}
}

func openTestDB(t *testing.T) *database.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "optimize-test.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}
