package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/ethereum/go-ethereum/log"
	"github.com/multimarket-labs/event-pod-services/config"
	"github.com/redis/go-redis/v9"
)

// RedisCache Redis 缓存实现
type RedisCache struct {
	client *redis.Client
	locker *redislock.Client
	ctx    context.Context
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(cfg config.RedisConfig) (*RedisCache, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// 设置默认值
	poolSize := cfg.PoolSize
	if poolSize == 0 {
		poolSize = 10
	}
	minIdle := cfg.MinIdle
	if minIdle == 0 {
		minIdle = 5
	}

	// 设置超时时间默认值
	dialTimeout := cfg.DialTimeout
	if dialTimeout == 0 {
		dialTimeout = 5 * time.Second
	}
	readTimeout := cfg.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 3 * time.Second
	}
	writeTimeout := cfg.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = 3 * time.Second
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Info("Redis cache initialized successfully", "addr", cfg.Addr, "db", cfg.DB)

	return &RedisCache{
		client: rdb,
		locker: redislock.New(rdb),
		ctx:    ctx,
	}, nil
}

// Get 获取缓存值
func (r *RedisCache) Get(key string) (string, error) {
	if r == nil || r.client == nil {
		return "", nil
	}
	return r.client.Get(r.ctx, key).Result()
}

// Set 设置缓存值
func (r *RedisCache) Set(key string, value string, expiration time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Delete 删除缓存
func (r *RedisCache) Delete(key string) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Del(r.ctx, key).Err()
}

// Exists 检查 key 是否存在
func (r *RedisCache) Exists(key string) (bool, error) {
	if r == nil || r.client == nil {
		return false, nil
	}
	result, err := r.client.Exists(r.ctx, key).Result()
	return result > 0, err
}

// SetNX 设置 key-value，仅当 key 不存在时
func (r *RedisCache) SetNX(key string, value string, expiration time.Duration) (bool, error) {
	if r == nil || r.client == nil {
		return false, nil
	}
	return r.client.SetNX(r.ctx, key, value, expiration).Result()
}

// Incr 递增 key 的值
func (r *RedisCache) Incr(key string) (int64, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	return r.client.Incr(r.ctx, key).Result()
}

// IncrBy 按指定值递增 key
func (r *RedisCache) IncrBy(key string, value int64) (int64, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	return r.client.IncrBy(r.ctx, key, value).Result()
}

// HGet 获取哈希字段值
func (r *RedisCache) HGet(key string, field string) (string, error) {
	if r == nil || r.client == nil {
		return "", nil
	}
	return r.client.HGet(r.ctx, key, field).Result()
}

// HSet 设置哈希字段值
func (r *RedisCache) HSet(key string, field string, value string) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.HSet(r.ctx, key, field, value).Err()
}

// HIncrBy 递增哈希字段的值
func (r *RedisCache) HIncrBy(key string, field string, incr int64) (int64, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	return r.client.HIncrBy(r.ctx, key, field, incr).Result()
}

// HGetAll 获取哈希的所有字段和值
func (r *RedisCache) HGetAll(key string) (map[string]string, error) {
	if r == nil || r.client == nil {
		return make(map[string]string), nil
	}
	return r.client.HGetAll(r.ctx, key).Result()
}

// Expire 设置 key 的过期时间
func (r *RedisCache) Expire(key string, expiration time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// Close 关闭 Redis 连接
func (r *RedisCache) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Close()
}

// GetClient 获取底层 Redis 客户端（用于高级操作）
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}

// Ping 检查 Redis 连接
func (r *RedisCache) Ping(ctx context.Context) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("Redis client is nil")
	}
	return r.client.Ping(ctx).Err()
}

// ========== 分布式锁功能 ==========

// ObtainLock 获取分布式锁
// 使用 redislock 库提供的 Obtain 方法
func (r *RedisCache) ObtainLock(ctx context.Context, key string, ttl time.Duration) (*redislock.Lock, error) {
	if r == nil || r.locker == nil {
		return nil, fmt.Errorf("Redis locker is nil")
	}
	return r.locker.Obtain(ctx, key, ttl, nil)
}

// ObtainLockWithRetry 获取分布式锁（带重试）
func (r *RedisCache) ObtainLockWithRetry(ctx context.Context, key string, ttl time.Duration, retryDelay time.Duration) (*redislock.Lock, error) {
	if r == nil || r.locker == nil {
		return nil, fmt.Errorf("Redis locker is nil")
	}
	opts := &redislock.Options{
		RetryStrategy: redislock.LinearBackoff(retryDelay),
	}
	return r.locker.Obtain(ctx, key, ttl, opts)
}

// GetLocker 获取 redislock 客户端（用于高级用法）
func (r *RedisCache) GetLocker() *redislock.Client {
	return r.locker
}

// ========== Set 集合操作 ==========

// SAdd 向 Set 集合添加成员
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	return r.client.SAdd(ctx, key, members...).Result()
}

// SMembers 获取 Set 集合的所有成员
func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	if r == nil || r.client == nil {
		return []string{}, nil
	}
	return r.client.SMembers(ctx, key).Result()
}

// SRem 从 Set 集合移除成员
func (r *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	return r.client.SRem(ctx, key, members...).Result()
}

// SIsMember 检查成员是否在 Set 集合中
func (r *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	if r == nil || r.client == nil {
		return false, nil
	}
	return r.client.SIsMember(ctx, key, member).Result()
}

// ========== 限流器功能 ==========

// RateLimiter 限流器
type RateLimiter struct {
	cache  *RedisCache
	key    string
	limit  int64
	window time.Duration
}

// NewRateLimiter 创建限流器
func (r *RedisCache) NewRateLimiter(key string, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{
		cache:  r,
		key:    key,
		limit:  limit,
		window: window,
	}
}

// Allow 检查是否允许通过
func (rl *RateLimiter) Allow(ctx context.Context) (bool, error) {
	if rl.cache == nil || rl.cache.client == nil {
		return false, fmt.Errorf("Redis client is nil")
	}
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local current = redis.call("INCR", key)
		
		if current == 1 then
			redis.call("EXPIRE", key, window)
		end
		
		return current <= limit
	`
	result, err := rl.cache.client.Eval(ctx, script, []string{rl.key}, rl.limit, int(rl.window.Seconds())).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

// GetRemaining 获取剩余配额
func (rl *RateLimiter) GetRemaining(ctx context.Context) (int64, error) {
	if rl.cache == nil || rl.cache.client == nil {
		return 0, fmt.Errorf("Redis client is nil")
	}
	current, err := rl.cache.client.Get(ctx, rl.key).Int64()
	if err == redis.Nil {
		return rl.limit, nil
	}
	if err != nil {
		return 0, err
	}
	remaining := rl.limit - current
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}
