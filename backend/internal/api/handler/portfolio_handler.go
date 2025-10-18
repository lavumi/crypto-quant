package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/response"
	"github.com/lavumi/crypto-quant/internal/portfolio"
)

// PortfolioHandler handles portfolio-related requests
type PortfolioHandler struct {
	portfolioService *portfolio.Service
}

// NewPortfolioHandler creates a new portfolio handler
func NewPortfolioHandler(portfolioService *portfolio.Service) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: portfolioService,
	}
}

// GetPosition godoc
// @Summary Get position
// @Description Get position for a specific symbol
// @Tags portfolio
// @Param symbol path string true "Trading symbol (e.g., BTCUSDT)"
// @Success 200 {object} response.Response
// @Router /portfolio/position/{symbol} [get]
func (h *PortfolioHandler) GetPosition(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		response.BadRequestResponse(c, "symbol is required")
		return
	}

	position, err := h.portfolioService.GetPosition(symbol)
	if err != nil {
		response.NotFoundResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, position)
}

// GetAllPositions godoc
// @Summary Get all positions
// @Description Get all portfolio positions
// @Tags portfolio
// @Success 200 {object} response.Response
// @Router /portfolio/positions [get]
func (h *PortfolioHandler) GetAllPositions(c *gin.Context) {
	positions := h.portfolioService.GetAllPositions()
	response.SuccessResponse(c, positions)
}

// GetPnL godoc
// @Summary Get PnL
// @Description Calculate profit/loss for a symbol
// @Tags portfolio
// @Param symbol path string true "Trading symbol (e.g., BTCUSDT)"
// @Success 200 {object} response.Response
// @Router /portfolio/pnl/{symbol} [get]
func (h *PortfolioHandler) GetPnL(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		response.BadRequestResponse(c, "symbol is required")
		return
	}

	unrealizedPnL, realizedPnL, err := h.portfolioService.CalculatePnL(c.Request.Context(), symbol)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{
		"symbol":         symbol,
		"unrealized_pnl": unrealizedPnL,
		"realized_pnl":   realizedPnL,
		"total_pnl":      unrealizedPnL + realizedPnL,
	})
}

// GetTotalPnL godoc
// @Summary Get total PnL
// @Description Get total profit/loss across all positions
// @Tags portfolio
// @Success 200 {object} response.Response
// @Router /portfolio/pnl [get]
func (h *PortfolioHandler) GetTotalPnL(c *gin.Context) {
	unrealizedPnL, realizedPnL := h.portfolioService.GetTotalPnL()

	response.SuccessResponse(c, gin.H{
		"unrealized_pnl": unrealizedPnL,
		"realized_pnl":   realizedPnL,
		"total_pnl":      unrealizedPnL + realizedPnL,
	})
}
