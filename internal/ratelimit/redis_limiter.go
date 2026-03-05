package ratelimiter

import (
	"context"
	_ "embed"
	"log/slog"
	"time"

	"github.com/go-redis/redis"
)

//go:embed gcra.lua
var gcraLua string

var luaScript = redis.NewScript(gcraLua)

type RedisLimiter struct {
	client *redis.Client
	rate   float64
	burst  float64
	ttl    time.Duration

	logger *slog.Logger
}

func NewRedis(client *redis.Client, rate float64, burst int, ttl time.Duration) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		rate:   rate,
		burst:  float64(burst),
		ttl:    ttl,
		logger: slog.Default(),
	}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string) Result {
	now := time.Now().UnixNano()

	res, err := luaScript.Run(
		l.client.WithContext(ctx),
		[]string{"ratelimit:" + key},
		l.rate,
		l.burst,
		now,
		l.ttl.Milliseconds(),
	).Result()
	if err != nil {
		l.logger.Error("redis limiter error", "err", err)

		return Result{
			Decision: DecisionAllow,
		}
	}

	values := res.([]interface{})
	allowed := values[0].(int64)

	if allowed == 0 {

		retryNs := values[1].(int64)

		return Result{
			Decision:   DecisionDeny,
			RetryAfter: time.Duration(retryNs),
			Reason:     ReasonExceeded,
		}
	}

	return Result{
		Decision: DecisionAllow,
	}
}
