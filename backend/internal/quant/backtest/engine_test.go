package backtest

import (
	"testing"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

func TestExecuteSignalAppliesSlippageAndFees(t *testing.T) {
	engine := NewEngine(&Config{
		InitialBalance: 1000,
		Commission:     0.001,
		SlippageModel:  BpsSlippageModel{BasisPoints: 100},
	})

	candle := &domain.Candle{
		OpenTime: time.Unix(0, 0),
		Close:    100,
	}

	err := engine.executeSignal(candle, &Signal{
		Action:   domain.OrderSideBuy,
		Quantity: 0.5,
		Reason:   "test buy",
	})
	if err != nil {
		t.Fatalf("executeSignal returned error: %v", err)
	}

	if len(engine.trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(engine.trades))
	}

	trade := engine.trades[0]
	if trade.Price != 101 {
		t.Fatalf("expected slipped price 101, got %.2f", trade.Price)
	}
	if trade.Fee <= 0 {
		t.Fatalf("expected positive fee, got %.8f", trade.Fee)
	}
}

func TestExecuteSignalBuysFullBalanceWithoutOverspending(t *testing.T) {
	engine := NewEngine(&Config{
		InitialBalance: 10000,
		Commission:     0.001,
	})

	candle := &domain.Candle{
		OpenTime: time.Unix(0, 0),
		Close:    100,
	}

	err := engine.executeSignal(candle, &Signal{
		Action:   domain.OrderSideBuy,
		Quantity: 1.0,
		Reason:   "full balance",
	})
	if err != nil {
		t.Fatalf("executeSignal returned error: %v", err)
	}

	if len(engine.trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(engine.trades))
	}

	if engine.balance < -1e-6 {
		t.Fatalf("expected non-negative remaining balance, got %.8f", engine.balance)
	}
}

func TestExecuteSignalRejectsOrderWhenRiskManagerBlocksExposure(t *testing.T) {
	engine := NewEngine(&Config{
		InitialBalance: 1000,
		Commission:     0.001,
		RiskManager:    MaxExposureRiskManager{MaxExposure: 0.25},
	})

	candle := &domain.Candle{
		OpenTime: time.Unix(0, 0),
		Close:    100,
	}

	err := engine.executeSignal(candle, &Signal{
		Action:   domain.OrderSideBuy,
		Quantity: 0.5,
		Reason:   "too large",
	})
	if err == nil {
		t.Fatalf("expected risk manager to reject order")
	}

	if len(engine.trades) != 0 {
		t.Fatalf("expected no trades after rejection, got %d", len(engine.trades))
	}
}

func TestExecuteSignalIgnoresSellWhenNoPosition(t *testing.T) {
	engine := NewEngine(&Config{
		InitialBalance: 1000,
		Commission:     0.001,
	})

	candle := &domain.Candle{
		OpenTime: time.Unix(0, 0),
		Close:    100,
	}

	err := engine.executeSignal(candle, &Signal{
		Action:   domain.OrderSideSell,
		Quantity: 1.0,
		Reason:   "no position",
	})
	if err != nil {
		t.Fatalf("expected no error for zero-position sell, got %v", err)
	}
	if len(engine.trades) != 0 {
		t.Fatalf("expected no trades, got %d", len(engine.trades))
	}
}
