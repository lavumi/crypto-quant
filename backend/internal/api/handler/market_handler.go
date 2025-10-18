package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/response"
	"github.com/lavumi/crypto-quant/internal/datasource/market/price"
)

// MarketHandler handles market-related requests
type MarketHandler struct {
	marketService *price.Service
}

// NewMarketHandler creates a new market handler
func NewMarketHandler(marketService *price.Service) *MarketHandler {
	return &MarketHandler{
		marketService: marketService,
	}
}

// GetPrice godoc
// @Summary Get current price
// @Description Get the current price for a symbol
// @Tags market
// @Param symbol path string true "Trading symbol (e.g., BTCUSDT)"
// @Success 200 {object} response.Response
// @Router /market/price/{symbol} [get]
func (h *MarketHandler) GetPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		response.BadRequestResponse(c, "symbol is required")
		return
	}

	price, err := h.marketService.GetPrice(c.Request.Context(), symbol)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{
		"symbol": symbol,
		"price":  price,
	})
}

// GetMultiplePrices godoc
// @Summary Get multiple prices
// @Description Get current prices for multiple symbols
// @Tags market
// @Param symbols query string true "Comma-separated symbols (e.g., BTCUSDT,ETHUSDT)"
// @Success 200 {object} response.Response
// @Router /market/prices [get]
func (h *MarketHandler) GetMultiplePrices(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		response.BadRequestResponse(c, "symbols parameter is required")
		return
	}

	// Parse comma-separated symbols
	symbols := parseSymbols(symbolsParam)
	if len(symbols) == 0 {
		response.BadRequestResponse(c, "no valid symbols provided")
		return
	}

	prices, err := h.marketService.GetMultiplePrices(c.Request.Context(), symbols)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, prices)
}

// StreamPrice godoc
// @Summary Stream real-time price (SSE)
// @Description Stream real-time price updates via Server-Sent Events
// @Tags market
// @Param symbol path string true "Trading symbol (e.g., BTCUSDT)"
// @Param interval query int false "Update interval in seconds (default: 1)"
// @Success 200 {string} string "text/event-stream"
// @Router /market/stream/{symbol} [get]
func (h *MarketHandler) StreamPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		response.BadRequestResponse(c, "symbol is required")
		return
	}

	// Parse interval (default: 1 second)
	interval := 1
	if intervalParam := c.Query("interval"); intervalParam != "" {
		if i, err := parseIntParam(intervalParam); err == nil && i > 0 {
			interval = i
		}
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Stream prices
	ticker := createTicker(interval)
	defer ticker.Stop()

	ctx := c.Request.Context()
	clientGone := ctx.Done()

	for {
		select {
		case <-clientGone:
			return
		case <-ticker.C:
			price, err := h.marketService.GetPrice(ctx, symbol)
			if err != nil {
				c.SSEvent("error", err.Error())
				return
			}

			c.SSEvent("price", gin.H{
				"symbol": symbol,
				"price":  price,
				"time":   currentTime(),
			})
			c.Writer.Flush()
		}
	}
}

// parseIntParam parses an integer parameter from a string
func parseIntParam(s string) (int, error) {
	var result int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, gin.Error{Err: gin.Error{}.Err}
		}
		result = result*10 + int(ch-'0')
	}
	return result, nil
}

// parseSymbols splits a comma-separated string into a slice of symbols
func parseSymbols(symbolsParam string) []string {
	var symbols []string
	for _, s := range splitByComma(symbolsParam) {
		if s != "" {
			symbols = append(symbols, s)
		}
	}
	return symbols
}

// splitByComma splits a string by comma
func splitByComma(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// createTicker creates a new ticker with the given interval in seconds
func createTicker(intervalSeconds int) *time.Ticker {
	return time.NewTicker(time.Duration(intervalSeconds) * time.Second)
}

// currentTime returns the current time in RFC3339 format
func currentTime() string {
	return time.Now().Format(time.RFC3339)
}
