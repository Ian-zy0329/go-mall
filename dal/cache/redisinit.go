package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/Ian-zy0329/go-mall/config"
	"github.com/redis/go-redis/v9"
	"time"
)

var redisClient *redis.Client
var redisStockService *redis.Client

func Redis() *redis.Client {
	return redisClient
}

func RedisStockService() *redis.Client {
	return redisStockService
}

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:         config.Redis.Addr,
		Password:     config.Redis.Password,
		DB:           config.Redis.DB,
		PoolSize:     config.Redis.PoolSize,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolTimeout:  30 * time.Second,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	redisStockService = redis.NewClient(&redis.Options{
		Addr:         config.RedisStockServiceConfig.Addr,
		DB:           config.RedisStockServiceConfig.DB,
		PoolSize:     config.RedisStockServiceConfig.PoolSize,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolTimeout:  30 * time.Second,
	})
	if err := redisStockService.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func acquireLock(ctx context.Context, redisClient *redis.Client, key string, expire time.Duration) (string, error) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	ok, err := redisClient.SetNX(ctx, key, token, expire).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("lock acquisition failed")
	}
	return token, nil
}

func releaseLock(ctx context.Context, redisClient *redis.Client, key, token string) error {
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)

	_, err := script.Run(ctx, redisClient, []string{key}, token).Result()
	return err
}
