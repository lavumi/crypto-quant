package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/response"
	"github.com/lavumi/crypto-quant/internal/domain"
	order "github.com/lavumi/crypto-quant/internal/trading"
)

// OrderHandler handles order-related requests
type OrderHandler struct {
	orderService *order.Service
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *order.Service) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// PlaceOrderRequest represents a request to place an order
type PlaceOrderRequest struct {
	Symbol   string  `json:"symbol" binding:"required"`
	Side     string  `json:"side" binding:"required"`
	Type     string  `json:"type" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price"`
}

// PlaceOrder godoc
// @Summary Place order
// @Description Place a new trading order
// @Tags orders
// @Accept json
// @Param order body PlaceOrderRequest true "Order details"
// @Success 201 {object} response.Response
// @Router /orders [post]
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, err.Error())
		return
	}

	// Validate order side
	var side domain.OrderSide
	switch req.Side {
	case "BUY":
		side = domain.OrderSideBuy
	case "SELL":
		side = domain.OrderSideSell
	default:
		response.BadRequestResponse(c, "invalid order side, must be BUY or SELL")
		return
	}

	// Validate order type
	var orderType domain.OrderType
	switch req.Type {
	case "MARKET":
		orderType = domain.OrderTypeMarket
	case "LIMIT":
		orderType = domain.OrderTypeLimit
	default:
		response.BadRequestResponse(c, "invalid order type, must be MARKET or LIMIT")
		return
	}

	// Validate price for limit orders
	if orderType == domain.OrderTypeLimit && req.Price <= 0 {
		response.BadRequestResponse(c, "price is required for limit orders")
		return
	}

	order := &domain.Order{
		Symbol:   req.Symbol,
		Side:     side,
		Type:     orderType,
		Quantity: req.Quantity,
		Price:    req.Price,
	}

	executedOrder, err := h.orderService.PlaceOrder(c.Request.Context(), order)
	if err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.CreatedResponse(c, executedOrder)
}

// GetOrder godoc
// @Summary Get order
// @Description Get order details by ID
// @Tags orders
// @Param orderId path string true "Order ID"
// @Success 200 {object} response.Response
// @Router /orders/{orderId} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		response.BadRequestResponse(c, "order ID is required")
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.NotFoundResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, order)
}

// CancelOrder godoc
// @Summary Cancel order
// @Description Cancel an existing order
// @Tags orders
// @Param orderId path string true "Order ID"
// @Success 200 {object} response.Response
// @Router /orders/{orderId} [delete]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		response.BadRequestResponse(c, "order ID is required")
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), orderID); err != nil {
		response.InternalErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{
		"message":  "Order cancelled successfully",
		"order_id": orderID,
	})
}
