package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/response"
	"github.com/lavumi/crypto-quant/internal/datasource/market/history"
)

// DataHandler handles historical data requests
type DataHandler struct {
	dataService *history.Service
}

// NewDataHandler creates a new data handler
func NewDataHandler(dataService *history.Service) *DataHandler {
	return &DataHandler{
		dataService: dataService,
	}
}

// CollectHistoricalData godoc
// @Summary Collect historical data
// @Description Collect historical candle data from Binance (limited to 90 days via API, use CLI for larger datasets)
// @Tags data
// @Param symbol query string true "Trading symbol (e.g., BTCUSDT)"
// @Param interval query string true "Candle interval (e.g., 1h, 1d)"
// @Param days query int false "Number of days to collect (from today backwards, max 90)"
// @Param start query string false "Start date (YYYY-MM-DD format)"
// @Param end query string false "End date (YYYY-MM-DD format)"
// @Success 200 {object} response.Response
// @Router /data/collect [post]
func (h *DataHandler) CollectHistoricalData(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.Query("interval")

	if symbol == "" || interval == "" {
		response.BadRequestResponse(c, "symbol and interval are required")
		return
	}

	// API limit to prevent timeouts
	const maxDaysViaAPI = 90

	var startTime, endTime time.Time
	var err error

	// Priority: start/end dates > days
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr != "" && endStr != "" {
		// Use start/end dates
		startTime, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			response.BadRequestResponse(c, "invalid start date format, use YYYY-MM-DD")
			return
		}
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)

		endTime, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			response.BadRequestResponse(c, "invalid end date format, use YYYY-MM-DD")
			return
		}
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)

		// Check date range limit
		daysDiff := int(endTime.Sub(startTime).Hours() / 24)
		if daysDiff > maxDaysViaAPI {
			response.BadRequestResponse(c,
				"Date range too large. API is limited to 90 days to prevent timeouts. "+
					"For larger datasets, please use the CLI: "+
					"./bin/collector -symbol "+symbol+" -interval "+interval+" -start "+startStr+" -end "+endStr)
			return
		}
	} else {
		// Use days parameter
		days := 30
		if daysParam := c.Query("days"); daysParam != "" {
			var d int
			for _, ch := range daysParam {
				if ch < '0' || ch > '9' {
					break
				}
				d = d*10 + int(ch-'0')
			}
			if d > 0 {
				days = d
			}
		}

		// Check days limit
		if days > maxDaysViaAPI {
			response.BadRequestResponse(c,
				"Too many days requested. API is limited to 90 days to prevent timeouts. "+
					"For larger datasets, please use the CLI: "+
					"./bin/collector -symbol "+symbol+" -interval "+interval+" -days "+c.Query("days"))
			return
		}

		endTime = time.Now()
		startTime = endTime.AddDate(0, 0, -days)
	}

	if err := h.dataService.CollectHistoricalData(c.Request.Context(), symbol, interval, startTime, endTime); err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{
		"message":    "Historical data collection completed",
		"symbol":     symbol,
		"interval":   interval,
		"start_time": startTime,
		"end_time":   endTime,
	})
}

// GetCandles godoc
// @Summary Get candles
// @Description Get historical candle data within a time range
// @Tags data
// @Param symbol query string true "Trading symbol (e.g., BTCUSDT)"
// @Param interval query string true "Candle interval (e.g., 1h, 1d)"
// @Param start query string true "Start time (RFC3339 format)"
// @Param end query string true "End time (RFC3339 format)"
// @Success 200 {object} response.Response
// @Router /data/candles [get]
func (h *DataHandler) GetCandles(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.Query("interval")
	startStr := c.Query("start")
	endStr := c.Query("end")

	if symbol == "" || interval == "" || startStr == "" || endStr == "" {
		response.BadRequestResponse(c, "symbol, interval, start, and end are required")
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		response.BadRequestResponse(c, "invalid start time format, use RFC3339")
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		response.BadRequestResponse(c, "invalid end time format, use RFC3339")
		return
	}

	candles, err := h.dataService.GetCandles(c.Request.Context(), symbol, interval, startTime, endTime)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, candles)
}

// GetLatestCandle godoc
// @Summary Get latest candle
// @Description Get the most recent candle for a symbol
// @Tags data
// @Param symbol query string true "Trading symbol (e.g., BTCUSDT)"
// @Param interval query string true "Candle interval (e.g., 1h, 1d)"
// @Success 200 {object} response.Response
// @Router /data/candles/latest [get]
func (h *DataHandler) GetLatestCandle(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.Query("interval")

	if symbol == "" || interval == "" {
		response.BadRequestResponse(c, "symbol and interval are required")
		return
	}

	candle, err := h.dataService.GetLatestCandle(c.Request.Context(), symbol, interval)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, candle)
}

// GetTradeHistory godoc
// @Summary Get trade history
// @Description Get all trades for a symbol
// @Tags data
// @Param symbol query string true "Trading symbol (e.g., BTCUSDT)"
// @Success 200 {object} response.Response
// @Router /data/trades [get]
func (h *DataHandler) GetTradeHistory(c *gin.Context) {
	symbol := c.Query("symbol")

	if symbol == "" {
		response.BadRequestResponse(c, "symbol is required")
		return
	}

	trades, err := h.dataService.GetTradeHistory(c.Request.Context(), symbol)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, trades)
}

// ValidateData godoc
// @Summary Validate data availability
// @Description Check if historical data is available for the requested period
// @Tags data
// @Param symbol query string true "Trading symbol (e.g., BTCUSDT)"
// @Param interval query string true "Candle interval (e.g., 1h, 1d)"
// @Param start query string true "Start date (YYYY-MM-DD)"
// @Param end query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.Response
// @Router /data/validate [get]
func (h *DataHandler) ValidateData(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.Query("interval")
	startStr := c.Query("start")
	endStr := c.Query("end")

	if symbol == "" || interval == "" || startStr == "" || endStr == "" {
		response.BadRequestResponse(c, "symbol, interval, start, and end are required")
		return
	}

	// Parse dates in UTC
	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		response.BadRequestResponse(c, "invalid start date format, use YYYY-MM-DD")
		return
	}
	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)

	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		response.BadRequestResponse(c, "invalid end date format, use YYYY-MM-DD")
		return
	}
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)

	// Validate data availability
	result, err := h.dataService.ValidateDataAvailability(c.Request.Context(), symbol, interval, startTime, endTime)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result)
}
