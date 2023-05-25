package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v9"
	"golang.org/x/time/rate"
)

var (
	limiterFallback *rate.Limiter
)

// DefaultRate sets a default rate when redis is down
func DefaultRate(ctx context.Context, period time.Duration, max int) {
	limiterFallback = rate.NewLimiter(rate.Every(period), max)
}

// RateLimit counts the request and returns whether the request is allowed or not
// Usage: RateLimit(context.Background(), "request1", 10, time.Second)
func RateLimit(ctx context.Context, name string, max int64, period time.Duration) (count int64, delay time.Duration, allow bool) {
	res, err := rateLimiter.Allow(ctx, name, redis_rate.Limit{
		Rate:   int(max),
		Burst:  int(max),
		Period: period,
	})
	if err == nil {
		return max - int64(res.Remaining), res.RetryAfter, res.Allowed != 0
	}

	allowed := limiterFallback.Allow()

	return 1, 1 * time.Second, allowed
}
