package backtest

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// Strategy defines the interface for trading strategies
type Strategy interface {
	// Initialize is called before backtesting starts
	Initialize(ctx context.Context) error

	// OnCandle is called for each candle in the backtest
	OnCandle(ctx context.Context, candle *domain.Candle) (*Signal, error)

	// Name returns the strategy name
	Name() string
}

// Signal represents a trading signal
type Signal struct {
	Action   domain.OrderSide // BUY or SELL
	Quantity float64          // Amount to trade
	Price    float64          // Limit price (0 for market order)
	Reason   string           // Reason for the signal
}

// Engine executes backtesting
type Engine struct {
	strategy       Strategy
	initialBalance float64
	commission     float64 // Commission rate (e.g., 0.001 for 0.1%)

	// State
	balance  float64
	position float64 // Current position size
	trades   []*Trade
	equity   []EquityPoint
}

// Trade represents a backtesting trade
type Trade struct {
	Timestamp time.Time
	Side      domain.OrderSide
	Price     float64
	Quantity  float64
	Fee       float64
	Balance   float64
	Position  float64
	Reason    string
}

// EquityPoint represents equity at a point in time
type EquityPoint struct {
	Timestamp time.Time
	Equity    float64
	Price     float64
}

// Config holds backtesting configuration
type Config struct {
	InitialBalance float64
	Commission     float64
	Strategy       Strategy
}

// NewEngine creates a new backtesting engine
func NewEngine(cfg *Config) *Engine {
	return &Engine{
		strategy:       cfg.Strategy,
		initialBalance: cfg.InitialBalance,
		commission:     cfg.Commission,
		balance:        cfg.InitialBalance,
		position:       0,
		trades:         make([]*Trade, 0),
		equity:         make([]EquityPoint, 0),
	}
}

// Run executes the backtest with the given candles
func (e *Engine) Run(ctx context.Context, candles []*domain.Candle) (*Result, error) {
	log.Printf("Starting backtest with %d candles", len(candles))
	log.Printf("Initial balance: %.2f, Commission: %.4f%%", e.initialBalance, e.commission*100)

	// Initialize strategy
	if err := e.strategy.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize strategy: %w", err)
	}

	// Process each candle
	for i, candle := range candles {
		// Generate signal
		signal, err := e.strategy.OnCandle(ctx, candle)
		if err != nil {
			return nil, fmt.Errorf("strategy error at candle %d: %w", i, err)
		}

		// Execute signal if present
		if signal != nil {
			if err := e.executeSignal(candle, signal); err != nil {
				log.Printf("Failed to execute signal: %v", err)
			}
		}

		// Record equity
		equity := e.calculateEquity(candle.Close)
		e.equity = append(e.equity, EquityPoint{
			Timestamp: candle.OpenTime,
			Equity:    equity,
			Price:     candle.Close,
		})
	}

	// Calculate final metrics
	result := e.calculateResult()

	log.Printf("Backtest complete: Final equity: %.2f, Total return: %.2f%%",
		result.FinalEquity, result.TotalReturn*100)

	return result, nil
}

// executeSignal executes a trading signal
func (e *Engine) executeSignal(candle *domain.Candle, signal *Signal) error {
	price := signal.Price
	if price == 0 {
		price = candle.Close // Market order uses close price
	}

	switch signal.Action {
	case domain.OrderSideBuy:
		return e.executeBuy(candle.OpenTime, price, signal.Quantity, signal.Reason)
	case domain.OrderSideSell:
		return e.executeSell(candle.OpenTime, price, signal.Quantity, signal.Reason)
	default:
		return fmt.Errorf("unknown order side: %s", signal.Action)
	}
}

// executeBuy executes a buy order
func (e *Engine) executeBuy(timestamp time.Time, price, quantity float64, reason string) error {
	cost := price * quantity
	fee := cost * e.commission
	totalCost := cost + fee

	if totalCost > e.balance {
		return fmt.Errorf("insufficient balance: need %.2f, have %.2f", totalCost, e.balance)
	}

	e.balance -= totalCost
	e.position += quantity

	trade := &Trade{
		Timestamp: timestamp,
		Side:      domain.OrderSideBuy,
		Price:     price,
		Quantity:  quantity,
		Fee:       fee,
		Balance:   e.balance,
		Position:  e.position,
		Reason:    reason,
	}
	e.trades = append(e.trades, trade)

	log.Printf("BUY: %.8f @ %.2f (Fee: %.2f) - Balance: %.2f, Position: %.8f - %s",
		quantity, price, fee, e.balance, e.position, reason)

	return nil
}

// executeSell executes a sell order
func (e *Engine) executeSell(timestamp time.Time, price, quantity float64, reason string) error {
	if quantity > e.position {
		return fmt.Errorf("insufficient position: need %.8f, have %.8f", quantity, e.position)
	}

	revenue := price * quantity
	fee := revenue * e.commission
	netRevenue := revenue - fee

	e.balance += netRevenue
	e.position -= quantity

	trade := &Trade{
		Timestamp: timestamp,
		Side:      domain.OrderSideSell,
		Price:     price,
		Quantity:  quantity,
		Fee:       fee,
		Balance:   e.balance,
		Position:  e.position,
		Reason:    reason,
	}
	e.trades = append(e.trades, trade)

	log.Printf("SELL: %.8f @ %.2f (Fee: %.2f) - Balance: %.2f, Position: %.8f - %s",
		quantity, price, fee, e.balance, e.position, reason)

	return nil
}

// calculateEquity calculates current equity (balance + position value)
func (e *Engine) calculateEquity(currentPrice float64) float64 {
	return e.balance + (e.position * currentPrice)
}

// GetTrades returns all trades
func (e *Engine) GetTrades() []*Trade {
	return e.trades
}

// GetEquity returns equity curve
func (e *Engine) GetEquity() []EquityPoint {
	return e.equity
}



