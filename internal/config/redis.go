package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

// InitRedis initializes Redis connection with connection pooling and health check
func InitRedis() {
	config := AppConfig.Redis

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.PoolSize / 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	if err := testRedisConnection(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Redis connected successfully with pool size %d", config.PoolSize)
}

// testRedisConnection tests the Redis connection
func testRedisConnection() error {
	ctx, cancel := context.WithTimeout(Ctx, 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	return err
}

// GetRedisClient returns the Redis client
func GetRedisClient() *redis.Client {
	return RedisClient
}

// HealthCheckRedis performs a health check on Redis
func HealthCheckRedis() error {
	ctx, cancel := context.WithTimeout(Ctx, 2*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	return err
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	return RedisClient.Close()
}
