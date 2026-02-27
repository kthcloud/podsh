package ratelimiter

import (
	"context"
	"time"
)

type Decision int

const (
	DecisionAllow Decision = iota
	DecisionDeny
)

type Reason string

const (
	ReasonExceeded = "rate_exceeded"
	ReasonBanned   = "banned"
	ReasonError    = "error"
)

type Result struct {
	Decision   Decision
	RetryAfter time.Duration // 0 if unknown
	Reason     Reason
}

type Limiter interface {
	// Key is already normalized (ex: hashed IP)
	Allow(ctx context.Context, key string) Result
}

func Allowed(l Limiter, ctx context.Context, key string) bool {
	return l.Allow(ctx, key).Decision == DecisionAllow
}
