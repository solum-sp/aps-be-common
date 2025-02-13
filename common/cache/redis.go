package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Service  string
}

type cacheRedis struct {
	redisClient *redis.Client
	service     string
}

var _ ICache = (*cacheRedis)(nil)

func NewRedisCache(config RedisConfig) (*cacheRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &cacheRedis{
		redisClient: client,
		service:     config.Service,
	}, nil
}

func (r *cacheRedis) Set(key string, value interface{}, expireTime *time.Duration) error {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	err := r.redisClient.Set(context.Background(), rKey, value, *expireTime).Err()
	return err
}

func (r *cacheRedis) Get(key string) (interface{}, error) {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	val, err := r.redisClient.Get(context.Background(), rKey).Result()
	return val, err
}

func (r *cacheRedis) GetWithPattern(pattern string) ([]string, error) {
	rKey := fmt.Sprintf("%s:%s", r.service, pattern)

	keys, err := r.redisClient.Keys(context.Background(), rKey).Result()
	return keys, err
}

func (r *cacheRedis) Delete(key string) error {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	err := r.redisClient.Del(context.Background(), rKey).Err()
	return err
}

func (r *cacheRedis) Clear() error {
	err := r.redisClient.FlushDB(context.Background()).Err()
	return err
}

func (r *cacheRedis) ClearWithPattern(pattern string) error {
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
