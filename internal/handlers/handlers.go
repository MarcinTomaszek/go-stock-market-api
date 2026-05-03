package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/marcintomaszek/stock-market/internal/models"
	"github.com/marcintomaszek/stock-market/internal/storage"
)

type API struct {
	store *storage.RedisStore
}

func NewAPI(store *storage.RedisStore) *API {
	return &API{store: store}
}

// POST /wallets/{wallet_id}/stocks/{stock_name}
func (api *API) TradeStock(c *gin.Context) {
	walletID := c.Param("wallet_id")
	stockName := c.Param("stock_name")

	var req models.TradeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or type"})
		return
	}

	logEntry := models.LogEntry{
		Type:      req.Type,
		WalletID:  walletID,
		StockName: stockName,
	}

	logData, _ := json.Marshal(logEntry)

	var err error
	ctx := c.Request.Context()

	switch req.Type {
	case "buy":
		err = api.store.BuyStock(ctx, walletID, stockName, logData)
	case "sell":
		err = api.store.SellStock(ctx, walletID, stockName, logData)
	}

	if err != nil {
		if errors.Is(err, storage.ErrStockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Stock does not exist"})
			return
		}

		if errors.Is(err, storage.ErrInsufficientFunds) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

// GET /wallets/{wallet_id}
func (api *API) GetWallet(c *gin.Context) {
	walletID := c.Param("wallet_id")

	wallet, err := api.store.GetWallet(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// GET /wallets/{wallet_id}/stocks/{stock_name}
func (api *API) GetWalletStock(c *gin.Context) {
	walletID := c.Param("wallet_id")
	stockName := c.Param("stock_name")

	qty, err := api.store.GetWalletStockQuantity(c.Request.Context(), walletID, stockName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, qty)
}

// GET /stocks
func (api *API) GetBankState(c *gin.Context) {
	stocks, err := api.store.GetBankState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stocks": stocks})
}

// POST /stocks
func (api *API) SetBankState(c *gin.Context) {
	var req models.BankStateRequest // Struktura {stocks: [...]}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := api.store.SetBankState(c.Request.Context(), req.Stocks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

// GET /log
func (api *API) GetAuditLog(c *gin.Context) {
	logs, err := api.store.GetAuditLog(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if logs == nil {
		logs = []models.LogEntry{}
	}

	c.JSON(http.StatusOK, gin.H{"log": logs})
}

// POST /chaos
func (api *API) Chaos(c *gin.Context) {
	os.Exit(1)
}
