# Golden Cross + RSI + Bollinger Bands Strategy

## Overview

This is an advanced trading strategy that combines multiple technical indicators to identify high-probability entry points and implement strict risk management.

## Strategy Name

`golden_rsi_bb`

## Entry Conditions (ALL must be true)

### 1. Golden Cross - MA5 > MA20
- Fast moving average (5-period) crosses above slow moving average (20-period)
- Indicates a bullish trend reversal
- Only enters when fast MA is higher than slow MA

### 2. RSI Range - 40 to 70
- Relative Strength Index must be between 40 and 70
- RSI < 40: Too weak, skip entry
- RSI > 70: Too hot (overbought), skip entry
- This range indicates healthy momentum without extreme conditions

### 3. Price > Bollinger Middle Band
- Current price must be above the Bollinger Bands middle line (SMA 20)
- Confirms the uptrend
- Middle band acts as dynamic support level

### 4. Volume Spike - 1.3x or Higher
- Current volume must be at least 1.3x the average volume (20-period average)
- High volume confirms the reliability of the signal
- Indicates strong market participation

## Exit Conditions (ANY triggers exit)

### 1. Take Profit - +6%
- Exit entire position when profit reaches 6%
- Calculated as: `current_price >= entry_price × 1.06`
- Locks in gains quickly

### 2. Stop Loss - -3%
- Exit entire position when loss reaches 3%
- Calculated as: `current_price <= entry_price × 0.97`
- Limits downside risk
- Risk-reward ratio: 1:2 (risk 3% to gain 6%)

### 3. Death Cross - MA5 < MA20
- Exit when fast MA crosses below slow MA
- Indicates the end of the uptrend
- Protects against trend reversal

## Parameters

All parameters can be customized via API request:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `golden_fast_period` | integer | 5 | Fast MA period for golden/death cross |
| `golden_slow_period` | integer | 20 | Slow MA period for golden/death cross |
| `golden_rsi_period` | integer | 14 | RSI calculation period |
| `golden_rsi_lower_bound` | float | 40 | RSI lower bound (entry requires RSI ≥ this) |
| `golden_rsi_upper_bound` | float | 70 | RSI upper bound (entry requires RSI ≤ this) |
| `golden_bb_period` | integer | 20 | Bollinger Bands period |
| `golden_bb_multiplier` | float | 2.0 | Bollinger Bands standard deviation multiplier |
| `golden_volume_threshold` | float | 1.3 | Volume multiplier threshold (1.3 = 130%) |
| `golden_take_profit_pct` | float | 0.06 | Take profit percentage (0.06 = 6%) |
| `golden_stop_loss_pct` | float | 0.03 | Stop loss percentage (0.03 = 3%) |
| `position_size` | float | 0.01 | Position size per trade |

## API Usage Example

### Basic Usage (with defaults)

```bash
curl -X POST http://localhost:8080/api/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-01",
    "end_date": "2025-10-17",
    "initial_balance": 10000,
    "commission": 0.001,
    "strategy": "golden_rsi_bb",
    "position_size": 0.01
  }'
```

### Custom Parameters

```bash
curl -X POST http://localhost:8080/api/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-01",
    "end_date": "2025-10-17",
    "initial_balance": 10000,
    "commission": 0.001,
    "strategy": "golden_rsi_bb",
    "position_size": 0.01,
    "golden_fast_period": 5,
    "golden_slow_period": 20,
    "golden_rsi_period": 14,
    "golden_rsi_lower_bound": 40,
    "golden_rsi_upper_bound": 70,
    "golden_bb_period": 20,
    "golden_bb_multiplier": 2.0,
    "golden_volume_threshold": 1.3,
    "golden_take_profit_pct": 0.06,
    "golden_stop_loss_pct": 0.03
  }'
```

### Conservative Parameters (Tighter RSI Range)

```bash
curl -X POST http://localhost:8080/api/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-01",
    "end_date": "2025-10-17",
    "initial_balance": 10000,
    "commission": 0.001,
    "strategy": "golden_rsi_bb",
    "position_size": 0.01,
    "golden_rsi_lower_bound": 45,
    "golden_rsi_upper_bound": 65,
    "golden_volume_threshold": 1.5
  }'
```

### Aggressive Parameters (Wider Take Profit)

```bash
curl -X POST http://localhost:8080/api/backtest/run \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "interval": "1h",
    "start_date": "2025-07-01",
    "end_date": "2025-10-17",
    "initial_balance": 10000,
    "commission": 0.001,
    "strategy": "golden_rsi_bb",
    "position_size": 0.01,
    "golden_take_profit_pct": 0.10,
    "golden_stop_loss_pct": 0.05
  }'
```

## Strategy Logic Flow

```
For each candle:
  1. Calculate MA5, MA20, RSI, Bollinger Bands, Average Volume
  
  2. If IN position:
     a. Check profit/loss
        - If profit >= 6% → SELL (take profit)
        - If loss >= 3% → SELL (stop loss)
     b. Check for death cross
        - If MA5 < MA20 → SELL (trend reversal)
  
  3. If NOT in position:
     a. Check all entry conditions:
        - MA5 > MA20? (Golden Cross)
        - 40 <= RSI <= 70? (Healthy momentum)
        - Price > BB Middle? (Above support)
        - Volume >= Avg Vol × 1.3? (High volume)
     b. If ALL conditions met → BUY
```

## Risk Management

- **Risk-Reward Ratio**: 1:2 (risk 3% to gain 6%)
- **Position Sizing**: Configurable, default 0.01 BTC
- **Maximum Loss per Trade**: 3% of entry price
- **Trend Following**: Exits on trend reversal (death cross)
- **Volume Confirmation**: Requires 30% above average volume for entry

## Advantages

1. **Multi-Indicator Confirmation**: Reduces false signals by requiring multiple conditions
2. **Volume Filter**: Ensures market participation and signal reliability
3. **Clear Exit Rules**: Defined take profit and stop loss levels
4. **Trend Awareness**: Uses MA crossover for entry and exit
5. **Risk Management**: Strict stop loss and quick take profit

## Disadvantages

1. **Strict Entry Conditions**: May miss some opportunities due to multiple filters
2. **Quick Exit**: 6% take profit might exit too early in strong trends
3. **Parameter Sensitivity**: Performance may vary with different parameter values
4. **Choppy Markets**: May generate false signals in sideways markets

## Recommended Usage

- **Market Conditions**: Best in trending markets with clear momentum
- **Timeframes**: Works well on 1h, 4h, or daily timeframes
- **Symbols**: Suitable for liquid cryptocurrencies with good volume
- **Risk Tolerance**: Moderate risk with defined stop loss

## Backtesting Tips

1. Start with default parameters
2. Test on historical data from different market conditions (bull, bear, sideways)
3. Adjust RSI bounds for different volatility levels
4. Consider increasing volume threshold in highly volatile periods
5. Monitor win rate and Sharpe ratio to evaluate performance

