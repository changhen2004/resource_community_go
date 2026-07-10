package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis(ctx context.Context, cfg *Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Println("redis initialized")
	return client, nil
}
