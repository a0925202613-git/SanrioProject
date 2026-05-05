package redis

import (
	"context"
	"log"

	"sanrio-auction-api/config"

	"github.com/redis/go-redis/v9"
)

// noopLogger 取代 go-redis 內建的 logger，避免連線失敗時印出大量 pool retry 訊息。
type noopLogger struct{}

func (noopLogger) Printf(_ context.Context, _ string, _ ...interface{}) {}

// NewClient 連線到 Redis。若 Redis 未啟動則回傳 nil（並印出一次警告），
// 應用程式會在沒有快取的情況下繼續運作。
func NewClient(cfg config.RedisConfig) *redis.Client {
	// 先靜音 go-redis 內建 logger，避免 Ping 失敗時的重試訊息污染輸出
	redis.SetLogger(&noopLogger{})

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("WARNING: Redis unavailable — caching disabled")
		return nil
	}

	return client
}
