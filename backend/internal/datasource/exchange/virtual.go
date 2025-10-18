package exchange

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/lavumi/crypto-quant/internal/domain"
)

// VirtualExchange implements a virtual exchange for testing and simulation
type VirtualExchange struct {
	mu               sync.RWMutex
	prices           map[string]float64
	orders           map[string]*domain.Order
	priceSubscribers map[string][]chan float64
	closeCh          chan struct{}
	closed           bool
}

// NewVirtualExchange creates a new virtual exchange
func NewVirtualExchange(initialPrices map[string]float64) *VirtualExchange {
	if initialPrices == nil {
		initialPrices = map[string]float64{
			"BTCUSDT": 45000.0,
			"ETHUSDT": 3000.0,
		}
	}

	ve := &VirtualExchange{
		prices:           initialPrices,
		orders:           make(map[string]*domain.Order),
		priceSubscribers: make(map[string][]chan float64),
		closeCh:          make(chan struct{}),
	}

	// Start price simulation
	go ve.simulatePriceUpdates()

	return ve
}

// simulatePriceUpdates simulates price movements using random walk
func (ve *VirtualExchange) simulatePriceUpdates() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ve.updatePrices()
			ve.checkLimitOrders()
		case <-ve.closeCh:
			return
		}
	}
}

// updatePrices updates prices with random walk
func (ve *VirtualExchange) updatePrices() {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	for symbol, price := range ve.prices {
		// Random walk: ±0.5% price change
		change := (rand.Float64() - 0.5) * 0.01 * price
		newPrice := price + change

		// Ensure price stays positive
		if newPrice <= 0 {
			newPrice = price
		}

		ve.prices[symbol] = newPrice

		// Notify subscribers
		if subscribers, ok := ve.priceSubscribers[symbol]; ok {
			for _, ch := range subscribers {
				select {
				case ch <- newPrice:
				default:
					// Skip if channel is full
				}
			}
		}
	}
}

// checkLimitOrders checks if any limit orders should be filled
func (ve *VirtualExchange) checkLimitOrders() {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	for _, order := range ve.orders {
		if order.Status != domain.OrderStatusNew || order.Type != domain.OrderTypeLimit {
			continue
		}

		currentPrice, ok := ve.prices[order.Symbol]
		if !ok {
			continue
		}

		shouldFill := false
		if order.Side == domain.OrderSideBuy && currentPrice <= order.Price {
			shouldFill = true
		} else if order.Side == domain.OrderSideSell && currentPrice >= order.Price {
			shouldFill = true
		}

		if shouldFill {
			ve.fillOrder(order, order.Price)
		}
	}
}

// fillOrder fills an order
func (ve *VirtualExchange) fillOrder(order *domain.Order, price float64) {
	now := time.Now()
	order.Status = domain.OrderStatusFilled
	order.FilledQty = order.Quantity
	order.AvgPrice = price
	order.ExecutedAt = &now
	order.UpdatedAt = now
}

// GetCurrentPrice returns the current market price
func (ve *VirtualExchange) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	ve.mu.RLock()
	defer ve.mu.RUnlock()

	price, ok := ve.prices[symbol]
	if !ok {
		return 0, fmt.Errorf("symbol not found: %s", symbol)
	}

	return price, nil
}

// GetCandles retrieves historical candlestick data
func (ve *VirtualExchange) GetCandles(ctx context.Context, symbol, interval string, limit int) ([]*domain.Candle, error) {
	price, err := ve.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	candles := make([]*domain.Candle, limit)
	for i := 0; i < limit; i++ {
		candles[i] = &domain.Candle{
			Symbol:    symbol,
			OpenTime:  now.Add(time.Duration(-limit+i) * time.Minute),
			CloseTime: now.Add(time.Duration(-limit+i+1) * time.Minute),
			Open:      price,
			High:      price * 1.001,
			Low:       price * 0.999,
			Close:     price,
			Volume:    rand.Float64() * 1000,
		}
	}

	return candles, nil
}

// PlaceOrder submits a new order
func (ve *VirtualExchange) PlaceOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	if ve.closed {
		return nil, fmt.Errorf("exchange is closed")
	}

	// Generate order ID if not provided
	if order.ID == "" {
		order.ID = fmt.Sprintf("ORDER_%d", time.Now().UnixNano())
	}

	// Validate order
	if order.Quantity <= 0 {
		order.Status = domain.OrderStatusRejected
		return order, fmt.Errorf("invalid quantity: %f", order.Quantity)
	}

	currentPrice, ok := ve.prices[order.Symbol]
	if !ok {
		order.Status = domain.OrderStatusRejected
		return order, fmt.Errorf("symbol not found: %s", order.Symbol)
	}

	now := time.Now()
	order.Status = domain.OrderStatusNew
	order.CreatedAt = now
	order.UpdatedAt = now

	// Market orders are filled immediately
	if order.Type == domain.OrderTypeMarket {
		ve.fillOrder(order, currentPrice)
	}

	// Store order
	ve.orders[order.ID] = order

	return order, nil
}

// GetOrder retrieves order information
func (ve *VirtualExchange) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	ve.mu.RLock()
	defer ve.mu.RUnlock()

	order, ok := ve.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	return order, nil
}

// CancelOrder cancels an existing order
func (ve *VirtualExchange) CancelOrder(ctx context.Context, orderID string) error {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	order, ok := ve.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status != domain.OrderStatusNew {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	return nil
}

// SubscribePrice subscribes to price updates
func (ve *VirtualExchange) SubscribePrice(ctx context.Context, symbol string) (<-chan float64, error) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	if _, ok := ve.prices[symbol]; !ok {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	ch := make(chan float64, 100)
	ve.priceSubscribers[symbol] = append(ve.priceSubscribers[symbol], ch)

	// Send current price immediately
	ch <- ve.prices[symbol]

	return ch, nil
}

// Close closes the exchange connection
func (ve *VirtualExchange) Close() error {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	if ve.closed {
		return nil
	}

	ve.closed = true
	close(ve.closeCh)

	// Close all subscriber channels
	for _, subscribers := range ve.priceSubscribers {
		for _, ch := range subscribers {
			close(ch)
		}
	}

	return nil
}

// SetPrice sets the price for a symbol (for testing purposes)
func (ve *VirtualExchange) SetPrice(symbol string, price float64) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	if price <= 0 {
		return
	}

	ve.prices[symbol] = math.Round(price*100) / 100
}
