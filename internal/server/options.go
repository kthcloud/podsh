package server

import (
	"context"
	"log/slog"

	"github.com/kthcloud/podsh/internal/defaults"
	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/kthcloud/podsh/pkg/metrics"
)

type Config struct {
	Ctx context.Context

	Address        string
	MetricsAddress string

	SSHDConfig sshd.Config
	Metrics    metrics.Metrics

	Logger *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		Ctx: context.Background(),

		Address:        defaults.DefaultBindAddress,
		MetricsAddress: defaults.DefaultMetricsAddr,

		SSHDConfig: sshd.DefaultConfig(),
		Metrics:    metrics.NewPrometheus(),

		Logger: slog.New(slog.DiscardHandler),
	}
}

type Option func(cfg *Config)

func WithConfig(config Config) Option {
	return func(cfg *Config) {
		WithContext(config.Ctx)(cfg)
		WithLogger(config.Logger)(cfg)

		WithAddress(config.Address)(cfg)
		WithMetricsAddress(config.MetricsAddress)(cfg)

		cfg.SSHDConfig = config.SSHDConfig
		WithMetrics(config.Metrics)(cfg)
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

func WithAddress(address string) Option {
	if address == "" {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.Address = address
	}
}

func WithMetricsAddress(address string) Option {
	if address == "" {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.MetricsAddress = address
	}
}

func WithMetrics(metrics metrics.Metrics) Option {
	if metrics == nil {
		return func(_ *Config) {}
	}
	return func(cfg *Config) {
		cfg.Metrics = metrics
	}
}
