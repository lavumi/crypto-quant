# Backtest Quality Plan

## Goal

This document describes how to raise the backtesting system from a functional prototype into a research-grade evaluation tool.

The current project already supports:

- historical candle collection
- strategy-based backtests
- core performance metrics
- parameter sweep
- walk-forward skeleton
- persistence for sweep and walk-forward runs

The next step is improving the realism, reliability, and interpretability of the results.

## Current State

The system is already useful for strategy experiments, but it still has several limitations:

- execution assumptions are still simplified
- only candle-based simulation is used
- position sizing is strategy-coupled in many cases
- risk constraints are minimal
- result interpretation is mostly metric-driven, not regime-aware
- there is no formal benchmark comparison
- out-of-sample validation exists structurally, but reporting is still basic

In practice, this means the platform can rank parameter sets, but it is not yet strong enough to trust raw performance numbers without caution.

## Quality Priorities

### 1. Execution Realism

Backtests should become harder to game and closer to live trading conditions.

Priority items:

- separate maker and taker fee models
- add configurable slippage models by order type
- model partial fill assumptions for large orders
- support spread-aware execution assumptions
- distinguish signal price from execution price more explicitly
- make execution assumptions visible in saved configs and reports

Recommended implementation direction:

- keep `FeeModel`, `SlippageModel`, and `RiskManager` as pluggable interfaces
- add named execution profiles such as `optimistic`, `base`, and `conservative`
- store the selected execution profile in run metadata

### 2. Data Quality Controls

A backtest is only as good as the input data.

Priority items:

- detect missing candles by interval
- detect duplicate candles
- validate monotonic time ordering
- validate symbol and interval consistency
- flag suspicious candles such as zero volume or impossible OHLC structure
- add collection coverage reporting before a backtest runs

Recommended implementation direction:

- create a `data validation report` before backtest execution
- optionally block runs when data quality falls below a threshold
- expose data completeness in API and UI later

### 3. Portfolio and Sizing Quality

Strategies should generate signals, not directly own capital allocation logic.

Priority items:

- move position sizing rules out of strategies where possible
- support fixed-fraction sizing
- support fixed-risk sizing based on stop distance
- support symbol-level exposure caps
- support portfolio-level max exposure

Recommended implementation direction:

- introduce a `Sizer` layer between strategy and execution
- let strategies express intent, and let sizing decide actual order size

### 4. Risk Controls

A research-grade backtest should include the same kinds of controls that would exist in production.

Priority items:

- max position size
- max concurrent positions
- max daily loss
- max drawdown cutoff
- cooldown after loss streak
- volatility-based trade blocking
- kill switch behavior

Recommended implementation direction:

- extend `RiskManager` with reusable policy modules
- make blocked trades visible in reports
- distinguish between strategy weakness and risk-policy suppression

### 5. Metrics and Reporting Depth

The system should help answer "why did this strategy work or fail?"

Priority items:

- profit factor
- expectancy
- average win / average loss
- monthly returns
- drawdown curve
- benchmark-relative return
- rolling Sharpe
- trade duration metrics
- regime split metrics

Recommended implementation direction:

- keep raw run results separate from derived report metrics
- add benchmark support such as buy-and-hold BTC/ETH over the same period
- add aggregate walk-forward summary metrics

## Validation Standards

### Parameter Sweep

Parameter sweep should be treated as a discovery tool, not an optimizer to blindly trust.

Rules:

- prefer stable regions over sharp peaks
- compare multiple ranking metrics, not only return
- inspect the distribution of near-top candidates
- store the full sweep, not only the top result

Good signal:

- many nearby parameter sets remain acceptable

Bad signal:

- only one narrow combination performs well

### Walk-Forward Validation

Walk-forward should become the default standard for serious evaluation.

Purpose:

- optimize on one period
- test on the next unseen period
- repeat over time

This helps identify:

- overfitting
- regime dependence
- unstable parameter selection

Walk-forward reporting should eventually include:

- per-window selected params
- per-window out-of-sample return
- per-window out-of-sample Sharpe
- pass/fail consistency by window
- average and median out-of-sample metrics
- positive-window ratio

## Recommended Build Order

### Phase A: Backtest Engine Quality

1. Improve fee and slippage realism
2. Add data quality validation
3. Expand risk policies
4. Separate sizing from strategy logic

### Phase B: Research Quality

1. Improve sweep summaries
2. Improve walk-forward summaries
3. Add benchmark comparison
4. Add regime-aware metrics

### Phase C: Product Quality

1. Readonly API for sweep and walk-forward retrieval
2. UI for experiment list and detail views
3. Comparison screens for run vs benchmark
4. Visualization of sweep and walk-forward distributions

## Concrete Near-Term Tasks

The most practical next tasks for this repository are:

1. Add benchmark metrics to every backtest result
2. Add average walk-forward out-of-sample summary fields
3. Add data validation reporting before execution
4. Add more realistic execution presets
5. Add readonly endpoints for persisted sweep and walk-forward runs

## Example Commands

These commands are useful as a starting point for real experiments in the current repository.

### Collect Candle Data

```bash
cd /Users/lavumi/private/crypto-quant

./server collect --symbol BTCUSDT --interval 1h --days 365 --db data/trading.db
```

### Run a Parameter Sweep

```bash
cd /Users/lavumi/private/crypto-quant

./server sweep \
  --strategy golden_rsi_bb \
  --symbol BTCUSDT \
  --interval 1h \
  --start 2025-01-01 \
  --end 2025-12-31 \
  --param golden_fast=5,10 \
  --param golden_slow=20,30 \
  --param tp=0.04,0.06 \
  --param sl=0.02,0.03 \
  --sort sharpe \
  --top 5 \
  --db data/trading.db
```

### Run a Walk-Forward Validation

```bash
cd /Users/lavumi/private/crypto-quant

./server walk-forward \
  --strategy golden_rsi_bb \
  --symbol BTCUSDT \
  --interval 1h \
  --start 2025-01-01 \
  --end 2025-12-31 \
  --param golden_fast=5,10 \
  --param golden_slow=20,30 \
  --param tp=0.04,0.06 \
  --param sl=0.02,0.03 \
  --sort sharpe \
  --train-days 180 \
  --test-days 60 \
  --step-days 60 \
  --top 5 \
  --db data/trading.db
```

### Notes

- Collect data first, then run sweep or walk-forward on the same symbol and interval.
- `--sort sharpe` is a good default, but comparing `return`, `calmar`, and `profit_factor` is also important.
- Start with a small parameter grid first, then widen the search once the pipeline looks healthy.

## Success Criteria

The backtesting system should be considered materially improved when:

- backtests fail fast on low-quality data
- execution assumptions are explicit and configurable
- parameter sweep results are interpretable at a glance
- walk-forward output highlights out-of-sample robustness
- saved experiment runs can be queried and compared later
- strategy evaluation depends less on single-point return and more on robustness

## Notes

This document is intentionally biased toward research quality first.

For this project, the right order is:

- do not prioritize more strategies first
- do not prioritize live trading first
- improve confidence in the backtest pipeline first

That creates a much stronger foundation for every later stage, including paper trading, live execution, and UI analysis tools.
