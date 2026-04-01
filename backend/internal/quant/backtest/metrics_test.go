package backtest

import (
	"math"
	"testing"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

func TestCalculateResultBuildsExtendedMetrics(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	engine := &Engine{
		initialBalance: 1000,
		balance:        1120,
		trades: []*Trade{
			{Timestamp: start, Side: domain.OrderSideBuy, Price: 100, Quantity: 1, Fee: 1},
			{Timestamp: start.Add(24 * time.Hour), Side: domain.OrderSideSell, Price: 120, Quantity: 1, Fee: 1},
			{Timestamp: start.Add(48 * time.Hour), Side: domain.OrderSideBuy, Price: 130, Quantity: 1, Fee: 1},
			{Timestamp: start.Add(72 * time.Hour), Side: domain.OrderSideSell, Price: 110, Quantity: 1, Fee: 1},
		},
		equity: []EquityPoint{
			{Timestamp: start, Equity: 1000},
			{Timestamp: start.Add(24 * time.Hour), Equity: 1200},
			{Timestamp: start.Add(48 * time.Hour), Equity: 1100},
			{Timestamp: start.AddDate(0, 1, 0), Equity: 1300},
			{Timestamp: start.AddDate(0, 1, 1), Equity: 1120},
		},
	}

	result := engine.calculateResult()

	if result.ProfitFactor <= 0 {
		t.Fatalf("expected positive profit factor, got %.4f", result.ProfitFactor)
	}
	if result.Expectancy == 0 {
		t.Fatalf("expected non-zero expectancy")
	}
	if len(result.DrawdownCurve) != len(engine.equity) {
		t.Fatalf("expected drawdown curve length %d, got %d", len(engine.equity), len(result.DrawdownCurve))
	}
	if len(result.MonthlyReturns) != 2 {
		t.Fatalf("expected 2 monthly return buckets, got %d", len(result.MonthlyReturns))
	}
	if result.AverageWin <= 0 {
		t.Fatalf("expected positive average win, got %.4f", result.AverageWin)
	}
	if result.AverageLoss >= 0 {
		t.Fatalf("expected negative average loss, got %.4f", result.AverageLoss)
	}

	expectedMaxDDPct := 180.0 / 1300.0
	if math.Abs(result.MaxDrawdownPct-expectedMaxDDPct) > 1e-9 {
		t.Fatalf("expected max drawdown pct %.8f, got %.8f", expectedMaxDDPct, result.MaxDrawdownPct)
	}
}

func TestCalculateMonthlyReturnsUsesCalendarBuckets(t *testing.T) {
	result := &Result{
		EquityCurve: []EquityPoint{
			{Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Equity: 1000},
			{Timestamp: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), Equity: 1100},
			{Timestamp: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Equity: 1210},
			{Timestamp: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC), Equity: 1180},
		},
	}

	monthly := result.calculateMonthlyReturns()
	if len(monthly) != 2 {
		t.Fatalf("expected 2 months, got %d", len(monthly))
	}

	if math.Abs(monthly[0].Return-0.10) > 1e-9 {
		t.Fatalf("expected january return 0.10, got %.8f", monthly[0].Return)
	}
	if math.Abs(monthly[1].Return-((1180.0-1100.0)/1100.0)) > 1e-9 {
		t.Fatalf("unexpected february return %.8f", monthly[1].Return)
	}
}
