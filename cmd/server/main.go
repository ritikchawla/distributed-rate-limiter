package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"distributed-rate-limiter/internal/algo"
	"distributed-rate-limiter/pkg/ratelimit"
	"distributed-rate-limiter/pkg/store"
)

var (
	addr       = flag.String("addr", ":8080", "HTTP server address")
	redisAddr  = flag.String("redis", "localhost:6379", "Redis server address")
	rateLimit  = flag.Int64("rate", 10, "Number of requests per second")
	burstLimit = flag.Int64("burst", 20, "Maximum burst size")
)

func main() {
	flag.Parse()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create Redis store
	redisStore := store.NewRedisStore(rdb)

	// Configure rate limiters
	opts := ratelimit.Options{
		Rate:       *rateLimit,
		Burst:      *burstLimit,
		WindowSize: time.Second,
		KeyPrefix:  "ratelimit:",
	}

	tokenBucket := algo.NewTokenBucket(redisStore, opts)
	leakyBucket := algo.NewLeakyBucket(redisStore, opts)

	// Set up HTTP handlers
	http.HandleFunc("/token", createLimiterHandler(tokenBucket))
	http.HandleFunc("/leaky", createLimiterHandler(leakyBucket))

	// Start server
	log.Printf("Starting server on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func createLimiterHandler(limiter ratelimit.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr

		allowed, err := limiter.Allow(r.Context(), clientIP)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Rate limiting error: %v", err)
			return
		}

		response := map[string]interface{}{
			"allowed": allowed,
			"ip":      clientIP,
		}

		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			response["message"] = "Rate limit exceeded"
		} else {
			response["message"] = "Request allowed"
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}
