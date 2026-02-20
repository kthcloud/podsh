package server

import (
	"context"
	"errors"

	"github.com/kthcloud/podsh/internal/sshd"
)

type Server struct {
	sshServer *sshd.Server
}

func New(opts ...Option) *Server {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	s := &Server{
		sshServer: sshd.New(sshd.WithConfig(cfg.SSHDConfig)),
	}

	return s
}

func (s *Server) Validate() (err error) {
	if errs := s.sshServer.Validate(); errs != nil {
		err = errors.Join(errs, err)
	}
	return
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.sshServer.ListenAndServe(ctx, "localhost:2222"); err != nil {
		return err
	}
	return nil
}
