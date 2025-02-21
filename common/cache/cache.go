package cache

import "time"

type ICache interface {
	Set(key string, value interface{}, expireTime *time.Duration) error
	Get(key string) (interface{}, error)
	GetAll() ([]string, error)
	GetWithPattern(pattern string) ([]string, error)
	Delete(key string) error
	Clear() error
	ClearWithPattern(pattern string) error
}
