package cache

import (
	"context"
	"fmt"
	"log"

	"api/infra/config"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Printf("Redis connection established: %s:%s", cfg.Redis.Host, cfg.Redis.Port)
	return rdb, nil
}

func CloseConnection(rdb *redis.Client) error {
	return rdb.Close()
}