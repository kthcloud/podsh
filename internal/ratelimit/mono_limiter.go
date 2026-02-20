package ratelimiter

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

type MonoLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   float64
	ttl     time.Duration

	logger *slog.Logger
}

func New(rate float64, burst int, ttl time.Duration) *MonoLimiter {
	return &MonoLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   float64(burst),
		ttl:     ttl,

		logger: slog.Default(),
	}
}

func (l *MonoLimiter) Allow(ctx context.Context, key string) Result {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}

	// TTL eviction
	if now.Sub(b.last) > l.ttl {
		b.tokens = l.burst
		b.last = now
	}

	// refill only when meaningful time passed
	elapsed := now.Sub(b.last)
	tokensToAdd := elapsed.Seconds() * l.rate
	if tokensToAdd >= 1 {
		b.tokens += tokensToAdd
		if b.tokens > l.burst {
			b.tokens = l.burst
		}
		consumed := time.Duration(tokensToAdd / l.rate * float64(time.Second))
		b.last = b.last.Add(consumed)
	}

	// consume atomically
	newTokens := b.tokens - 1
	if newTokens < 0 {
		retryAfter := time.Duration((-newTokens) / l.rate * float64(time.Second))
		l.logger.Debug("ratelimited", "retryAfter", retryAfter, "key", key)
		return Result{
			Decision:   DecisionDeny,
			RetryAfter: retryAfter,
			Reason:     ReasonExceeded,
		}
	}

	b.tokens = newTokens

	l.logger.Debug("limiter allowed", "tokens", b.tokens, "key", key)
	return Result{Decision: DecisionAllow}
}
