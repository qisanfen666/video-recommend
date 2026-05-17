package cache

import (
	"context"
	"fmt"
	"log"
	"video-recommend/config"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("[Redis] Connection established successfully")
	return nil
}

func CloseRedis() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}
