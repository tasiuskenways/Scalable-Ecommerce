package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/config"
)

func NewRedisConnection(cfg *config.Config) (*redis.Client, error) {
	return NewRedisConnectionWithRetry(cfg, 5)
}

func NewRedisConnectionWithRetry(cfg *config.Config, maxRetries int) (*redis.Client, error) {
	for i := 0; i < maxRetries; i++ {
		client, err := connectToRedis(cfg)
		if err == nil {
			log.Println("Redis connected successfully")
			return client, nil
		}

		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * 2 * time.Second
			log.Printf("Failed to connect to Redis (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, waitTime)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts", maxRetries)
}

func connectToRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
