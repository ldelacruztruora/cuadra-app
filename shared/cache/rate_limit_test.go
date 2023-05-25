package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultRate(t *testing.T) {
	c := require.New(t)

	limiterFallback = nil

	DefaultRate(context.Background(), time.Second, 1)
	c.NotNil(rateLimiter)
}

func TestIsRequestAllowed(t *testing.T) {
	c := require.New(t)

	_, _, allowed := RateLimit(context.Background(), "add-user", 10, time.Second)
	c.True(allowed)
}

func TestIsRequestNotAllowed(t *testing.T) {
	c := require.New(t)

	InitMock()

	_, _, allowed := RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.False(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.False(allowed)
}

func TestDefaultRateLimitFailure(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	DefaultRate(context.Background(), time.Second, 4)

	_, _, allowed := RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.True(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.False(allowed)

	_, _, allowed = RateLimit(context.Background(), "add-user", 4, time.Second)
	c.False(allowed)
}

func BenchmarkRateLimit(b *testing.B) {
	c := require.New(b)

	InitMock()

	times := int64(4)

	for i := int64(0); i < times; i++ {
		_, _, allowed := RateLimit(context.Background(), "add-user", times, time.Hour)
		c.True(allowed)
	}

	for i := 0; i < b.N; i++ {
		_, _, allowed := RateLimit(context.Background(), "add-user", times, time.Hour)
		c.False(allowed)
	}
}
