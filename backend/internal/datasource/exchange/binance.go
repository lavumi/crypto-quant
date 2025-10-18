package exchange

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/lavumi/crypto-quant/internal/domain"
)

// BinanceExchange implements Exchange interface for Binance
type BinanceExchange struct {
	client           *binance.Client
	mu               sync.RWMutex
	prices           map[string]float64
	priceSubscribers map[string][]chan float64
	wsCleanup        []func()
	closed           bool
}

// NewBinanceExchange creates a new Binance exchange client
func NewBinanceExchange(apiKey, secretKey string, useTestnet bool) (*BinanceExchange, error) {
	// Set testnet flag before creating client
	if useTestnet {
		binance.UseTestnet = true
	}

	client := binance.NewClient(apiKey, secretKey)

	be := &BinanceExchange{
		client:           client,
		prices:           make(map[string]float64),
		priceSubscribers: make(map[string][]chan float64),
		wsCleanup:        make([]func(), 0),
	}

	return be, nil
}

// SetClient sets the Binance client (used for public API without authentication)
func (be *BinanceExchange) SetClient(client *binance.Client) {
	be.client = client
	if be.prices == nil {
		be.prices = make(map[string]float64)
	}
	if be.priceSubscribers == nil {
		be.priceSubscribers = make(map[string][]chan float64)
	}
	if be.wsCleanup == nil {
		be.wsCleanup = make([]func(), 0)
	}
}

// GetCurrentPrice returns the current market price (always fetches fresh data from API)
func (be *BinanceExchange) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	// Always fetch fresh price from API
	prices, err := be.client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}

	if len(prices) == 0 {
		return 0, fmt.Errorf("no price data for symbol: %s", symbol)
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	// Update cache for WebSocket subscribers
	be.mu.Lock()
	be.prices[symbol] = price
	be.mu.Unlock()

	return price, nil
}

// GetCandles retrieves historical candlestick data
func (be *BinanceExchange) GetCandles(ctx context.Context, symbol, interval string, limit int) ([]*domain.Candle, error) {
	klines, err := be.client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get candles: %w", err)
	}

	candles := make([]*domain.Candle, 0, len(klines))
	for _, k := range klines {
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		closePrice, _ := strconv.ParseFloat(k.Close, 64)
		volume, _ := strconv.ParseFloat(k.Volume, 64)

		candle := &domain.Candle{
			Symbol:    symbol,
			OpenTime:  time.Unix(k.OpenTime/1000, 0),
			CloseTime: time.Unix(k.CloseTime/1000, 0),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

// PlaceOrder submits a new order
func (be *BinanceExchange) PlaceOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	be.mu.RLock()
	if be.closed {
		be.mu.RUnlock()
		return nil, fmt.Errorf("exchange is closed")
	}
	be.mu.RUnlock()

	// Create order service
	var orderService *binance.CreateOrderService

	if order.Side == domain.OrderSideBuy {
		orderService = be.client.NewCreateOrderService().
			Symbol(order.Symbol).
			Side(binance.SideTypeBuy)
	} else {
		orderService = be.client.NewCreateOrderService().
			Symbol(order.Symbol).
			Side(binance.SideTypeSell)
	}

	// Set order type
	if order.Type == domain.OrderTypeMarket {
		orderService = orderService.Type(binance.OrderTypeMarket)
	} else {
		orderService = orderService.
			Type(binance.OrderTypeLimit).
			Price(fmt.Sprintf("%.8f", order.Price)).
			TimeInForce(binance.TimeInForceTypeGTC)
	}

	// Set quantity
	orderService = orderService.Quantity(fmt.Sprintf("%.8f", order.Quantity))

	// Execute order
	response, err := orderService.Do(ctx)
	if err != nil {
		order.Status = domain.OrderStatusRejected
		return order, fmt.Errorf("failed to place order: %w", err)
	}

	// Map response to our order type
	order.ID = fmt.Sprintf("%d", response.OrderID)
	order.CreatedAt = time.Unix(response.TransactTime/1000, 0)
	order.UpdatedAt = order.CreatedAt

	switch response.Status {
	case binance.OrderStatusTypeNew:
		order.Status = domain.OrderStatusNew
	case binance.OrderStatusTypeFilled:
		order.Status = domain.OrderStatusFilled
		executedQty, _ := strconv.ParseFloat(response.ExecutedQuantity, 64)
		order.FilledQty = executedQty

		if len(response.Fills) > 0 {
			totalCost := 0.0
			totalQty := 0.0
			for _, fill := range response.Fills {
				price, _ := strconv.ParseFloat(fill.Price, 64)
				qty, _ := strconv.ParseFloat(fill.Quantity, 64)
				totalCost += price * qty
				totalQty += qty
			}
			if totalQty > 0 {
				order.AvgPrice = totalCost / totalQty
			}
		}
		now := time.Now()
		order.ExecutedAt = &now
	case binance.OrderStatusTypeCanceled:
		order.Status = domain.OrderStatusCancelled
	case binance.OrderStatusTypeRejected:
		order.Status = domain.OrderStatusRejected
	default:
		order.Status = domain.OrderStatusNew
	}

	return order, nil
}

// GetOrder retrieves order information
func (be *BinanceExchange) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	_, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	// Note: We need the symbol to query order, but we don't have it here
	// This is a limitation of the current interface design
	// In production, you might want to store orders locally or change the interface
	return nil, fmt.Errorf("GetOrder not fully implemented for Binance - needs symbol")
}

// CancelOrder cancels an existing order
func (be *BinanceExchange) CancelOrder(ctx context.Context, orderID string) error {
	_, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	// Note: We need the symbol to cancel order
	// This is a limitation of the current interface design
	return fmt.Errorf("CancelOrder not fully implemented for Binance - needs symbol")
}

// StreamKlines streams real-time kline/candle data
func (be *BinanceExchange) StreamKlines(ctx context.Context, symbol, interval string, callback func(*domain.Candle)) error {
	wsHandler := func(event *binance.WsKlineEvent) {
		k := event.Kline
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		closePrice, _ := strconv.ParseFloat(k.Close, 64)
		volume, _ := strconv.ParseFloat(k.Volume, 64)

		candle := &domain.Candle{
			Symbol:    symbol,
			OpenTime:  time.Unix(k.StartTime/1000, 0),
			CloseTime: time.Unix(k.EndTime/1000, 0),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
		}

		callback(candle)
	}

	errHandler := func(err error) {
		fmt.Printf("WebSocket error for %s klines: %v\n", symbol, err)
	}

	doneC, stopC, err := binance.WsKlineServe(symbol, interval, wsHandler, errHandler)
	if err != nil {
		return fmt.Errorf("failed to start kline WebSocket: %w", err)
	}

	be.mu.Lock()
	be.wsCleanup = append(be.wsCleanup, func() {
		close(stopC)
	})
	be.mu.Unlock()

	go func() {
		<-ctx.Done()
		close(stopC)
	}()

	<-doneC
	return nil
}

// SubscribePrice subscribes to price updates via WebSocket
func (be *BinanceExchange) SubscribePrice(ctx context.Context, symbol string) (<-chan float64, error) {
	be.mu.Lock()
	defer be.mu.Unlock()

	ch := make(chan float64, 100)
	be.priceSubscribers[symbol] = append(be.priceSubscribers[symbol], ch)

	// Start WebSocket if this is the first subscriber for this symbol
	if len(be.priceSubscribers[symbol]) == 1 {
		go be.startPriceWebSocket(symbol)
	}

	// Send current price if available
	if price, ok := be.prices[symbol]; ok {
		ch <- price
	}

	return ch, nil
}

// startPriceWebSocket starts a WebSocket connection for price updates
func (be *BinanceExchange) startPriceWebSocket(symbol string) {
	wsHandler := func(event *binance.WsTradeEvent) {
		price, err := strconv.ParseFloat(event.Price, 64)
		if err != nil {
			return
		}

		be.mu.Lock()
		be.prices[symbol] = price
		subscribers := be.priceSubscribers[symbol]
		be.mu.Unlock()

		// Notify all subscribers
		for _, ch := range subscribers {
			select {
			case ch <- price:
			default:
				// Skip if channel is full
			}
		}
	}

	errHandler := func(err error) {
		// Log error but don't crash
		fmt.Printf("WebSocket error for %s: %v\n", symbol, err)
	}

	doneC, stopC, err := binance.WsTradeServe(symbol, wsHandler, errHandler)
	if err != nil {
		fmt.Printf("Failed to start WebSocket for %s: %v\n", symbol, err)
		return
	}

	be.mu.Lock()
	be.wsCleanup = append(be.wsCleanup, func() {
		close(stopC)
	})
	be.mu.Unlock()

	<-doneC
}

// Close closes the exchange connection
func (be *BinanceExchange) Close() error {
	be.mu.Lock()
	defer be.mu.Unlock()

	if be.closed {
		return nil
	}

	be.closed = true

	// Close all WebSocket connections
	for _, cleanup := range be.wsCleanup {
		cleanup()
	}

	// Close all subscriber channels
	for _, subscribers := range be.priceSubscribers {
		for _, ch := range subscribers {
			close(ch)
		}
	}

	return nil
}
