package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lavumi/crypto-quant/internal/api/response"
	"github.com/lavumi/crypto-quant/internal/portfolio/wallet"
)

// WalletHandler handles wallet-related requests
type WalletHandler struct {
	walletService *wallet.Service
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(walletService *wallet.Service) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// GetBalance godoc
// @Summary Get balance
// @Description Get balance for a specific asset
// @Tags wallet
// @Param asset path string true "Asset symbol (e.g., USDT, BTC)"
// @Success 200 {object} response.Response
// @Router /wallet/balance/{asset} [get]
func (h *WalletHandler) GetBalance(c *gin.Context) {
	asset := c.Param("asset")
	if asset == "" {
		response.BadRequestResponse(c, "asset is required")
		return
	}

	balance, err := h.walletService.GetBalance(asset)
	if err != nil {
		response.NotFoundResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, balance)
}

// GetAllBalances godoc
// @Summary Get all balances
// @Description Get all wallet balances
// @Tags wallet
// @Success 200 {object} response.Response
// @Router /wallet/balances [get]
func (h *WalletHandler) GetAllBalances(c *gin.Context) {
	balances := h.walletService.GetAllBalances()
	response.SuccessResponse(c, balances)
}
