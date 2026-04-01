package optimize

import (
	"testing"

	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
)

func TestExpandBuildsCartesianProduct(t *testing.T) {
	configs, err := Expand(SweepSpec{
		BaseConfig: strategy.Config{Name: strategy.NameMACross},
		Parameters: map[string][]string{
			"fast": []string{"5", "10"},
			"slow": []string{"20", "30"},
		},
	})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	if len(configs) != 4 {
		t.Fatalf("expected 4 configs, got %d", len(configs))
	}

	seen := map[[2]int]bool{}
	for _, cfg := range configs {
		seen[[2]int{cfg.FastPeriod, cfg.SlowPeriod}] = true
	}

	expected := [][2]int{{5, 20}, {5, 30}, {10, 20}, {10, 30}}
	for _, combo := range expected {
		if !seen[combo] {
			t.Fatalf("missing combo %v", combo)
		}
	}
}

func TestSortResultsByMetric(t *testing.T) {
	results := []SweepResult{
		{Config: strategy.Config{Name: strategy.NameMACross}, Result: &backtest.Result{SharpeRatio: 0.5, TotalReturn: 0.10}},
		{Config: strategy.Config{Name: strategy.NameMACross}, Result: &backtest.Result{SharpeRatio: 1.2, TotalReturn: 0.08}},
		{Config: strategy.Config{Name: strategy.NameMACross}, Result: &backtest.Result{SharpeRatio: 0.8, TotalReturn: 0.25}},
	}

	sortResults(results, "return")
	if results[0].Result.TotalReturn != 0.25 {
		t.Fatalf("expected highest return first, got %.2f", results[0].Result.TotalReturn)
	}

	sortResults(results, "sharpe")
	if results[0].Result.SharpeRatio != 1.2 {
		t.Fatalf("expected highest sharpe first, got %.2f", results[0].Result.SharpeRatio)
	}
}

func TestSummarizeBuildsBestMedianWorst(t *testing.T) {
	results := []SweepResult{
		{Config: strategy.Config{Name: strategy.NameMACross}, Result: &backtest.Result{SharpeRatio: 1.2, TotalReturn: 0.15}},
		{Config: strategy.Config{Name: strategy.NameMACross}, Result: &backtest.Result{SharpeRatio: 0.4, TotalReturn: 0.05}},
		{Config: strategy.Config{Name: strategy.NameMACross}, Result: &backtest.Result{SharpeRatio: -0.2, TotalReturn: -0.03}},
		{Config: strategy.Config{Name: strategy.NameMACross}, Err: assertErr{}},
	}

	sortResults(results, "sharpe")
	summary := Summarize(results, "sharpe")

	if summary.TotalCandidates != 4 || summary.SuccessfulRuns != 3 || summary.FailedRuns != 1 {
		t.Fatalf("unexpected counts: %+v", summary)
	}
	if summary.Best == nil || summary.Best.Result.SharpeRatio != 1.2 {
		t.Fatalf("unexpected best result: %+v", summary.Best)
	}
	if summary.Median == nil || summary.Median.Result.SharpeRatio != 0.4 {
		t.Fatalf("unexpected median result: %+v", summary.Median)
	}
	if summary.Worst == nil || summary.Worst.Result.SharpeRatio != -0.2 {
		t.Fatalf("unexpected worst result: %+v", summary.Worst)
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }
