package api

import (
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
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

	// Serve embedded frontend static files
	frontendFS, err := GetFrontendFS()
	if err != nil {
		log.Printf("Warning: Failed to load embedded frontend: %v", err)
	} else {
		// Serve static files (CSS, JS, images, etc.)
		router.Use(serveStaticFiles(frontendFS))
		// SPA fallback - serve index.html for all non-API routes
		router.NoRoute(serveSPAFallback(frontendFS))
	}

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

		// Skip logging for static files and frontend assets
		if shouldSkipLogging(path) {
			return
		}

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

// shouldSkipLogging determines if a request should be skipped from logging
func shouldSkipLogging(path string) bool {
	// Skip frontend static assets
	if len(path) >= 5 && path[:5] == "/_app" {
		return true
	}
	// Skip common static files
	switch path {
	case "/favicon.ico", "/robots.txt", "/":
		return true
	}
	// Skip static file extensions
	if len(path) > 4 {
		ext := path[len(path)-4:]
		switch ext {
		case ".js", ".css", ".svg", ".png", ".jpg", "jpeg", ".gif", ".ico", "woff", "ttf", "eot", ".map", "json":
			return true
		}
	}
	return false
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

// serveStaticFiles serves static files from the embedded filesystem
func serveStaticFiles(frontendFS fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(frontendFS))

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes, health check, and swagger
		if len(path) >= 4 && path[:4] == "/api" {
			c.Next()
			return
		}
		if path == "/health" || len(path) >= 8 && path[:8] == "/swagger" {
			c.Next()
			return
		}

		// Check if file exists in embedded FS
		if _, err := fs.Stat(frontendFS, filepath.Clean(path[1:])); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		c.Next()
	}
}

// serveSPAFallback serves index.html for all unmatched routes (SPA routing)
func serveSPAFallback(frontendFS fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read index.html from embedded FS
		indexHTML, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not available")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}
}
