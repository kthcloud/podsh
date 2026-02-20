package server

import (
	"context"
	"log/slog"

	"github.com/kthcloud/podsh/internal/sshd"
)

type Config struct {
	Ctx context.Context

	SSHDConfig sshd.Config

	Logger *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		Ctx: context.Background(),

		SSHDConfig: sshd.DefaultConfig(),

		Logger: slog.New(slog.DiscardHandler),
	}
}

type Option func(cfg *Config)

func WithConfig(config Config) Option {
	return func(cfg *Config) {
		WithContext(config.Ctx)(cfg)
		WithLogger(config.Logger)(cfg)

		cfg.SSHDConfig = config.SSHDConfig
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
