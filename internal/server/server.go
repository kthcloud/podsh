package server

import (
	"context"
	"errors"

	"github.com/kthcloud/podsh/internal/sshd"
)

type Server struct {
	sshServer *sshd.Server
	address   string
}

func New(opts ...Option) *Server {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	s := &Server{
		sshServer: sshd.New(sshd.WithConfig(cfg.SSHDConfig), sshd.WithHandler(cfg.Handler)),
		address:   cfg.Address,
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
