package ratelimit

import (
	"context"
	"time"
)

// RateLimiter defines the interface for rate limiting operations
type RateLimiter interface {
	// Allow checks if a request should be allowed based on the key
	Allow(ctx context.Context, key string) (bool, error)

	// AllowN checks if N requests should be allowed based on the key
	AllowN(ctx context.Context, key string, n int64) (bool, error)

	// Reset resets the rate limiter for the given key
	Reset(ctx context.Context, key string) error
}

// Options contains configuration for rate limiters
type Options struct {
	// Rate is the number of requests per second
	Rate int64

	// Burst is the maximum number of requests allowed to exceed the rate
	Burst int64

	// WindowSize is the time window for rate limiting
	WindowSize time.Duration

	// KeyPrefix is used to namespace rate limit keys in Redis
	KeyPrefix string
}
