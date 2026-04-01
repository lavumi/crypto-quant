package backtest

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/lavumi/crypto-quant/internal/datasource/database"
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
	Quantity float64          // Position size as percentage (0.0 - 1.0, e.g., 0.5 = 50% of balance/position)
	Price    float64          // Limit price (0 for market order)
	Reason   string           // Reason for the signal
}

// Engine executes backtesting
type Engine struct {
	strategy       Strategy
	initialBalance float64
	commission     float64 // Commission rate (e.g., 0.001 for 0.1%)
	feeModel       FeeModel
	slippageModel  SlippageModel
	riskManager    RiskManager

	// Persistence
	persist    bool
	db         *database.DB
	symbol     string
	interval   string
	configJSON string

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
	FeeModel       FeeModel
	SlippageModel  SlippageModel
	RiskManager    RiskManager
	// Optional persistence settings
	Persist    bool
	DB         *database.DB
	Symbol     string
	Interval   string
	ConfigJSON string
}

// NewEngine creates a new backtesting engine
func NewEngine(cfg *Config) *Engine {
	feeModel := cfg.FeeModel
	if feeModel == nil {
		feeModel = NewFixedRateFeeModel(cfg.Commission)
	}

	slippageModel := cfg.SlippageModel
	if slippageModel == nil {
		slippageModel = NoSlippageModel{}
	}

	riskManager := cfg.RiskManager
	if riskManager == nil {
		riskManager = NoopRiskManager{}
	}

	return &Engine{
		strategy:       cfg.Strategy,
		initialBalance: cfg.InitialBalance,
		commission:     cfg.Commission,
		feeModel:       feeModel,
		slippageModel:  slippageModel,
		riskManager:    riskManager,
		persist:        cfg.Persist,
		db:             cfg.DB,
		symbol:         cfg.Symbol,
		interval:       cfg.Interval,
		configJSON:     cfg.ConfigJSON,
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

	// Persist if configured
	if e.persist && e.db != nil {
		if err := e.persistRun(ctx, result); err != nil {
			log.Printf("Failed to persist backtest run: %v", err)
		}
	}

	return result, nil
}

// persistRun saves run summary, trades, and equity curve into DB
func (e *Engine) persistRun(ctx context.Context, result *Result) error {
	tx, err := e.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	res, err := tx.Exec(`
        INSERT INTO backtest_runs (
            strategy_name, symbol, interval, start_time, end_time,
            initial_balance, final_equity, total_return, sharpe_ratio, max_drawdown,
            max_drawdown_pct, win_rate, total_trades, commission, config_json
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
		e.strategy.Name(), e.symbol, e.interval,
		result.StartTime.Unix(), result.EndTime.Unix(),
		e.initialBalance, result.FinalEquity, result.TotalReturn, result.SharpeRatio, result.MaxDrawdown,
		result.MaxDrawdownPct, result.WinRate, result.TotalTrades, e.commission, e.configJSON,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insert run: %w", err)
	}

	runID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("last insert id: %w", err)
	}

	if len(e.trades) > 0 {
		stmtTrades, err := tx.Prepare(`
            INSERT INTO backtest_run_trades (
                run_id, timestamp, side, price, quantity, fee, balance, position, reason
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("prepare trades: %w", err)
		}
		for _, tr := range e.trades {
			if _, err := stmtTrades.Exec(runID, tr.Timestamp.Unix(), string(tr.Side), tr.Price, tr.Quantity, tr.Fee, tr.Balance, tr.Position, tr.Reason); err != nil {
				stmtTrades.Close()
				tx.Rollback()
				return fmt.Errorf("insert trade: %w", err)
			}
		}
		stmtTrades.Close()
	}

	if len(e.equity) > 0 {
		stmtEquity, err := tx.Prepare(`
            INSERT INTO backtest_run_equity (
                run_id, timestamp, equity, price
            ) VALUES (?, ?, ?, ?)
        `)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("prepare equity: %w", err)
		}
		for _, pt := range e.equity {
			if _, err := stmtEquity.Exec(runID, pt.Timestamp.Unix(), pt.Equity, pt.Price); err != nil {
				stmtEquity.Close()
				tx.Rollback()
				return fmt.Errorf("insert equity: %w", err)
			}
		}
		stmtEquity.Close()
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Printf("Saved backtest run: id=%d, strategy=%s, symbol=%s, interval=%s", runID, e.strategy.Name(), e.symbol, e.interval)
	return nil
}

// executeSignal executes a trading signal
func (e *Engine) executeSignal(candle *domain.Candle, signal *Signal) error {
	price := signal.Price
	if price == 0 {
		price = candle.Close // Market order uses close price
	}

	// Convert percentage to actual quantity
	var actualQuantity float64
	switch signal.Action {
	case domain.OrderSideBuy:
		// For buy: use percentage of available balance
		// Calculate how many coins we can buy with (balance * percentage)
		availableAmount := e.balance*signal.Quantity - 1
		// Truncate to 2 decimal places to prevent floating point errors
		availableAmount = math.Floor(availableAmount*100) / 100
		if availableAmount <= 0 {
			return fmt.Errorf("insufficient balance allocated for buy: %.2f", availableAmount)
		}
		actualQuantity = availableAmount / price
	case domain.OrderSideSell:
		// For sell: use percentage of current position
		actualQuantity = e.position * signal.Quantity
		if signal.Quantity <= 0 || signal.Quantity > 1 {
			actualQuantity = e.position
		}
	default:
		return fmt.Errorf("unknown order side: %s", signal.Action)
	}

	if actualQuantity <= 0 {
		return fmt.Errorf("non-positive order quantity: %.8f", actualQuantity)
	}

	intent := OrderIntent{
		Timestamp: candle.OpenTime,
		Side:      signal.Action,
		Price:     price,
		Quantity:  actualQuantity,
		Reason:    signal.Reason,
	}
	intent.Price = e.slippageModel.Apply(intent)
	if intent.Price <= 0 {
		return fmt.Errorf("invalid execution price: %.8f", intent.Price)
	}

	state := PortfolioState{
		Balance:  e.balance,
		Position: e.position,
		Equity:   e.calculateEquity(candle.Close),
	}

	if err := e.riskManager.Validate(intent, state); err != nil {
		return fmt.Errorf("risk check failed: %w", err)
	}

	switch intent.Side {
	case domain.OrderSideBuy:
		return e.executeBuy(intent)
	case domain.OrderSideSell:
		return e.executeSell(intent)
	default:
		return fmt.Errorf("unknown order side: %s", intent.Side)
	}
}

// executeBuy executes a buy order
func (e *Engine) executeBuy(intent OrderIntent) error {
	cost := intent.Price * intent.Quantity
	fee := e.feeModel.CalculateFee(intent)
	totalCost := cost + fee

	if totalCost > e.balance {
		return fmt.Errorf("insufficient balance: need %.2f, have %.2f", totalCost, e.balance)
	}

	e.balance -= totalCost
	e.position += intent.Quantity

	trade := &Trade{
		Timestamp: intent.Timestamp,
		Side:      domain.OrderSideBuy,
		Price:     intent.Price,
		Quantity:  intent.Quantity,
		Fee:       fee,
		Balance:   e.balance,
		Position:  e.position,
		Reason:    intent.Reason,
	}
	e.trades = append(e.trades, trade)

	log.Printf("BUY: %.8f @ %.2f (Fee: %.2f) - Balance: %.2f, Position: %.8f - %s",
		intent.Quantity, intent.Price, fee, e.balance, e.position, intent.Reason)

	return nil
}

// executeSell executes a sell order
func (e *Engine) executeSell(intent OrderIntent) error {
	if intent.Quantity > e.position {
		return fmt.Errorf("insufficient position: need %.8f, have %.8f", intent.Quantity, e.position)
	}

	revenue := intent.Price * intent.Quantity
	fee := e.feeModel.CalculateFee(intent)
	netRevenue := revenue - fee

	e.balance += netRevenue
	e.position -= intent.Quantity

	trade := &Trade{
		Timestamp: intent.Timestamp,
		Side:      domain.OrderSideSell,
		Price:     intent.Price,
		Quantity:  intent.Quantity,
		Fee:       fee,
		Balance:   e.balance,
		Position:  e.position,
		Reason:    intent.Reason,
	}
	e.trades = append(e.trades, trade)

	log.Printf("SELL: %.8f @ %.2f (Fee: %.2f) - Balance: %.2f, Position: %.8f - %s",
		intent.Quantity, intent.Price, fee, e.balance, e.position, intent.Reason)

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
