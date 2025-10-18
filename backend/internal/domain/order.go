package domain

import "time"

// OrderSide represents buy or sell direction
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "NEW"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

// Order represents a trading order
type Order struct {
	ID         string      `json:"id"`
	Symbol     string      `json:"symbol"`
	Side       OrderSide   `json:"side"`
	Type       OrderType   `json:"type"`
	Quantity   float64     `json:"quantity"`
	Price      float64     `json:"price"` // 0 for market orders
	Status     OrderStatus `json:"status"`
	FilledQty  float64     `json:"filled_qty"`
	AvgPrice   float64     `json:"avg_price"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	ExecutedAt *time.Time  `json:"executed_at,omitempty"`
}

// Trade represents an executed trade
type Trade struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Symbol    string    `json:"symbol"`
	Side      OrderSide `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Fee       float64   `json:"fee"`
	FeeAsset  string    `json:"fee_asset"`
	Timestamp time.Time `json:"timestamp"`
}
