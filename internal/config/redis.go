package config

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *RedisConfig) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("unable to reach redis: %w", err)
    }

    return client, nil
}
