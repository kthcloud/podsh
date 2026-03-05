package server

import (
	"context"
	"errors"
	"log/slog"

	register "github.com/kthcloud/podsh/internal/metrics"
	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/kthcloud/podsh/pkg/metrics"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	ctx            context.Context
	sshServer      *sshd.Server
	address        string
	metricsAddress string
	metrics        metrics.Metrics
	logger         *slog.Logger
}

func New(opts ...Option) *Server {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Metrics != nil {
		register.RegisterSSHdMetrics(cfg.Metrics)
	}

	s := &Server{
		ctx:            cfg.Ctx,
		sshServer:      sshd.New(sshd.WithConfig(cfg.SSHDConfig)),
		address:        cfg.Address,
		metricsAddress: cfg.MetricsAddress,
		metrics:        cfg.Metrics,
		logger:         cfg.Logger,
	}

	return s
}

func (s *Server) Validate() (err error) {
	// TODO: parse s.address to Validate  it
	if errs := s.sshServer.Validate(); errs != nil {
		err = errors.Join(errs, err)
	}
	return
}

func (s *Server) Start(ctx context.Context) error {
	var errg errgroup.Group
	if s.metrics != nil {
		ms := metrics.NewServer(
			metrics.WithMetrics(s.metrics),
			metrics.WithHealth(metrics.NewHealth()),
			metrics.WithLiveness(metrics.NewHealth()),
			metrics.WithReadiness(metrics.NewHealth()),
		)
		errg.Go(func() error {
			s.logger.Info("Metrics server started on", "address", "http://"+s.metricsAddress)
			if err := ms.ListenAndServe(ctx, s.metricsAddress); err != nil {
				s.logger.Error("Metrics server exited", "error", err)
				return err
			}
			return nil
		})
	}

	if err := s.sshServer.ListenAndServe(ctx, s.address); err != nil {
		return err
	}

	return errg.Wait()
}
