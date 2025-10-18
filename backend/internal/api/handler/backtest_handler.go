package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/response"
	"github.com/lavumi/crypto-quant/internal/datasource/market/history"
	"github.com/lavumi/crypto-quant/internal/quant/backtest"
	"github.com/lavumi/crypto-quant/internal/quant/strategy"
)

// BacktestHandler handles backtest API requests
type BacktestHandler struct {
	historyService *history.Service
}

// NewBacktestHandler creates a new backtest handler
func NewBacktestHandler(historyService *history.Service) *BacktestHandler {
	return &BacktestHandler{
		historyService: historyService,
	}
}

// BacktestRequest represents a backtest request
type BacktestRequest struct {
	Symbol         string  `json:"symbol" binding:"required" example:"BTCUSDT"`
	Interval       string  `json:"interval" binding:"required" example:"1h"`
	StartDate      string  `json:"start_date" binding:"required" example:"2025-07-01"`
	EndDate        string  `json:"end_date" binding:"required" example:"2025-10-17"`
	InitialBalance float64 `json:"initial_balance" example:"10000"`
	Commission     float64 `json:"commission" example:"0.001"`
	Strategy       string  `json:"strategy" example:"ma_cross"`

	// Strategy parameters (MA Cross)
	FastPeriod int `json:"fast_period" example:"10"`
	SlowPeriod int `json:"slow_period" example:"30"`

	// Strategy parameters (RSI)
	RSIPeriod     int     `json:"rsi_period" example:"14"`
	RSIOversold   float64 `json:"rsi_oversold" example:"30"`
	RSIOverbought float64 `json:"rsi_overbought" example:"70"`

	// Position size
	PositionSize float64 `json:"position_size" example:"0.01"`
}

// BacktestResponse represents a backtest response
type BacktestResponse struct {
	// Time metrics
	StartTime string `json:"start_time" example:"2025-07-19T17:00:00Z"`
	EndTime   string `json:"end_time" example:"2025-10-17T08:00:00Z"`
	Duration  string `json:"duration" example:"2151h0m0s"`

	// Financial metrics
	InitialBalance float64 `json:"initial_balance" example:"10000.00"`
	FinalEquity    float64 `json:"final_equity" example:"9866.65"`
	TotalReturn    float64 `json:"total_return" example:"-1.33"`
	TotalReturnPct string  `json:"total_return_pct" example:"-1.33%"`

	// Risk metrics
	SharpeRatio    float64 `json:"sharpe_ratio" example:"-2.16"`
	MaxDrawdown    float64 `json:"max_drawdown" example:"206.59"`
	MaxDrawdownPct float64 `json:"max_drawdown_pct" example:"2.07"`

	// Trade statistics
	TotalTrades   int     `json:"total_trades" example:"82"`
	WinningTrades int     `json:"winning_trades" example:"12"`
	LosingTrades  int     `json:"losing_trades" example:"29"`
	WinRate       float64 `json:"win_rate" example:"29.27"`
	WinRatePct    string  `json:"win_rate_pct" example:"29.27%"`

	// Configuration
	Strategy    string  `json:"strategy" example:"MA_Cross_10_30"`
	Symbol      string  `json:"symbol" example:"BTCUSDT"`
	Interval    string  `json:"interval" example:"1h"`
	Commission  float64 `json:"commission" example:"0.001"`
	CandlesUsed int     `json:"candles_used" example:"2152"`

	// Trade details (optional, limited to first 20)
	RecentTrades []TradeDetail `json:"recent_trades,omitempty"`

	// Chart data for visualization
	ChartData *ChartData `json:"chart_data,omitempty"`
}

// ChartData contains data for chart visualization
type ChartData struct {
	// Equity curve points (time series of portfolio value)
	EquityCurve []EquityChartPoint `json:"equity_curve"`

	// Trade markers for buy/sell points on the chart
	Trades []TradeChartPoint `json:"trades"`

	// Strategy-specific indicator data (e.g., moving averages for MA Cross)
	Indicators *IndicatorData `json:"indicators,omitempty"`
}

// IndicatorData contains indicator values for charting
type IndicatorData struct {
	// Price data (for reference)
	PriceData []PricePoint `json:"price_data,omitempty"`

	// Moving averages (for MA Cross strategy)
	FastMA []IndicatorPoint `json:"fast_ma,omitempty"`
	SlowMA []IndicatorPoint `json:"slow_ma,omitempty"`

	// RSI values (for RSI strategy)
	RSI []IndicatorPoint `json:"rsi,omitempty"`
}

// PricePoint represents a price point
type PricePoint struct {
	Timestamp string  `json:"timestamp" example:"2025-07-19T17:00:00Z"`
	Price     float64 `json:"price" example:"60000.00"`
}

// IndicatorPoint represents an indicator value at a point in time
type IndicatorPoint struct {
	Timestamp string  `json:"timestamp" example:"2025-07-19T17:00:00Z"`
	Value     float64 `json:"value" example:"59850.50"`
}

// EquityChartPoint represents a point on the equity curve
type EquityChartPoint struct {
	Timestamp string  `json:"timestamp" example:"2025-07-19T17:00:00Z"`
	Equity    float64 `json:"equity" example:"10000.50"`
	Price     float64 `json:"price" example:"60000.00"`
}

// TradeChartPoint represents a trade point on the chart
type TradeChartPoint struct {
	Timestamp string  `json:"timestamp" example:"2025-07-21T16:00:00Z"`
	Side      string  `json:"side" example:"BUY"`
	Price     float64 `json:"price" example:"119362.67"`
	Equity    float64 `json:"equity" example:"10100.00"`
}

// TradeDetail represents a single trade
type TradeDetail struct {
	Timestamp string  `json:"timestamp" example:"2025-07-21T16:00:00Z"`
	Side      string  `json:"side" example:"BUY"`
	Price     float64 `json:"price" example:"119362.67"`
	Quantity  float64 `json:"quantity" example:"0.01"`
	Fee       float64 `json:"fee" example:"1.19"`
	Balance   float64 `json:"balance" example:"9328.59"`
	Position  float64 `json:"position" example:"0.01"`
	Reason    string  `json:"reason" example:"Golden Cross: Fast MA (118128.46) > Slow MA (118117.20)"`
}

// RunBacktest godoc
// @Summary Run a backtest
// @Description Execute a backtest with specified parameters and strategy
// @Tags backtest
// @Accept json
// @Produce json
// @Param request body BacktestRequest true "Backtest parameters"
// @Success 200 {object} BacktestResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /backtest/run [post]
func (h *BacktestHandler) RunBacktest(c *gin.Context) {
	var req BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Set defaults
	if req.InitialBalance == 0 {
		req.InitialBalance = 10000
	}
	if req.Commission == 0 {
		req.Commission = 0.001
	}
	if req.Strategy == "" {
		req.Strategy = "ma_cross"
	}
	if req.PositionSize == 0 {
		req.PositionSize = 0.01
	}

	// Parse dates in UTC
	startTime, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}
	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)

	endTime, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)

	// Load candles
	ctx := context.Background()
	candles, err := h.historyService.GetCandles(ctx, req.Symbol, req.Interval, startTime, endTime)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load candles: "+err.Error())
		return
	}

	if len(candles) == 0 {
		response.Error(c, http.StatusNotFound, "No candles found for the specified period. Please collect data first.")
		return
	}

	// Create strategy
	var strat backtest.Strategy
	switch req.Strategy {
	case "ma_cross":
		fastPeriod := req.FastPeriod
		slowPeriod := req.SlowPeriod
		if fastPeriod == 0 {
			fastPeriod = 10
		}
		if slowPeriod == 0 {
			slowPeriod = 30
		}
		strat = strategy.NewMACrossStrategy(fastPeriod, slowPeriod)

	case "rsi":
		period := req.RSIPeriod
		oversold := req.RSIOversold
		overbought := req.RSIOverbought
		if period == 0 {
			period = 14
		}
		if oversold == 0 {
			oversold = 30
		}
		if overbought == 0 {
			overbought = 70
		}
		strat = strategy.NewRSIStrategy(period, oversold, overbought, req.PositionSize)

	case "bb_rsi":
		bbPeriod := 20
		bbMultiplier := 2.0
		rsiPeriod := req.RSIPeriod
		rsiOversold := req.RSIOversold
		rsiOverbought := req.RSIOverbought
		if rsiPeriod == 0 {
			rsiPeriod = 14
		}
		if rsiOversold == 0 {
			rsiOversold = 30
		}
		if rsiOverbought == 0 {
			rsiOverbought = 70
		}
		strat = strategy.NewBBRSIStrategy(bbPeriod, bbMultiplier, rsiPeriod, rsiOversold, rsiOverbought, req.PositionSize)

	default:
		response.Error(c, http.StatusBadRequest, "Unknown strategy: "+req.Strategy+". Available: ma_cross, rsi, bb_rsi")
		return
	}

	// Create and run backtest engine
	engine := backtest.NewEngine(&backtest.Config{
		InitialBalance: req.InitialBalance,
		Commission:     req.Commission,
		Strategy:       strat,
	})

	result, err := engine.Run(ctx, candles)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Backtest failed: "+err.Error())
		return
	}

	// Convert trades to response format (limit to 20 most recent)
	recentTrades := make([]TradeDetail, 0)
	maxTrades := 20
	if len(result.Trades) > maxTrades {
		for _, trade := range result.Trades[:maxTrades] {
			recentTrades = append(recentTrades, TradeDetail{
				Timestamp: trade.Timestamp.Format(time.RFC3339),
				Side:      string(trade.Side),
				Price:     trade.Price,
				Quantity:  trade.Quantity,
				Fee:       trade.Fee,
				Balance:   trade.Balance,
				Position:  trade.Position,
				Reason:    trade.Reason,
			})
		}
	} else {
		for _, trade := range result.Trades {
			recentTrades = append(recentTrades, TradeDetail{
				Timestamp: trade.Timestamp.Format(time.RFC3339),
				Side:      string(trade.Side),
				Price:     trade.Price,
				Quantity:  trade.Quantity,
				Fee:       trade.Fee,
				Balance:   trade.Balance,
				Position:  trade.Position,
				Reason:    trade.Reason,
			})
		}
	}

	// Build chart data with strategy-specific indicators
	chartData := buildChartData(result, strat)

	// Build response
	resp := BacktestResponse{
		StartTime:      result.StartTime.Format(time.RFC3339),
		EndTime:        result.EndTime.Format(time.RFC3339),
		Duration:       result.Duration.String(),
		InitialBalance: result.InitialBalance,
		FinalEquity:    result.FinalEquity,
		TotalReturn:    result.TotalReturn * 100,
		TotalReturnPct: formatPercent(result.TotalReturn),
		SharpeRatio:    result.SharpeRatio,
		MaxDrawdown:    result.MaxDrawdown,
		MaxDrawdownPct: result.MaxDrawdownPct * 100,
		TotalTrades:    result.TotalTrades,
		WinningTrades:  result.WinningTrades,
		LosingTrades:   result.LosingTrades,
		WinRate:        result.WinRate * 100,
		WinRatePct:     formatPercent(result.WinRate),
		Strategy:       strat.Name(),
		Symbol:         req.Symbol,
		Interval:       req.Interval,
		Commission:     req.Commission,
		CandlesUsed:    len(candles),
		RecentTrades:   recentTrades,
		ChartData:      chartData,
	}

	response.Success(c, resp)
}

// GetStrategies godoc
// @Summary Get available strategies
// @Description Get list of available backtest strategies and their parameters
// @Tags backtest
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /backtest/strategies [get]
func (h *BacktestHandler) GetStrategies(c *gin.Context) {
	strategies := map[string]interface{}{
		"ma_cross": map[string]interface{}{
			"name":        "Moving Average Crossover",
			"description": "Buy when fast MA crosses above slow MA, sell when it crosses below",
			"parameters": map[string]interface{}{
				"fast_period": map[string]interface{}{
					"type":        "integer",
					"default":     10,
					"description": "Fast moving average period",
				},
				"slow_period": map[string]interface{}{
					"type":        "integer",
					"default":     30,
					"description": "Slow moving average period",
				},
			},
		},
		"rsi": map[string]interface{}{
			"name":        "RSI Strategy",
			"description": "Buy when RSI is oversold, sell when overbought",
			"parameters": map[string]interface{}{
				"rsi_period": map[string]interface{}{
					"type":        "integer",
					"default":     14,
					"description": "RSI calculation period",
				},
				"rsi_oversold": map[string]interface{}{
					"type":        "float",
					"default":     30,
					"description": "Oversold threshold (buy signal)",
				},
				"rsi_overbought": map[string]interface{}{
					"type":        "float",
					"default":     70,
					"description": "Overbought threshold (sell signal)",
				},
				"position_size": map[string]interface{}{
					"type":        "float",
					"default":     0.01,
					"description": "Position size per trade",
				},
			},
		},
		"bb_rsi": map[string]interface{}{
			"name":        "Bollinger Bands + RSI",
			"description": "Combined strategy using BB and RSI confirmation",
			"parameters": map[string]interface{}{
				"rsi_period": map[string]interface{}{
					"type":        "integer",
					"default":     14,
					"description": "RSI calculation period",
				},
				"rsi_oversold": map[string]interface{}{
					"type":        "float",
					"default":     30,
					"description": "Oversold threshold",
				},
				"rsi_overbought": map[string]interface{}{
					"type":        "float",
					"default":     70,
					"description": "Overbought threshold",
				},
				"position_size": map[string]interface{}{
					"type":        "float",
					"default":     0.01,
					"description": "Position size per trade",
				},
			},
		},
	}

	response.Success(c, strategies)
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value*100)
}

// buildChartData creates chart data from backtest result
func buildChartData(result *backtest.Result, strat backtest.Strategy) *ChartData {
	// Build equity curve
	equityCurve := make([]EquityChartPoint, 0, len(result.EquityCurve))
	for _, point := range result.EquityCurve {
		equityCurve = append(equityCurve, EquityChartPoint{
			Timestamp: point.Timestamp.Format(time.RFC3339),
			Equity:    point.Equity,
			Price:     point.Price,
		})
	}

	// Build trade chart points
	trades := make([]TradeChartPoint, 0, len(result.Trades))
	for _, trade := range result.Trades {
		// Calculate equity at trade time
		equity := trade.Balance + (trade.Position * trade.Price)

		trades = append(trades, TradeChartPoint{
			Timestamp: trade.Timestamp.Format(time.RFC3339),
			Side:      string(trade.Side),
			Price:     trade.Price,
			Equity:    equity,
		})
	}

	// Build indicator data based on strategy type
	var indicators *IndicatorData
	if maStrategy, ok := strat.(*strategy.MACrossStrategy); ok {
		prices, timestamps, fastMA, slowMA := maStrategy.GetIndicatorData()

		// Build price data
		priceData := make([]PricePoint, 0, len(prices))
		for i, price := range prices {
			priceData = append(priceData, PricePoint{
				Timestamp: timestamps[i],
				Price:     price,
			})
		}

		// Build MA data - align with slowPeriod start
		fastMAData := make([]IndicatorPoint, 0, len(fastMA))
		slowMAData := make([]IndicatorPoint, 0, len(slowMA))

		// Calculate offset for alignment
		offset := len(timestamps) - len(slowMA)

		for i := range slowMA {
			timestamp := timestamps[i+offset]

			slowMAData = append(slowMAData, IndicatorPoint{
				Timestamp: timestamp,
				Value:     slowMA[i],
			})

			if i < len(fastMA) {
				fastMAData = append(fastMAData, IndicatorPoint{
					Timestamp: timestamp,
					Value:     fastMA[i],
				})
			}
		}

		indicators = &IndicatorData{
			PriceData: priceData,
			FastMA:    fastMAData,
			SlowMA:    slowMAData,
		}
	}

	return &ChartData{
		EquityCurve: equityCurve,
		Trades:      trades,
		Indicators:  indicators,
	}
}
