package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
)

var redisClient *redis.Client

func InitRedis(cfg *config.Config) error {
	logger := logging.GetLogger(cfg)
	client := redis.NewClient(&redis.Options{
		Addr:               fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:           cfg.Redis.Password,
		DB:                 cfg.Redis.Database,
		DialTimeout:        cfg.Redis.DialTimeout * time.Second,
		ReadTimeout:        cfg.Redis.ReadTimeout * time.Second,
		WriteTimeout:       cfg.Redis.WriteTimeout * time.Second,
		PoolSize:           cfg.Redis.PoolSize,
		PoolTimeout:        cfg.Redis.PoolTimeout * time.Second,
		IdleCheckFrequency: cfg.Redis.IdleCheckFrequency * time.Millisecond,
		IdleTimeout:        cfg.Redis.IdleTimeout * time.Second,
	})
	err := client.Ping().Err()
	if err != nil {
		return err
	}
	logger.Info(logging.Redis, logging.StartUp, "Redis connection established", nil)

	redisClient = client
	return nil
}

func CloseRedis(cfg *config.Config) {
	logger := logging.GetLogger(cfg)
	err := redisClient.Close()
	if err != nil {
		logger.Fatal(logging.Redis, logging.Close, err.Error(), nil)
	}
}

func GetRedis() *redis.Client {
	return redisClient
}

func Set[T any](client *redis.Client, key string, value T, duration time.Duration) (err error) {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return client.Set(key, v, duration).Err()
}

func Get[T any](client *redis.Client, key string) (T, error) {
	value := *new(T)
	res, err := client.Get(key).Result()
	if err != nil {
		return value, err
	}

	err = json.Unmarshal([]byte(res), &value)
	if err != nil {
		return value, err
	}
	return value, nil
}

func Del(client *redis.Client, key string) error {
	_, err := client.Del(key).Result()
	return err
}
