package algo

import (
	"context"
	"time"

	"distributed-rate-limiter/pkg/ratelimit"
	"distributed-rate-limiter/pkg/store"
)

// LeakyBucket implements a distributed leaky bucket algorithm
type LeakyBucket struct {
	store      store.Store
	rate       int64
	capacity   int64
	windowSize time.Duration
	keyPrefix  string
}

// NewLeakyBucket creates a new leaky bucket rate limiter
func NewLeakyBucket(store store.Store, opts ratelimit.Options) *LeakyBucket {
	return &LeakyBucket{
		store:      store,
		rate:       opts.Rate,
		capacity:   opts.Burst,
		windowSize: opts.WindowSize,
		keyPrefix:  opts.KeyPrefix,
	}
}

// Allow implements RateLimiter.Allow
func (lb *LeakyBucket) Allow(ctx context.Context, key string) (bool, error) {
	return lb.AllowN(ctx, key, 1)
}

// AllowN implements RateLimiter.AllowN
func (lb *LeakyBucket) AllowN(ctx context.Context, key string, n int64) (bool, error) {
	key = lb.keyPrefix + key

	// Get current queue size
	queueSize, err := lb.store.Increment(ctx, key, n, lb.windowSize)
	if err != nil {
		return false, err
	}

	// Check if we've exceeded capacity
	if queueSize > lb.capacity {
		// Rollback the increment since we're over capacity
		_, err := lb.store.Increment(ctx, key, -n, lb.windowSize)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	// Calculate maximum allowed requests based on rate
	maxRequests := lb.rate * int64(lb.windowSize.Seconds())
	if queueSize > maxRequests {
		// Rollback if we exceed the rate
		_, err := lb.store.Increment(ctx, key, -n, lb.windowSize)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

// Reset implements RateLimiter.Reset
func (lb *LeakyBucket) Reset(ctx context.Context, key string) error {
	return lb.store.Reset(ctx, lb.keyPrefix+key)
}
