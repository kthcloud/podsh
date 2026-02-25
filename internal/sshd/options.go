package sshd

import (
	"context"
	"log/slog"

	ratelimiter "github.com/kthcloud/podsh/internal/ratelimit"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Ctx context.Context

	Signer                 ssh.Signer
	PublicKeyAuthenticator PublicKeyAuthenticator
	Limiter                ratelimiter.Limiter
	Hasher                 ratelimiter.Hasher

	Handler  SessionHandler
	Handler2 Handler

	Logger *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		Ctx: context.Background(),

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

		cfg.Handler = config.Handler
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
	return func(cfg *Config) {
		cfg.Limiter = limiter
	}
}

func WithHasher(hasher ratelimiter.Hasher) Option {
	return func(cfg *Config) {
		cfg.Hasher = hasher
	}
}

func WithHandler(handler SessionHandler) Option {
	return func(cfg *Config) {
		cfg.Handler = handler
	}
}
