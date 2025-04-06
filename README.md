# Distributed Rate Limiter

A distributed rate limiting service using Redis as a backing store, implementing both Token Bucket and Leaky Bucket algorithms.

## Features

- Distributed rate limiting using Redis
- Token Bucket algorithm implementation
- Leaky Bucket algorithm implementation
- HTTP server for testing rate limits
- Thread-safe and concurrent access support
- Configurable rates and burst limits

## Architecture

The service is built with the following components:

- **Store Package**: Handles Redis operations and Lua scripts for atomic operations
- **Rate Limit Package**: Defines core interfaces and configurations
- **Algorithms**: Implements Token Bucket and Leaky Bucket algorithms
- **HTTP Server**: Provides REST endpoints to test rate limiting

## Prerequisites

- Go 1.19 or later
- Redis 6.x or later

## Installation

```bash
git clone https://github.com/yourusername/distributed-rate-limiter.git
cd distributed-rate-limiter
go mod download
```

## Running the Server

1. Start Redis server:
```bash
redis-server
```

2. Run the rate limiter server:
```bash
go run cmd/server/main.go
```

Optional flags:
- `-addr`: HTTP server address (default: ":8080")
- `-redis`: Redis server address (default: "localhost:6379")
- `-rate`: Requests per second (default: 10)
- `-burst`: Maximum burst size (default: 20)

## Usage

The server exposes two endpoints:

1. Token Bucket Rate Limiter:
```bash
curl http://localhost:8080/token
```

2. Leaky Bucket Rate Limiter:
```bash
curl http://localhost:8080/leaky
```

Example Response:
```json
{
    "allowed": true,
    "ip": "127.0.0.1",
    "message": "Request allowed"
}
```

When rate limit is exceeded:
```json
{
    "allowed": false,
    "ip": "127.0.0.1",
    "message": "Rate limit exceeded"
}
```

## Implementation Details

### Token Bucket Algorithm
- Maintains a bucket of tokens that refills at a fixed rate
- Each request consumes one token
- Requests are allowed if tokens are available
- Supports burst traffic up to the bucket capacity

### Leaky Bucket Algorithm
- Models a fixed capacity bucket with a constant outflow rate
- Requests are added to the bucket if capacity allows
- Maintains a consistent outflow rate
- Helps smooth out traffic spikes

## Redis Implementation

The service uses Redis with Lua scripts for atomic operations to ensure consistency in a distributed environment. The Redis store handles:

- Atomic increments with TTL
- Key expiration for sliding windows
- Distributed state management

## Testing

To be Added