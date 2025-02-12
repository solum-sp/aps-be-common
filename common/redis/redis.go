package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IRedisCache interface {
	Set(key string, value interface{}, expireTime *time.Duration) error
	Get(key string) (interface{}, error)
	GetWithPattern(pattern string) ([]string, error)
	Delete(key string) error
	Clear() error
	ClearWithPattern(pattern string) error
}

type Config struct {
	Addr     string
	Password string
	DB       int
	Service  string
}

type appRedis struct {
	redisClient *redis.Client
	service     string
}

func NewRedisCache(config Config) (*appRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &appRedis{
		redisClient: client,
		service:     config.Service,
	}, nil
}

func (r *appRedis) Set(key string, value interface{}, expireTime *time.Duration) error {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	err := r.redisClient.Set(context.Background(), rKey, value, *expireTime).Err()
	return err
}

func (r *appRedis) Get(key string) (interface{}, error) {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	val, err := r.redisClient.Get(context.Background(), rKey).Result()
	return val, err
}

func (r *appRedis) GetWithPattern(pattern string) ([]string, error) {
	rKey := fmt.Sprintf("%s:%s", r.service, pattern)

	keys, err := r.redisClient.Keys(context.Background(), rKey).Result()
	return keys, err
}

func (r *appRedis) Delete(key string) error {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	err := r.redisClient.Del(context.Background(), rKey).Err()
	return err
}

func (r *appRedis) Clear() error {
	err := r.redisClient.FlushDB(context.Background()).Err()
	return err
}

func (r *appRedis) ClearWithPattern(pattern string) error {
	rKey := fmt.Sprintf("%s:%s", r.service, pattern)

	keys, err := r.redisClient.Keys(context.Background(), rKey).Result()
	if err != nil {
		return err
	}
	for _, key := range keys {
		err = r.redisClient.Del(context.Background(), key).Err()
		if err != nil {
			return err
		}
	}
	return nil
}
