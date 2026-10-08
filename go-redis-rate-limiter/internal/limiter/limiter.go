package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Result struct {
	Allowed   bool
	Limit     int
	Remaining int
	Reset     time.Duration
}

type RateLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func New(redisClient *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		limit:  limit,
		window: window,
	}
}

func (l *RateLimiter) Allow(ctx context.Context, clientID, algorithm string) (Result, error) {
	switch algorithm {
	case "sliding":
		return l.allowSliding(ctx, clientID)
	default:
		return l.allowFixed(ctx, clientID)
	}
}
