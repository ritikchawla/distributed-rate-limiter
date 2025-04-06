package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store defines the interface for rate limit storage operations
type Store interface {
	// Increment atomically increments the counter and returns the new value
	Increment(ctx context.Context, key string, value int64, windowSize time.Duration) (int64, error)

	// Get returns the current value for a key
	Get(ctx context.Context, key string) (int64, error)

	// Reset deletes the key and its associated data
	Reset(ctx context.Context, key string) error
}

// RedisStore implements Store using Redis
type RedisStore struct {
	client *redis.Client
	// Lua script for atomic increment operation
	incrementScript *redis.Script
}

// incrementLuaScript is an atomic increment operation that also handles TTL
const incrementLuaScript = `
local key = KEYS[1]
local value = tonumber(ARGV[1])
local windowSize = tonumber(ARGV[2])

local current = redis.call('INCRBY', key, value)
redis.call('EXPIRE', key, windowSize)

return current
`

// NewRedisStore creates a new Redis-backed store
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client:          client,
		incrementScript: redis.NewScript(incrementLuaScript),
	}
}

// Increment implements Store.Increment
func (s *RedisStore) Increment(ctx context.Context, key string, value int64, windowSize time.Duration) (int64, error) {
	result, err := s.incrementScript.Run(ctx, s.client, []string{key}, value, int(windowSize.Seconds())).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Get implements Store.Get
func (s *RedisStore) Get(ctx context.Context, key string) (int64, error) {
	val, err := s.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// Reset implements Store.Reset
func (s *RedisStore) Reset(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
