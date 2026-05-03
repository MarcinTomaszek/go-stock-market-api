package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/marcintomaszek/stock-market/internal/handlers"
	"github.com/marcintomaszek/stock-market/internal/storage"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	store, err := storage.NewRedisStore(redisAddr)
	if err != nil {
		log.Fatalf("Can not connect with Redis database: %v", err)
	}

	api := handlers.NewAPI(store)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/stocks", api.SetBankState)
	router.GET("/stocks", api.GetBankState)

	router.POST("/wallets/:wallet_id/stocks/:stock_name", api.TradeStock)
	router.GET("/wallets/:wallet_id", api.GetWallet)
	router.GET("/wallets/:wallet_id/stocks/:stock_name", api.GetWalletStock)

	router.GET("/log", api.GetAuditLog)

	router.POST("/chaos", api.Chaos)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting stock-market API on port %s", port)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
