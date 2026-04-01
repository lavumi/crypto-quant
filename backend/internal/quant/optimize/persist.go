package optimize

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lavumi/crypto-quant/internal/datasource/database"
)

// PersistSweep stores a completed sweep job and its ranked candidates.
func PersistSweep(ctx context.Context, db *database.DB, runnerCfg RunnerConfig, spec SweepSpec, results []SweepResult) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("nil database")
	}

	summary := Summarize(results, spec.SortBy)
	baseConfigJSON, err := json.Marshal(spec.BaseConfig.Normalize())
	if err != nil {
		return 0, fmt.Errorf("marshal base config: %w", err)
	}
	gridJSON, err := json.Marshal(spec.Parameters)
	if err != nil {
		return 0, fmt.Errorf("marshal parameter grid: %w", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO backtest_sweeps (
			strategy_name, symbol, interval, sort_by,
			base_config_json, parameter_grid_json,
			total_candidates, successful_runs, failed_runs
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		spec.BaseConfig.Normalize().Name,
		runnerCfg.Symbol,
		runnerCfg.Interval,
		summary.SortBy,
		string(baseConfigJSON),
		string(gridJSON),
		summary.TotalCandidates,
		summary.SuccessfulRuns,
		summary.FailedRuns,
	)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("insert sweep: %w", err)
	}

	sweepID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO backtest_sweep_results (
			sweep_id, rank, config_json, total_return, sharpe_ratio, calmar_ratio,
			max_drawdown_pct, profit_factor, win_rate, total_trades, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("prepare sweep results: %w", err)
	}
	defer stmt.Close()

	for idx, item := range results {
		var totalReturn, sharpe, calmar, mddPct, profitFactor, winRate any
		var totalTrades any
		var errorMessage any

		if item.Result != nil {
			totalReturn = item.Result.TotalReturn
			sharpe = item.Result.SharpeRatio
			calmar = item.Result.CalmarRatio
			mddPct = item.Result.MaxDrawdownPct
			profitFactor = item.Result.ProfitFactor
			winRate = item.Result.WinRate
			totalTrades = item.Result.TotalTrades
		}
		if item.Err != nil {
			errorMessage = item.Err.Error()
		}

		if _, err := stmt.ExecContext(ctx,
			sweepID,
			idx+1,
			item.ConfigJSON,
			totalReturn,
			sharpe,
			calmar,
			mddPct,
			profitFactor,
			winRate,
			totalTrades,
			errorMessage,
		); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("insert sweep result: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit sweep persistence: %w", err)
	}

	return sweepID, nil
}

// PersistWalkForward stores a walk-forward run and per-window out-of-sample results.
func PersistWalkForward(ctx context.Context, db *database.DB, runnerCfg RunnerConfig, spec WalkForwardSpec, summary *WalkForwardSummary) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("nil database")
	}
	if summary == nil {
		return 0, fmt.Errorf("nil walk-forward summary")
	}

	baseConfigJSON, err := json.Marshal(spec.SweepSpec.BaseConfig.Normalize())
	if err != nil {
		return 0, fmt.Errorf("marshal base config: %w", err)
	}
	gridJSON, err := json.Marshal(spec.SweepSpec.Parameters)
	if err != nil {
		return 0, fmt.Errorf("marshal parameter grid: %w", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO walk_forward_runs (
			strategy_name, symbol, interval, selection_metric,
			base_config_json, parameter_grid_json,
			train_duration_seconds, test_duration_seconds, step_duration_seconds,
			total_windows, completed_windows
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		spec.SweepSpec.BaseConfig.Normalize().Name,
		runnerCfg.Symbol,
		runnerCfg.Interval,
		summary.SelectionMetric,
		string(baseConfigJSON),
		string(gridJSON),
		int64(spec.TrainDuration.Seconds()),
		int64(spec.TestDuration.Seconds()),
		int64(spec.StepDuration.Seconds()),
		len(summary.Windows),
		summary.Completed,
	)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("insert walk-forward run: %w", err)
	}

	runID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO walk_forward_windows (
			walk_forward_id, window_index, train_start, train_end, test_start, test_end,
			selected_config_json, train_sharpe, test_return, test_sharpe, test_max_drawdown_pct, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("prepare walk-forward windows: %w", err)
	}
	defer stmt.Close()

	for idx, window := range summary.Windows {
		var selectedConfig any
		var trainSharpe, testReturn, testSharpe, testMDD any
		var errorMessage any

		if window.Selected != nil {
			selectedConfig = window.Selected.ConfigJSON
			if window.Selected.Result != nil {
				trainSharpe = window.Selected.Result.SharpeRatio
			}
		}
		if window.OutOfSample != nil {
			if window.OutOfSample.Result != nil {
				testReturn = window.OutOfSample.Result.TotalReturn
				testSharpe = window.OutOfSample.Result.SharpeRatio
				testMDD = window.OutOfSample.Result.MaxDrawdownPct
			}
			if window.OutOfSample.Err != nil {
				errorMessage = window.OutOfSample.Err.Error()
			}
		}

		if _, err := stmt.ExecContext(ctx,
			runID,
			idx+1,
			window.Window.TrainStart.Unix(),
			window.Window.TrainEnd.Unix(),
			window.Window.TestStart.Unix(),
			window.Window.TestEnd.Unix(),
			selectedConfig,
			trainSharpe,
			testReturn,
			testSharpe,
			testMDD,
			errorMessage,
		); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("insert walk-forward window: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit walk-forward persistence: %w", err)
	}

	return runID, nil
}
