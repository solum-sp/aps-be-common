package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
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
	rsync       *redsync.Redsync
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

	pool := goredis.NewPool(client)
	rsync := redsync.New(pool)

	return &cacheRedis{
		redisClient: client,
		service:     config.Service,
		rsync:       rsync,
	}, nil
}

func (r *cacheRedis) GetClient() *redis.Client {
	return r.redisClient
}

func (r *cacheRedis) Close() error {
	return r.redisClient.Close()
}

func (r *cacheRedis) Set(key string, value interface{}, expireTime *time.Duration) error {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	if expireTime == nil {
		err := r.redisClient.Set(context.Background(), rKey, value, 0).Err()
		return err
	}
	err := r.redisClient.Set(context.Background(), rKey, value, *expireTime).Err()
	return err
}

func (r *cacheRedis) Get(key string) (interface{}, error) {
	rKey := fmt.Sprintf("%s:%s", r.service, key)
	val, err := r.redisClient.Get(context.Background(), rKey).Result()
	return val, err
}

func (r *cacheRedis) GetAll() ([]string, error) {
	keys, err := r.redisClient.Keys(context.Background(), fmt.Sprintf("%s:*", r.service)).Result()
	return keys, err
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

// Distributed lock
func (r *cacheRedis) Lock(key string, ttl time.Duration) (*redsync.Mutex, error) {
	mutex := r.rsync.NewMutex(fmt.Sprintf("%s:%s", r.service, key), redsync.WithExpiry(ttl))
	err := mutex.Lock()
	if err != nil {
		return nil, err
	}
	return mutex, nil
}

func (r *cacheRedis) Unlock(m *redsync.Mutex) error {
	_, err := m.Unlock()
	return err
}
