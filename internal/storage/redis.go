package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/marcintomaszek/stock-market/internal/models"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(redisAddr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("Can not connect with Redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

func (s *RedisStore) AppendLog(ctx context.Context, logEntry any) error {
	data, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}

	pipe := s.client.Pipeline()
	pipe.RPush(ctx, "audit_log", data)
	pipe.LTrim(ctx, "audit_log", -10000, -1)
	_, err = pipe.Exec(ctx)
	return err
}

var ErrInsufficientFunds = errors.New("insufficient stocks")

var ErrStockNotFound = errors.New("stock not found")

func (s *RedisStore) BuyStock(ctx context.Context, walletID, stockName string, logData []byte) error {
	walletKey := fmt.Sprintf("wallet:%s:stocks", walletID)
	bankKey := "bank:stocks"

	err := s.client.Watch(ctx, func(tx *redis.Tx) error {
		bankQuantity, err := tx.HGet(ctx, bankKey, stockName).Int()

		if err != nil {
			if err == redis.Nil {
				return ErrStockNotFound
			}
			return err
		}

		if bankQuantity <= 0 {
			return ErrInsufficientFunds
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HIncrBy(ctx, bankKey, stockName, -1)
			pipe.HIncrBy(ctx, walletKey, stockName, 1)

			pipe.RPush(ctx, "audit_log", logData)
			pipe.LTrim(ctx, "audit_log", -10000, -1)

			return nil
		})

		return err
	}, bankKey)

	return err
}

func (s *RedisStore) SellStock(ctx context.Context, walletID, stockName string, logData []byte) error {
	walletKey := fmt.Sprintf("wallet:%s:stocks", walletID)
	bankKey := "bank:stocks"

	err := s.client.Watch(ctx, func(tx *redis.Tx) error {
		walletQuantity, err := tx.HGet(ctx, walletKey, stockName).Int()

		if err != nil {
			if err == redis.Nil {
				return ErrStockNotFound
			}
			return err
		}

		if walletQuantity <= 0 {
			return ErrInsufficientFunds
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HIncrBy(ctx, walletKey, stockName, -1)
			pipe.HIncrBy(ctx, bankKey, stockName, 1)

			pipe.RPush(ctx, "audit_log", logData)
			pipe.LTrim(ctx, "audit_log", -10000, -1)

			return nil
		})

		return err
	}, walletKey)

	return err
}

func (s *RedisStore) GetWallet(ctx context.Context, walletID string) (models.Wallet, error) {
	walletKey := fmt.Sprintf("wallet:%s:stocks", walletID)

	data, err := s.client.HGetAll(ctx, walletKey).Result()
	if err != nil {
		return models.Wallet{}, err
	}

	wallet := models.Wallet{ID: walletID, Stocks: []models.Stock{}}

	for name, qtyStr := range data {
		qty, _ := strconv.Atoi(qtyStr)
		if qty > 0 {
			wallet.Stocks = append(wallet.Stocks, models.Stock{Name: name, Quantity: qty})
		}
	}
	return wallet, nil
}

func (s *RedisStore) GetWalletStockQuantity(ctx context.Context, walletID, stockName string) (int, error) {
	walletKey := fmt.Sprintf("wallet:%s:stocks", walletID)
	qty, err := s.client.HGet(ctx, walletKey, stockName).Int()

	if err == redis.Nil {
		return 0, nil
	}
	return qty, err
}

func (s *RedisStore) GetBankState(ctx context.Context) ([]models.Stock, error) {
	data, err := s.client.HGetAll(ctx, "bank:stocks").Result()
	if err != nil {
		return nil, err
	}

	stocks := []models.Stock{}
	for name, qtyStr := range data {
		qty, _ := strconv.Atoi(qtyStr)
		stocks = append(stocks, models.Stock{Name: name, Quantity: qty})
	}
	return stocks, nil
}

func (s *RedisStore) SetBankState(ctx context.Context, stocks []models.Stock) error {

	pipe := s.client.Pipeline()

	pipe.Del(ctx, "bank:stocks")

	for _, stock := range stocks {
		pipe.HSet(ctx, "bank:stocks", stock.Name, stock.Quantity)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetAuditLog(ctx context.Context) ([]models.LogEntry, error) {
	rawData, err := s.client.LRange(ctx, "audit_log", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	logs := make([]models.LogEntry, 0, len(rawData))
	for _, item := range rawData {
		var entry models.LogEntry
		json.Unmarshal([]byte(item), &entry)
		logs = append(logs, entry)
	}
	return logs, nil
}
