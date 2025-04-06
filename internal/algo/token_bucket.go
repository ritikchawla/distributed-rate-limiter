package algo

import (
	"context"
	"time"

	"distributed-rate-limiter/pkg/ratelimit"
	"distributed-rate-limiter/pkg/store"
)

// TokenBucket implements a distributed token bucket algorithm
type TokenBucket struct {
	store      store.Store
	rate       int64
	burst      int64
	windowSize time.Duration
	keyPrefix  string
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(store store.Store, opts ratelimit.Options) *TokenBucket {
	return &TokenBucket{
		store:      store,
		rate:       opts.Rate,
		burst:      opts.Burst,
		windowSize: opts.WindowSize,
		keyPrefix:  opts.KeyPrefix,
	}
}

// Allow implements RateLimiter.Allow
func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	return tb.AllowN(ctx, key, 1)
}

// AllowN implements RateLimiter.AllowN
func (tb *TokenBucket) AllowN(ctx context.Context, key string, n int64) (bool, error) {
	key = tb.keyPrefix + key

	// Get current token count
	tokens, err := tb.store.Increment(ctx, key, n, tb.windowSize)
	if err != nil {
		return false, err
	}

	// Check if we have exceeded our burst limit
	if tokens > tb.burst {
		// Rollback the increment since we're over limit
		_, err := tb.store.Increment(ctx, key, -n, tb.windowSize)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	// We're within our limits
	return true, nil
}

// Reset implements RateLimiter.Reset
func (tb *TokenBucket) Reset(ctx context.Context, key string) error {
	return tb.store.Reset(ctx, tb.keyPrefix+key)
}
