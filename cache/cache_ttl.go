package cache

import (
	"errors"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type CacheTTL struct {
	cache *ristretto.Cache[string, any]
}

// NewCache 初始化缓存
// numCounters: 需要计数的 key 数量（10x capacity 作为推荐值）
// maxCost: 最大缓存大小（可以用字节数，也可以用抽象“成本”）
// bufferItems: 并发写缓冲大小（推荐 64 或 128）
func NewCache(numCounters int64, maxCost int64, bufferItems int64) (*CacheTTL, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
	})
	if err != nil {
		return nil, err
	}
	return &CacheTTL{cache: cache}, nil
}

// Set ttl = 0 表示永不过期（但仍可能因容量淘汰）
func (c *CacheTTL) Set(key string, value any, cost int64, ttl time.Duration) {
	if ttl > 0 {
		c.cache.SetWithTTL(key, value, cost, ttl)
	} else {
		c.cache.Set(key, value, cost)
	}
	c.cache.Wait()
}

func (c *CacheTTL) Get(key string) (any, error) {
	value, found := c.cache.Get(key)
	if !found {
		return nil, errors.New("cache miss")
	}
	return value, nil
}

func (c *CacheTTL) Update(key string, value any, cost int64, ttl time.Duration) error {
	_, found := c.cache.Get(key)
	if !found {
		return errors.New("cache miss, cannot update")
	}
	c.Set(key, value, cost, ttl)
	return nil
}

func (c *CacheTTL) Delete(key string) {
	c.cache.Del(key)
}

func (c *CacheTTL) Close() {
	c.cache.Close()
}

var globalCache *CacheTTL

func Init(numCounters int64, maxCost int64, bufferItems int64) error {
	var err error
	globalCache, err = NewCache(numCounters, maxCost, bufferItems)
	return err
}

func Get(ctx interface{}, key string) (string, error) {
	if globalCache == nil {
		return "", errors.New("cache not initialized")
	}
	val, err := globalCache.Get(key)
	if err != nil {
		return "", err
	}
	if str, ok := val.(string); ok {
		return str, nil
	}
	return "", errors.New("value is not a string")
}

func Set(ctx interface{}, key string, value string, ttl time.Duration) error {
	if globalCache == nil {
		return errors.New("cache not initialized")
	}
	globalCache.Set(key, value, 1, ttl)
	return nil
}

func Delete(ctx interface{}, key string) {
	if globalCache != nil {
		globalCache.Delete(key)
	}
}
