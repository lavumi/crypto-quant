# Backtesting Guide

## Overview

The backtesting engine allows you to test trading strategies against historical market data to evaluate their performance before risking real capital.

## Features

- **Historical Data Collection**: Automatically fetches and stores historical candle data from Binance
- **Strategy Interface**: Clean interface for implementing custom trading strategies
- **Performance Metrics**: 
  - Total Return
  - Sharpe Ratio
  - Maximum Drawdown (MDD)
  - Win Rate
  - Trade Statistics
- **Commission Modeling**: Realistic trading fees simulation
- **Flexible Configuration**: Command-line parameters for easy testing

## Quick Start

### 1. Prepare Configuration

Create or update `configs/config.yaml`:

```yaml
exchange:
  type: binance
  binance:
    api_key: "your-api-key"
    secret_key: "your-secret-key"
    use_testnet: false

portfolio:
  initial_balances:
    USDT: 10000.0

trading:
  symbols:
    - BTCUSDT
    - ETHUSDT
  update_interval_sec: 5
```

### 2. Build the Binary

```bash
go build -o bin/backtest cmd/backtest/main.go
```

### 3. Run a Backtest

```bash
# Simple backtest with defaults
./bin/backtest

# Custom parameters
./bin/backtest \
  -symbol BTCUSDT \
  -interval 1h \
  -start 2024-01-01 \
  -end 2024-10-01 \
  -balance 10000 \
  -fast 10 \
  -slow 30
```

## Understanding the Results

### Sample Output

```
========== Backtest Results ==========
Time Period:
  Start: 2024-01-01 00:00:00
  End:   2024-10-01 00:00:00
  Duration: 6552h0m0s

Financial Performance:
  Initial Balance: $10000.00
  Final Equity:    $12450.75
  Total Return:    24.51%

Risk Metrics:
  Sharpe Ratio:    1.85
  Max Drawdown:    $850.25 (8.50%)

Trade Statistics:
  Total Trades:    45
  Winning Trades:  28
  Losing Trades:   17
  Win Rate:        62.22%
======================================
```

### Key Metrics Explained

#### 1. **Total Return**
- Percentage gain/loss from initial capital
- Formula: `(Final Equity - Initial Balance) / Initial Balance * 100`

#### 2. **Sharpe Ratio**
- Risk-adjusted return metric
- Higher is better (>1 is good, >2 is excellent)
- Measures excess return per unit of risk
- Formula: `(Average Return - Risk Free Rate) / Standard Deviation`

#### 3. **Maximum Drawdown (MDD)**
- Largest peak-to-trough decline
- Important for understanding worst-case scenarios
- Lower is better
- Example: 8.50% MDD means account dropped 8.5% from peak at worst

#### 4. **Win Rate**
- Percentage of profitable trades
- Note: High win rate doesn't always mean profitability
- Consider alongside average win/loss size

## Built-in Strategies

### Moving Average Crossover (MA Cross)

A simple trend-following strategy:

**Buy Signal**: Fast MA crosses above Slow MA (Golden Cross)
**Sell Signal**: Fast MA crosses below Slow MA (Death Cross)

**Parameters**:
- `fast`: Fast MA period (default: 10)
- `slow`: Slow MA period (default: 30)

**Example**:
```bash
./bin/backtest -fast 10 -slow 30
```

**Common Combinations**:
- Short-term: 5/20
- Medium-term: 10/30, 20/50
- Long-term: 50/200 (Golden/Death Cross)

## Creating Custom Strategies

### Strategy Interface

```go
type Strategy interface {
    // Initialize is called before backtesting starts
    Initialize(ctx context.Context) error
    
    // OnCandle is called for each candle in the backtest
    OnCandle(ctx context.Context, candle *types.Candle) (*Signal, error)
    
    // Name returns the strategy name
    Name() string
}
```

### Example: RSI Strategy

```go
package strategy

import (
    "context"
    "github.com/lavumi/crypto-quant/internal/backtest"
    "github.com/lavumi/crypto-quant/pkg/types"
)

type RSIStrategy struct {
    period       int
    oversold     float64
    overbought   float64
    prices       []float64
}

func NewRSIStrategy(period int, oversold, overbought float64) *RSIStrategy {
    return &RSIStrategy{
        period:     period,
        oversold:   oversold,
        overbought: overbought,
        prices:     make([]float64, 0),
    }
}

func (s *RSIStrategy) Name() string {
    return "RSI_Strategy"
}

func (s *RSIStrategy) Initialize(ctx context.Context) error {
    s.prices = make([]float64, 0)
    return nil
}

func (s *RSIStrategy) OnCandle(ctx context.Context, candle *types.Candle) (*backtest.Signal, error) {
    s.prices = append(s.prices, candle.Close)
    
    if len(s.prices) < s.period+1 {
        return nil, nil
    }
    
    rsi := s.calculateRSI()
    
    if rsi < s.oversold {
        return &backtest.Signal{
            Action:   types.OrderSideBuy,
            Quantity: 0.01,
            Price:    0,
            Reason:   fmt.Sprintf("RSI Oversold: %.2f", rsi),
        }, nil
    }
    
    if rsi > s.overbought {
        return &backtest.Signal{
            Action:   types.OrderSideSell,
            Quantity: 0.01,
            Price:    0,
            Reason:   fmt.Sprintf("RSI Overbought: %.2f", rsi),
        }, nil
    }
    
    return nil, nil
}

func (s *RSIStrategy) calculateRSI() float64 {
    // RSI calculation logic
    // ... implementation details
    return 50.0 // placeholder
}
```

## Tips for Better Backtesting

### 1. Data Quality
- Use sufficient historical data (at least 6 months)
- Check for data gaps or anomalies
- Consider different market conditions (bull, bear, sideways)

### 2. Realistic Assumptions
- Include commission fees (0.1% is typical for Binance)
- Consider slippage in high-volatility periods
- Account for position sizing constraints

### 3. Overfitting Prevention
- Don't over-optimize parameters
- Test on out-of-sample data
- Use walk-forward analysis
- Keep strategies simple

### 4. Strategy Development Process
1. **Hypothesis**: Define trading logic
2. **Backtest**: Test on historical data
3. **Analyze**: Review metrics and trades
4. **Refine**: Adjust parameters if needed
5. **Validate**: Test on different periods/symbols
6. **Paper Trade**: Test in real-time without capital
7. **Live Trade**: Deploy with small capital

## Common Pitfalls

### Look-Ahead Bias
❌ **Wrong**: Using future data in decisions
```go
// Don't do this: using candle.CloseTime in the decision
if time.Now().After(candle.CloseTime) {
    // make decision based on future knowledge
}
```

✅ **Correct**: Only use data available at decision time
```go
// Use only data from current and previous candles
signal := s.analyzeCurrentCandle(candle)
```

### Survivorship Bias
- Don't only backtest on currently successful assets
- Consider delisted or failed coins if relevant

### Overfitting
- Too many parameters = curve fitting to historical data
- Simple strategies often outperform complex ones in live trading

### Ignoring Transaction Costs
- Always include realistic commission rates
- Consider spread and slippage

## Performance Optimization

### For Faster Backtests
```bash
# Use higher timeframes
./bin/backtest -interval 4h  # faster than 1h

# Test shorter periods first
./bin/backtest -start 2024-09-01 -end 2024-10-01
```

### For Multiple Symbol Testing
```bash
# Create a simple bash script
for symbol in BTCUSDT ETHUSDT BNBUSDT; do
    echo "Testing $symbol..."
    ./bin/backtest -symbol $symbol
done
```

## Next Steps

1. **Implement More Strategies**: RSI, MACD, Bollinger Bands
2. **Add Parameter Optimization**: Grid search, genetic algorithms
3. **Visualization**: Export equity curves for plotting
4. **Walk-Forward Analysis**: Rolling window backtesting
5. **Portfolio Backtesting**: Test multiple strategies simultaneously

## References

- [Quantitative Trading by Ernest Chan](https://www.amazon.com/Quantitative-Trading-Build-Algorithmic-Business/dp/1119800064)
- [Binance API Documentation](https://binance-docs.github.io/apidocs/spot/en/)
- [Common Backtesting Pitfalls](https://www.quantstart.com/articles/Successful-Backtesting-of-Algorithmic-Trading-Strategies-Part-I/)




