package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/response"
	"github.com/lavumi/crypto-quant/internal/datasource/database"
)

type ExperimentsHandler struct {
	db *database.DB
}

func NewExperimentsHandler(db *database.DB) *ExperimentsHandler {
	return &ExperimentsHandler{db: db}
}

type ExperimentListItem struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	StrategyName string                 `json:"strategy_name"`
	Symbol       string                 `json:"symbol"`
	Interval     string                 `json:"interval"`
	CreatedAt    string                 `json:"created_at"`
	Summary      map[string]interface{} `json:"summary"`
}

type sweepListRow struct {
	ID               int64   `db:"id"`
	StrategyName     string  `db:"strategy_name"`
	Symbol           string  `db:"symbol"`
	Interval         string  `db:"interval"`
	CreatedAt        string  `db:"created_at"`
	SortBy           string  `db:"sort_by"`
	SuccessfulRuns   int     `db:"successful_runs"`
	FailedRuns       int     `db:"failed_runs"`
	TotalCandidates  int     `db:"total_candidates"`
	BestReturn       float64 `db:"best_return"`
	BestSharpe       float64 `db:"best_sharpe"`
	BestProfitFactor float64 `db:"best_profit_factor"`
	BestDrawdownPct  float64 `db:"best_drawdown_pct"`
}

type walkForwardListRow struct {
	ID               int64   `db:"id"`
	StrategyName     string  `db:"strategy_name"`
	Symbol           string  `db:"symbol"`
	Interval         string  `db:"interval"`
	CreatedAt        string  `db:"created_at"`
	SelectionMetric  string  `db:"selection_metric"`
	CompletedWindows int     `db:"completed_windows"`
	TotalWindows     int     `db:"total_windows"`
	AvgTestReturn    float64 `db:"avg_test_return"`
	AvgTestSharpe    float64 `db:"avg_test_sharpe"`
	PositiveRatio    float64 `db:"positive_ratio"`
}

type singleRunListRow struct {
	ID             int64   `db:"id"`
	StrategyName   string  `db:"strategy_name"`
	Symbol         string  `db:"symbol"`
	Interval       string  `db:"interval"`
	CreatedAt      string  `db:"created_at"`
	TotalReturn    float64 `db:"total_return"`
	SharpeRatio    float64 `db:"sharpe_ratio"`
	MaxDrawdownPct float64 `db:"max_drawdown_pct"`
	TotalTrades    int     `db:"total_trades"`
}

func (h *ExperimentsHandler) ListExperiments(c *gin.Context) {
	experimentType := c.Query("type")
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	items := make([]ExperimentListItem, 0)

	if experimentType == "" || experimentType == "single" {
		rows, err := h.listSingleRuns(limit)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to load single experiments: "+err.Error())
			return
		}
		for _, row := range rows {
			items = append(items, ExperimentListItem{
				ID:           "single_" + strconv.FormatInt(row.ID, 10),
				Type:         "single",
				StrategyName: row.StrategyName,
				Symbol:       row.Symbol,
				Interval:     row.Interval,
				CreatedAt:    row.CreatedAt,
				Summary: map[string]interface{}{
					"total_return":     row.TotalReturn,
					"sharpe_ratio":     row.SharpeRatio,
					"max_drawdown_pct": row.MaxDrawdownPct,
					"total_trades":     row.TotalTrades,
				},
			})
		}
	}

	if experimentType == "" || experimentType == "sweep" {
		rows, err := h.listSweeps(limit)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to load sweep experiments: "+err.Error())
			return
		}
		for _, row := range rows {
			items = append(items, ExperimentListItem{
				ID:           "sweep_" + strconv.FormatInt(row.ID, 10),
				Type:         "sweep",
				StrategyName: row.StrategyName,
				Symbol:       row.Symbol,
				Interval:     row.Interval,
				CreatedAt:    row.CreatedAt,
				Summary: map[string]interface{}{
					"sort_by":               row.SortBy,
					"total_candidates":      row.TotalCandidates,
					"successful_runs":       row.SuccessfulRuns,
					"failed_runs":           row.FailedRuns,
					"best_return":           row.BestReturn,
					"best_sharpe_ratio":     row.BestSharpe,
					"best_profit_factor":    row.BestProfitFactor,
					"best_max_drawdown_pct": row.BestDrawdownPct,
				},
			})
		}
	}

	if experimentType == "" || experimentType == "walk_forward" {
		rows, err := h.listWalkForwardRuns(limit)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to load walk-forward experiments: "+err.Error())
			return
		}
		for _, row := range rows {
			items = append(items, ExperimentListItem{
				ID:           "walk_forward_" + strconv.FormatInt(row.ID, 10),
				Type:         "walk_forward",
				StrategyName: row.StrategyName,
				Symbol:       row.Symbol,
				Interval:     row.Interval,
				CreatedAt:    row.CreatedAt,
				Summary: map[string]interface{}{
					"selection_metric":    row.SelectionMetric,
					"completed_windows":   row.CompletedWindows,
					"total_windows":       row.TotalWindows,
					"avg_test_return":     row.AvgTestReturn,
					"avg_test_sharpe":     row.AvgTestSharpe,
					"positive_test_ratio": row.PositiveRatio,
				},
			})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	if len(items) > limit {
		items = items[:limit]
	}

	response.Success(c, gin.H{"items": items})
}

func (h *ExperimentsHandler) GetSweep(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid sweep id")
		return
	}

	type sweepMeta struct {
		ID                int64  `db:"id"`
		StrategyName      string `db:"strategy_name"`
		Symbol            string `db:"symbol"`
		Interval          string `db:"interval"`
		SortBy            string `db:"sort_by"`
		BaseConfigJSON    string `db:"base_config_json"`
		ParameterGridJSON string `db:"parameter_grid_json"`
		TotalCandidates   int    `db:"total_candidates"`
		SuccessfulRuns    int    `db:"successful_runs"`
		FailedRuns        int    `db:"failed_runs"`
		CreatedAt         string `db:"created_at"`
	}

	var meta sweepMeta
	if err := h.db.Get(&meta, `
		SELECT id, strategy_name, symbol, interval, sort_by, base_config_json, parameter_grid_json,
		       total_candidates, successful_runs, failed_runs, CAST(created_at AS TEXT) AS created_at
		FROM backtest_sweeps WHERE id = ?
	`, id); err != nil {
		response.Error(c, http.StatusNotFound, "Sweep experiment not found")
		return
	}

	type sweepResultRow struct {
		Rank           int      `db:"rank"`
		ConfigJSON     string   `db:"config_json"`
		TotalReturn    *float64 `db:"total_return"`
		SharpeRatio    *float64 `db:"sharpe_ratio"`
		CalmarRatio    *float64 `db:"calmar_ratio"`
		MaxDrawdownPct *float64 `db:"max_drawdown_pct"`
		ProfitFactor   *float64 `db:"profit_factor"`
		WinRate        *float64 `db:"win_rate"`
		TotalTrades    *int     `db:"total_trades"`
		ErrorMessage   *string  `db:"error_message"`
	}

	var rows []sweepResultRow
	if err := h.db.Select(&rows, `
		SELECT rank, config_json, total_return, sharpe_ratio, calmar_ratio,
		       max_drawdown_pct, profit_factor, win_rate, total_trades, error_message
		FROM backtest_sweep_results
		WHERE sweep_id = ?
		ORDER BY rank ASC
	`, id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load sweep results: "+err.Error())
		return
	}

	results := make([]map[string]interface{}, 0, len(rows))
	var best map[string]interface{}
	var median map[string]interface{}
	var worst map[string]interface{}

	for _, row := range rows {
		item := map[string]interface{}{
			"rank":             row.Rank,
			"config":           decodeJSON(row.ConfigJSON),
			"config_json":      row.ConfigJSON,
			"total_return":     row.TotalReturn,
			"sharpe_ratio":     row.SharpeRatio,
			"calmar_ratio":     row.CalmarRatio,
			"max_drawdown_pct": row.MaxDrawdownPct,
			"profit_factor":    row.ProfitFactor,
			"win_rate":         row.WinRate,
			"total_trades":     row.TotalTrades,
			"error_message":    row.ErrorMessage,
		}
		results = append(results, item)
	}

	successful := successfulMaps(results)
	if len(successful) > 0 {
		best = successful[0]
		median = successful[len(successful)/2]
		worst = successful[len(successful)-1]
	}

	response.Success(c, gin.H{
		"id":             "sweep_" + strconv.FormatInt(meta.ID, 10),
		"type":           "sweep",
		"strategy_name":  meta.StrategyName,
		"symbol":         meta.Symbol,
		"interval":       meta.Interval,
		"created_at":     meta.CreatedAt,
		"sort_by":        meta.SortBy,
		"base_config":    decodeJSON(meta.BaseConfigJSON),
		"parameter_grid": decodeJSON(meta.ParameterGridJSON),
		"summary": gin.H{
			"total_candidates": meta.TotalCandidates,
			"successful_runs":  meta.SuccessfulRuns,
			"failed_runs":      meta.FailedRuns,
			"best":             best,
			"median":           median,
			"worst":            worst,
		},
		"results": results,
	})
}

func (h *ExperimentsHandler) GetSingle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid single experiment id")
		return
	}

	type singleMeta struct {
		ID             int64   `db:"id"`
		StrategyName   string  `db:"strategy_name"`
		Symbol         string  `db:"symbol"`
		Interval       string  `db:"interval"`
		StartTime      int64   `db:"start_time"`
		EndTime        int64   `db:"end_time"`
		InitialBalance float64 `db:"initial_balance"`
		FinalEquity    float64 `db:"final_equity"`
		TotalReturn    float64 `db:"total_return"`
		SharpeRatio    float64 `db:"sharpe_ratio"`
		MaxDrawdown    float64 `db:"max_drawdown"`
		MaxDrawdownPct float64 `db:"max_drawdown_pct"`
		WinRate        float64 `db:"win_rate"`
		TotalTrades    int     `db:"total_trades"`
		Commission     float64 `db:"commission"`
		ConfigJSON     string  `db:"config_json"`
		CreatedAt      string  `db:"created_at"`
	}

	var meta singleMeta
	if err := h.db.Get(&meta, `
		SELECT id, strategy_name, symbol, interval, start_time, end_time,
		       initial_balance, final_equity, total_return, sharpe_ratio,
		       max_drawdown, max_drawdown_pct, win_rate, total_trades, commission,
		       config_json, CAST(created_at AS TEXT) AS created_at
		FROM backtest_runs
		WHERE id = ?
	`, id); err != nil {
		response.Error(c, http.StatusNotFound, "Single experiment not found")
		return
	}

	type tradeRow struct {
		Timestamp int64   `db:"timestamp"`
		Side      string  `db:"side"`
		Price     float64 `db:"price"`
		Quantity  float64 `db:"quantity"`
		Fee       float64 `db:"fee"`
		Balance   float64 `db:"balance"`
		Position  float64 `db:"position"`
		Reason    string  `db:"reason"`
	}

	type equityRow struct {
		Timestamp int64   `db:"timestamp"`
		Equity    float64 `db:"equity"`
		Price     float64 `db:"price"`
	}

	var trades []tradeRow
	if err := h.db.Select(&trades, `
		SELECT timestamp, side, price, quantity, fee, balance, position, reason
		FROM backtest_run_trades
		WHERE run_id = ?
		ORDER BY timestamp ASC
	`, id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load single experiment trades: "+err.Error())
		return
	}

	var equity []equityRow
	if err := h.db.Select(&equity, `
		SELECT timestamp, equity, price
		FROM backtest_run_equity
		WHERE run_id = ?
		ORDER BY timestamp ASC
	`, id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load single experiment equity: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":            "single_" + strconv.FormatInt(meta.ID, 10),
		"type":          "single",
		"strategy_name": meta.StrategyName,
		"symbol":        meta.Symbol,
		"interval":      meta.Interval,
		"created_at":    meta.CreatedAt,
		"config":        decodeJSON(meta.ConfigJSON),
		"summary": gin.H{
			"start_time":       meta.StartTime,
			"end_time":         meta.EndTime,
			"initial_balance":  meta.InitialBalance,
			"final_equity":     meta.FinalEquity,
			"total_return":     meta.TotalReturn,
			"sharpe_ratio":     meta.SharpeRatio,
			"max_drawdown":     meta.MaxDrawdown,
			"max_drawdown_pct": meta.MaxDrawdownPct,
			"win_rate":         meta.WinRate,
			"total_trades":     meta.TotalTrades,
			"commission":       meta.Commission,
		},
		"equity_curve": equity,
		"trades":       trades,
	})
}

func (h *ExperimentsHandler) GetWalkForward(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid walk-forward id")
		return
	}

	type wfMeta struct {
		ID                int64  `db:"id"`
		StrategyName      string `db:"strategy_name"`
		Symbol            string `db:"symbol"`
		Interval          string `db:"interval"`
		SelectionMetric   string `db:"selection_metric"`
		BaseConfigJSON    string `db:"base_config_json"`
		ParameterGridJSON string `db:"parameter_grid_json"`
		TrainDurationSec  int64  `db:"train_duration_seconds"`
		TestDurationSec   int64  `db:"test_duration_seconds"`
		StepDurationSec   int64  `db:"step_duration_seconds"`
		TotalWindows      int    `db:"total_windows"`
		CompletedWindows  int    `db:"completed_windows"`
		CreatedAt         string `db:"created_at"`
	}

	var meta wfMeta
	if err := h.db.Get(&meta, `
		SELECT id, strategy_name, symbol, interval, selection_metric, base_config_json, parameter_grid_json,
		       train_duration_seconds, test_duration_seconds, step_duration_seconds,
		       total_windows, completed_windows, CAST(created_at AS TEXT) AS created_at
		FROM walk_forward_runs WHERE id = ?
	`, id); err != nil {
		response.Error(c, http.StatusNotFound, "Walk-forward experiment not found")
		return
	}

	type wfWindowRow struct {
		WindowIndex        int      `db:"window_index"`
		TrainStart         int64    `db:"train_start"`
		TrainEnd           int64    `db:"train_end"`
		TestStart          int64    `db:"test_start"`
		TestEnd            int64    `db:"test_end"`
		SelectedConfigJSON *string  `db:"selected_config_json"`
		TrainSharpe        *float64 `db:"train_sharpe"`
		TestReturn         *float64 `db:"test_return"`
		TestSharpe         *float64 `db:"test_sharpe"`
		TestMDDPct         *float64 `db:"test_max_drawdown_pct"`
		ErrorMessage       *string  `db:"error_message"`
	}

	var rows []wfWindowRow
	if err := h.db.Select(&rows, `
		SELECT window_index, train_start, train_end, test_start, test_end,
		       selected_config_json, train_sharpe, test_return, test_sharpe, test_max_drawdown_pct, error_message
		FROM walk_forward_windows
		WHERE walk_forward_id = ?
		ORDER BY window_index ASC
	`, id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load walk-forward windows: "+err.Error())
		return
	}

	windows := make([]map[string]interface{}, 0, len(rows))
	var avgTestReturn float64
	var avgTestSharpe float64
	var positiveCount int
	var validCount int
	for _, row := range rows {
		window := map[string]interface{}{
			"window_index":          row.WindowIndex,
			"train_start":           row.TrainStart,
			"train_end":             row.TrainEnd,
			"test_start":            row.TestStart,
			"test_end":              row.TestEnd,
			"selected_config":       decodeJSONStringPtr(row.SelectedConfigJSON),
			"train_sharpe":          row.TrainSharpe,
			"test_return":           row.TestReturn,
			"test_sharpe":           row.TestSharpe,
			"test_max_drawdown_pct": row.TestMDDPct,
			"error_message":         row.ErrorMessage,
		}
		windows = append(windows, window)

		if row.TestReturn != nil {
			avgTestReturn += *row.TestReturn
			validCount++
			if *row.TestReturn > 0 {
				positiveCount++
			}
		}
		if row.TestSharpe != nil {
			avgTestSharpe += *row.TestSharpe
		}
	}

	var positiveRatio float64
	if validCount > 0 {
		avgTestReturn /= float64(validCount)
		avgTestSharpe /= float64(validCount)
		positiveRatio = float64(positiveCount) / float64(validCount)
	}

	response.Success(c, gin.H{
		"id":               "walk_forward_" + strconv.FormatInt(meta.ID, 10),
		"type":             "walk_forward",
		"strategy_name":    meta.StrategyName,
		"symbol":           meta.Symbol,
		"interval":         meta.Interval,
		"created_at":       meta.CreatedAt,
		"selection_metric": meta.SelectionMetric,
		"base_config":      decodeJSON(meta.BaseConfigJSON),
		"parameter_grid":   decodeJSON(meta.ParameterGridJSON),
		"window_config": gin.H{
			"train_duration_seconds": meta.TrainDurationSec,
			"test_duration_seconds":  meta.TestDurationSec,
			"step_duration_seconds":  meta.StepDurationSec,
		},
		"summary": gin.H{
			"total_windows":       meta.TotalWindows,
			"completed_windows":   meta.CompletedWindows,
			"avg_test_return":     avgTestReturn,
			"avg_test_sharpe":     avgTestSharpe,
			"positive_test_ratio": positiveRatio,
		},
		"windows": windows,
	})
}

func (h *ExperimentsHandler) listSingleRuns(limit int) ([]singleRunListRow, error) {
	var rows []singleRunListRow
	err := h.db.Select(&rows, `
		SELECT id, strategy_name, symbol, interval,
		       CAST(created_at AS TEXT) AS created_at,
		       total_return, sharpe_ratio, max_drawdown_pct, total_trades
		FROM backtest_runs
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	return rows, err
}

func (h *ExperimentsHandler) listSweeps(limit int) ([]sweepListRow, error) {
	var rows []sweepListRow
	err := h.db.Select(&rows, `
		SELECT s.id, s.strategy_name, s.symbol, s.interval, CAST(s.created_at AS TEXT) AS created_at,
		       s.sort_by, s.successful_runs, s.failed_runs, s.total_candidates,
		       COALESCE(MIN(r.total_return), 0) AS best_return,
		       COALESCE(MAX(r.sharpe_ratio), 0) AS best_sharpe,
		       COALESCE(MAX(r.profit_factor), 0) AS best_profit_factor,
		       COALESCE(MIN(r.max_drawdown_pct), 0) AS best_drawdown_pct
		FROM backtest_sweeps s
		LEFT JOIN backtest_sweep_results r ON r.sweep_id = s.id AND r.rank = 1
		GROUP BY s.id
		ORDER BY s.created_at DESC
		LIMIT ?
	`, limit)
	return rows, err
}

func (h *ExperimentsHandler) listWalkForwardRuns(limit int) ([]walkForwardListRow, error) {
	var rows []walkForwardListRow
	err := h.db.Select(&rows, `
		SELECT w.id, w.strategy_name, w.symbol, w.interval, CAST(w.created_at AS TEXT) AS created_at,
		       w.selection_metric, w.completed_windows, w.total_windows,
		       COALESCE(AVG(ww.test_return), 0) AS avg_test_return,
		       COALESCE(AVG(ww.test_sharpe), 0) AS avg_test_sharpe,
		       COALESCE(AVG(CASE WHEN ww.test_return > 0 THEN 1.0 ELSE 0.0 END), 0) AS positive_ratio
		FROM walk_forward_runs w
		LEFT JOIN walk_forward_windows ww ON ww.walk_forward_id = w.id
		GROUP BY w.id
		ORDER BY w.created_at DESC
		LIMIT ?
	`, limit)
	return rows, err
}

func decodeJSON(raw string) interface{} {
	if raw == "" {
		return nil
	}
	var out interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return raw
	}
	return out
}

func decodeJSONStringPtr(raw *string) interface{} {
	if raw == nil {
		return nil
	}
	return decodeJSON(*raw)
}

func successfulMaps(items []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if item["error_message"] == nil {
			out = append(out, item)
		}
	}
	return out
}
