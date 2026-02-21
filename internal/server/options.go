package server

import (
	"context"
	"log/slog"

	"github.com/kthcloud/podsh/internal/defaults"
	"github.com/kthcloud/podsh/internal/sshd"
)

type Config struct {
	Ctx context.Context

	Address string

	SSHDConfig sshd.Config
	Handler    sshd.SessionHandler

	Logger *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		Ctx: context.Background(),

		Address: defaults.DefaultBindAddress,

		SSHDConfig: sshd.DefaultConfig(),

		Logger: slog.New(slog.DiscardHandler),
	}
}

type Option func(cfg *Config)

func WithConfig(config Config) Option {
	return func(cfg *Config) {
		WithContext(config.Ctx)(cfg)
		WithLogger(config.Logger)(cfg)

		cfg.Address = config.Address
		cfg.SSHDConfig = config.SSHDConfig
		WithHandler(config.Handler)(cfg)
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

func WithHandler(handler sshd.SessionHandler) Option {
	if handler == nil {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.Handler = handler
	}
}

func WithAddress(address string) Option {
	return func(cfg *Config) {
		cfg.Address = address
	}
}
