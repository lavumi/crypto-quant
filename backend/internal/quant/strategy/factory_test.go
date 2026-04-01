package strategy

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildGoldenRSIBBWithDefaults(t *testing.T) {
	strat, configJSON, err := Build(Config{Name: NameGoldenRSIBB})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if strat == nil {
		t.Fatalf("expected strategy instance")
	}
	if !strings.Contains(strat.Name(), "GoldenRSIBB") {
		t.Fatalf("unexpected strategy name %q", strat.Name())
	}

	var cfg Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		t.Fatalf("failed to unmarshal config json: %v", err)
	}
	if cfg.GoldenFastPeriod != 5 || cfg.GoldenSlowPeriod != 20 {
		t.Fatalf("expected default golden MA periods, got %d/%d", cfg.GoldenFastPeriod, cfg.GoldenSlowPeriod)
	}
}

func TestBuildRejectsUnsupportedStrategy(t *testing.T) {
	_, _, err := Build(Config{Name: "nope"})
	if err == nil {
		t.Fatalf("expected unsupported strategy error")
	}
}

func TestBuildParsesDCAConfig(t *testing.T) {
	strat, configJSON, err := Build(Config{
		Name:          NameDCA,
		DCAPeriod:     "48h",
		DCAAmountUSDT: 250,
	})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if strat == nil {
		t.Fatalf("expected DCA strategy instance")
	}

	var cfg Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		t.Fatalf("failed to unmarshal config json: %v", err)
	}
	if cfg.DCAPeriod != "48h" || cfg.DCAAmountUSDT != 250 {
		t.Fatalf("unexpected DCA config: %+v", cfg)
	}
}
