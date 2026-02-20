package sshd

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	ratelimiter "github.com/kthcloud/podsh/internal/ratelimit"
	"github.com/kthcloud/podsh/internal/tarpit"
	"golang.org/x/crypto/ssh"
)

var (
	ErrValidation       = errors.New("failed validation")
	ErrNoSessionHandler = errors.Join(errors.New("no session handler"), ErrValidation)
)

type Server struct {
	ctx    context.Context
	logger *slog.Logger

	hostSigner ssh.Signer
	auth       PublicKeyAuthenticator
	handler    SessionHandler
	limiter    ratelimiter.Limiter
	hasher     ratelimiter.Hasher
	tarpit     *tarpit.Tarpit

	mu sync.RWMutex
}

func New(opts ...Option) *Server {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	s := &Server{
		ctx:        cfg.Ctx,
		logger:     cfg.Logger,
		hostSigner: cfg.Signer,
		auth:       cfg.PublicKeyAuthenticator,
		limiter:    cfg.Limiter,
		hasher:     cfg.Hasher,
	}

	if s.limiter != nil && s.hasher == nil {
		s.hasher = ratelimiter.NewHasher([]byte("supersecret"))
	}
	if s.tarpit == nil {
		s.tarpit = tarpit.NewTarpit(context.Background(), 10)
	}

	return s
}

func (s *Server) Validate() (err error) {
	if s.handler == nil {
		err = errors.Join(ErrNoSessionHandler, err)
	}
	// TODO: validate all
	return
}
