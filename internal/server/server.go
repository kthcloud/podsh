package server

import (
	"context"
	"errors"
	"log/slog"

	"github.com/kthcloud/podsh/internal/metrics"
	"github.com/kthcloud/podsh/internal/sshd"
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

	s := &Server{
		ctx:            cfg.Ctx,
		sshServer:      sshd.New(sshd.WithConfig(cfg.SSHDConfig), sshd.WithHandler(cfg.Handler)),
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
		errg.Go(func() error {
			s.logger.Info("Metrics server started on", "address", "http://"+s.metricsAddress)
			if err := s.metrics.ListenAndServe(ctx, s.metricsAddress); err != nil {
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
