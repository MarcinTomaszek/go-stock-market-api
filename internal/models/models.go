package models

type Stock struct {
	Name     string `json:"name" binding:"required"`
	Quantity int    `json:"quantity" binding:"gte=0"`
}

type Wallet struct {
	ID     string  `json:"id"`
	Stocks []Stock `json:"stocks"`
}

type LogEntry struct {
	Type      string `json:"type"`
	WalletID  string `json:"wallet_id"`
	StockName string `json:"stock_name"`
}

type TradeRequest struct {
	Type string `json:"type" binding:"required,oneof=buy sell"`
}

type BankStateRequest struct {
	Stocks []Stock `json:"stocks" binding:"required,dive"`
}

type BankStateResponse struct {
	Stocks []Stock `json:"stocks"`
}

type LogResponse struct {
	Log []LogEntry `json:"log"`
}
