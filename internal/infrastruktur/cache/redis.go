package cache

import (
	"context"
	"fmt"
	"postgresDB/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClinet create a new redis clinet
func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis:%w", err)
	}

	return client, nil
}
