# Experiments Platform Redesign

## Why This Redesign Exists

The current project has evolved from a simple backtest form and result screen into a broader research platform with:

- single backtest runs
- parameter sweep
- walk-forward validation
- persisted experiment data

The problem is that the API and frontend still reflect the older "execute one backtest and show one result page" model.

That creates several issues:

- execution and retrieval concerns are mixed together
- sweep and walk-forward are not first-class product concepts
- persisted experiment data exists in the backend but not in a coherent UI model
- the frontend information architecture is too narrow for a research workflow
- old route assumptions are starting to fight the new domain model

This document proposes a new structure for the product around a single concept:

`experiments`

An experiment is a stored research artifact that can be inspected, compared, and revisited later.

## Redesign Principles

### 1. Readonly-first API

The API should primarily expose stored research results.

That means:

- querying experiments
- viewing summaries
- viewing detailed metrics
- viewing rankings
- viewing walk-forward windows
- viewing data quality state

Execution should not be the center of the web API.

Execution can remain in:

- CLI
- worker processes
- future internal admin endpoints if needed

### 2. Experiments as First-Class Resources

The frontend should think in terms of experiment collections and experiment details, not raw backtest handlers.

The main experiment types are:

- single backtest experiments
- sweep experiments
- walk-forward experiments

### 3. Stable Resource Shapes

Frontend pages should not need to know internal engine details.

Each resource should expose:

- metadata
- config
- key metrics
- detail payloads
- links to related resources

### 4. Preserve Domain Engine, Rebuild Product Surface

The engine, strategy, optimize, and persistence layers should mostly be kept.

The parts that should be substantially reworked are:

- API contracts
- route structure
- page structure
- frontend state model

## Domain Model

### Core Terms

#### Experiment

A stored research job. It is the top-level concept presented in the UI.

Fields:

- `id`
- `type`: `single`, `sweep`, `walk_forward`
- `strategy_name`
- `symbol`
- `interval`
- `created_at`
- `status`
- `summary`

#### Single Experiment

A stored backtest run based on one config.

Backed by:

- `backtest_runs`
- related trade/equity tables

#### Sweep Experiment

A stored parameter search job over one strategy and parameter grid.

Backed by:

- `backtest_sweeps`
- `backtest_sweep_results`

#### Walk-Forward Experiment

A stored rolling train/test validation job.

Backed by:

- `walk_forward_runs`
- `walk_forward_windows`

## API Direction

The new API should be centered around readonly resources.

Recommended top-level grouping:

- `/api/v2/experiments`
- `/api/v2/catalog`
- `/api/v2/data`

## API Resource Model

### 1. Experiments List

#### `GET /api/v2/experiments`

Unified list across all experiment types.

Query parameters:

- `type=single|sweep|walk_forward`
- `strategy=...`
- `symbol=...`
- `interval=...`
- `limit=...`
- `cursor=...`

Response shape:

```json
{
  "items": [
    {
      "id": "wf_42",
      "type": "walk_forward",
      "strategy_name": "golden_rsi_bb",
      "symbol": "BTCUSDT",
      "interval": "1h",
      "created_at": "2026-04-01T17:20:00Z",
      "summary": {
        "selection_metric": "sharpe",
        "completed_windows": 6,
        "avg_oos_return": 0.021,
        "avg_oos_sharpe": 0.44,
        "positive_window_ratio": 0.67
      }
    }
  ],
  "next_cursor": null
}
```

### 2. Single Experiment Detail

#### `GET /api/v2/experiments/single/:id`

Response sections:

- metadata
- config
- metrics
- equity curve
- drawdown curve
- monthly returns
- trades

### 3. Sweep Experiment Detail

#### `GET /api/v2/experiments/sweeps/:id`

Response sections:

- metadata
- base config
- parameter grid
- summary
- best / median / worst
- ranked candidates

#### `GET /api/v2/experiments/sweeps/:id/results`

Optional paginated table of ranked candidates.

### 4. Walk-Forward Experiment Detail

#### `GET /api/v2/experiments/walk-forward/:id`

Response sections:

- metadata
- base config
- parameter grid
- summary
- aggregate out-of-sample metrics
- window table

Each window should expose:

- train range
- test range
- selected params
- train metric
- test return
- test sharpe
- test drawdown

### 5. Catalog Endpoints

#### `GET /api/v2/catalog/strategies`

Returns strategy names, descriptions, and parameter schemas.

#### `GET /api/v2/catalog/metrics`

Returns metric definitions shown in the UI.

This is useful for tooltips and future consistency.

### 6. Data Endpoints

#### `GET /api/v2/data/coverage`

Returns:

- available symbols
- intervals
- earliest and latest data
- completeness summary

#### `GET /api/v2/data/validation`

Returns data quality report for a requested symbol/interval/range.

## Frontend Information Architecture

The frontend should no longer be organized around "new backtest form" and "one result screen".

Recommended route map:

- `/`
- `/experiments`
- `/experiments/single`
- `/experiments/sweeps`
- `/experiments/walk-forward`
- `/experiments/single/[id]`
- `/experiments/sweeps/[id]`
- `/experiments/walk-forward/[id]`
- `/catalog/strategies`
- `/data/coverage`

## Frontend Page Roles

### Home

Purpose:

- recent experiments
- quick health and data coverage summary
- links into main research flows

### Experiments Index

Purpose:

- unified research inbox
- filter by type, strategy, symbol, interval

Main columns:

- type
- strategy
- symbol
- interval
- created at
- headline metric

### Single Experiment Detail

Purpose:

- inspect one saved run deeply

Main sections:

- headline metrics
- equity chart
- drawdown chart
- monthly returns
- trade table
- config panel

### Sweep Detail

Purpose:

- inspect the quality of a parameter search

Main sections:

- experiment metadata
- summary stats
- best / median / worst cards
- ranked candidate table
- later: heatmap / scatter / distribution charts

### Walk-Forward Detail

Purpose:

- inspect out-of-sample robustness

Main sections:

- aggregate summary
- positive window ratio
- average OOS return
- average OOS sharpe
- window-by-window table
- later: rolling timeline chart

### Strategy Catalog

Purpose:

- show what the platform can evaluate
- document parameters and defaults

### Data Coverage

Purpose:

- show whether requested ranges are trustworthy
- reduce wasted backtests on bad data

## Component Model

Recommended frontend components:

- `ExperimentTable`
- `MetricGrid`
- `ConfigPanel`
- `EquityChart`
- `DrawdownChart`
- `MonthlyReturnsTable`
- `SweepCandidateTable`
- `WalkForwardWindowTable`
- `DataCoverageCard`

The important shift is:

- move away from page-specific ad hoc rendering
- move toward reusable research views

## Backend Migration Strategy

This should not be done as a single destructive rewrite.

Recommended phases:

### Phase 1: New Readonly API

Build `/api/v2` in parallel with the old API.

Do not delete current endpoints yet.

Implement first:

- experiments list
- sweep detail
- walk-forward detail
- strategy catalog

### Phase 2: New Frontend Routes

Add new pages alongside the old ones:

- `/experiments`
- `/experiments/sweeps/[id]`
- `/experiments/walk-forward/[id]`

Do not migrate the old `/backtest/new` flow yet.

### Phase 3: Reframe Old Backtest UI

Decide whether to:

- keep a lightweight execution UI as an admin/research tool
- or remove it from the public product surface

If execution stays, it should likely move under:

- `/lab/backtest`

That keeps the main product centered on experiment review.

### Phase 4: Decommission Legacy Routes

Once the new experiment pages are stable:

- remove old result-specific coupling
- archive or delete obsolete API docs and route assumptions

## Recommended First Implementation Slice

The best first slice is:

1. `GET /api/v2/experiments`
2. `GET /api/v2/experiments/sweeps/:id`
3. `GET /api/v2/experiments/walk-forward/:id`
4. frontend `/experiments` list page

Why this first:

- the data already exists in the database
- it proves the new product model
- it avoids getting blocked by execution UX questions
- it reduces pressure to keep old API patterns alive

## What Should Be Left Behind

The redesign should intentionally leave behind these assumptions:

- "the product is a backtest form"
- "every result is just one backtest response payload"
- "execution endpoints are the center of the UI"
- "sweep and walk-forward are advanced add-ons"

Those assumptions no longer match the direction of the system.

## Final Position

This repository should evolve into an experiments platform, not a one-shot backtest form.

That means:

- the backend should expose readonly experiment resources
- the frontend should organize around experiment discovery and inspection
- execution should move out of the product center

This is not a small refactor.

It is a product-surface redesign built on top of the existing engine and persistence work.
