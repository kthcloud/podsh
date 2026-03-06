package sshd

import (
	"context"
	"log/slog"

	ratelimiter "github.com/kthcloud/podsh/internal/ratelimit"
	"github.com/kthcloud/podsh/pkg/metrics"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Ctx context.Context

	Signer                 ssh.Signer
	PublicKeyAuthenticator PublicKeyAuthenticator
	Limiter                ratelimiter.Limiter
	Hasher                 ratelimiter.Hasher
	Metrics                metrics.Metrics

	Handler2 Handler

	Logger *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		Ctx: context.Background(),

		Metrics: metrics.NewNoop(),

		Logger: slog.New(slog.DiscardHandler),
	}
}

type Option func(*Config)

func WithConfig(config Config) Option {
	return func(cfg *Config) {
		WithContext(config.Ctx)(cfg)
		WithLogger(config.Logger)(cfg)

		cfg.Signer = config.Signer
		cfg.PublicKeyAuthenticator = config.PublicKeyAuthenticator
		cfg.Limiter = config.Limiter
		cfg.Hasher = config.Hasher
		WithMetrics(config.Metrics)(cfg)

		cfg.Handler2 = config.Handler2
	}
}

func WithContext(ctx context.Context) Option {
	if ctx == nil {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.Ctx = ctx
	}
}

func WithLogger(logger *slog.Logger) Option {
	if logger == nil {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}

func WithHostSigner(signer ssh.Signer) Option {
	return func(cfg *Config) {
		cfg.Signer = signer
	}
}

func WithPublicKeyAuth(a PublicKeyAuthenticator) Option {
	return func(cfg *Config) {
		cfg.PublicKeyAuthenticator = a
	}
}

func WithLimiter(limiter ratelimiter.Limiter) Option {
	if limiter == nil {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.Limiter = limiter
	}
}

func WithHasher(hasher ratelimiter.Hasher) Option {
	if hasher == nil {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.Hasher = hasher
	}
}

func WithMetrics(metrics metrics.Metrics) Option {
	if metrics == nil {
		return func(_ *Config) {}
	}
	return func(c *Config) {
		c.Metrics = metrics
	}
}
