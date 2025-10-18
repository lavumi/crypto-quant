package api

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/handler"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter initializes the Gin router with all routes and middleware
func SetupRouter(
	marketHandler *handler.MarketHandler,
	dataHandler *handler.DataHandler,
	walletHandler *handler.WalletHandler,
	portfolioHandler *handler.PortfolioHandler,
	orderHandler *handler.OrderHandler,
	backtestHandler *handler.BacktestHandler,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Market routes
		market := v1.Group("/market")
		{
			market.GET("/price/:symbol", marketHandler.GetPrice)
			market.GET("/prices", marketHandler.GetMultiplePrices)
			market.GET("/stream/:symbol", marketHandler.StreamPrice)
		}

		// Data routes
		data := v1.Group("/data")
		{
			data.POST("/collect", dataHandler.CollectHistoricalData)
			data.GET("/candles", dataHandler.GetCandles)
			data.GET("/candles/latest", dataHandler.GetLatestCandle)
			data.GET("/trades", dataHandler.GetTradeHistory)
			data.GET("/validate", dataHandler.ValidateData)
		}

		// Wallet routes
		wallet := v1.Group("/wallet")
		{
			wallet.GET("/balance/:asset", walletHandler.GetBalance)
			wallet.GET("/balances", walletHandler.GetAllBalances)
		}

		// Portfolio routes
		portfolio := v1.Group("/portfolio")
		{
			portfolio.GET("/position/:symbol", portfolioHandler.GetPosition)
			portfolio.GET("/positions", portfolioHandler.GetAllPositions)
			portfolio.GET("/pnl/:symbol", portfolioHandler.GetPnL)
			portfolio.GET("/pnl", portfolioHandler.GetTotalPnL)
		}

		// Order routes
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.PlaceOrder)
			orders.GET("/:orderId", orderHandler.GetOrder)
			orders.DELETE("/:orderId", orderHandler.CancelOrder)
		}

		// Backtest routes
		backtest := v1.Group("/backtest")
		{
			backtest.POST("/run", backtestHandler.RunBacktest)
			backtest.GET("/strategies", backtestHandler.GetStrategies)
		}
	}

	return router
}

// LoggerMiddleware is a custom logger middleware
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[API] %3d | %13v | %15s | %-7s %s",
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
