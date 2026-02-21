package server

import (
	"context"
	"errors"
	"log/slog"

	"github.com/kthcloud/podsh/internal/sshd"
)

type Server struct {
	ctx       context.Context
	sshServer *sshd.Server
	address   string
	logger    *slog.Logger
}

func New(opts ...Option) *Server {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	s := &Server{
		ctx:       cfg.Ctx,
		sshServer: sshd.New(sshd.WithConfig(cfg.SSHDConfig), sshd.WithHandler(cfg.Handler)),
		address:   cfg.Address,
		logger:    cfg.Logger,
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
	if err := s.sshServer.ListenAndServe(ctx, s.address); err != nil {
		return err
	}
	return nil
}
