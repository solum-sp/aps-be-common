package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) (*cacheRedis, func()) {
	// Start a miniredis server
	mredis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mredis.Addr(),
	})

	cache := &cacheRedis{
		redisClient: client,
		service:     "test-service",
	}

	cleanup := func() {
		client.Close()
		mredis.Close()
	}

	return cache, cleanup
}

func TestSetAndGet(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	expireTime := 5 * time.Second
	err := cache.Set("test-key", "test-value", &expireTime)
	assert.NoError(t, err)

	val, err := cache.Get("test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", val)
}

func TestDelete(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	expireTime := 5 * time.Second
	cache.Set("test-key", "test-value", &expireTime)

	err := cache.Delete("test-key")
	assert.NoError(t, err)

	val, err := cache.Get("test-key")
	assert.Error(t, err)
	assert.Empty(t, val)
}

func TestClear(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	expireTime := 5 * time.Second
	cache.Set("key1", "value1", &expireTime)
	cache.Set("key2", "value2", &expireTime)

	err := cache.Clear()
	assert.NoError(t, err)

	val, err := cache.Get("key1")
	assert.Error(t, err)
	assert.Empty(t, val)
}

func TestGetWithPattern(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	expireTime := 5 * time.Second
	cache.Set("prefix:key1", "value1", &expireTime)
	cache.Set("prefix:key2", "value2", &expireTime)
	cache.Set("other:key3", "value3", &expireTime)

	keys, err := cache.GetWithPattern("prefix:*")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"test-service:prefix:key1", "test-service:prefix:key2"}, keys)
}

func TestClearWithPattern(t *testing.T) {
	cache, cleanup := setupTestRedis(t)
	defer cleanup()

	expireTime := 5 * time.Second
	cache.Set("prefix:key1", "value1", &expireTime)
	cache.Set("prefix:key2", "value2", &expireTime)
	cache.Set("other:key3", "value3", &expireTime)

	err := cache.ClearWithPattern("prefix:*")
	assert.NoError(t, err)

	val, err := cache.Get("prefix:key1")
	assert.Error(t, err)
	assert.Empty(t, val)

	val, err = cache.Get("prefix:key2")
	assert.Error(t, err)
	assert.Empty(t, val)

	val, err = cache.Get("other:key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)
}
