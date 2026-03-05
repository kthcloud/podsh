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
	ErrValidation               = errors.New("failed validation")
	ErrNoContext                = errors.Join(errors.New("context is nil"), ErrValidation)
	ErrNoLogger                 = errors.Join(errors.New("logger is nil"), ErrValidation)
	ErrNoHostSigner             = errors.Join(errors.New("no host signer"), ErrValidation)
	ErrNoSessionHandler         = errors.Join(errors.New("no session handler"), ErrValidation)
	ErrNoPublicKeyAuthenticator = errors.Join(errors.New("no publickey authenticator"), ErrValidation)
	ErrRateLimiterNoHasher      = errors.Join(errors.New("ratelimiter provided but no hasher was provided"), ErrValidation)
)

type Server struct {
	ctx    context.Context
	logger *slog.Logger

	hostSigner ssh.Signer
	auth       PublicKeyAuthenticator
	limiter    ratelimiter.Limiter
	hasher     ratelimiter.Hasher
	tarpit     *tarpit.Tarpit

	connector Connector

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
	if s.tarpit == nil {
		s.tarpit = tarpit.NewTarpit(s.ctx, 10)
	}

	if s.connector == nil {
		scfg := &ssh.ServerConfig{
			PublicKeyCallback: s.publicKeyCallback(s.ctx, s.logger),
			ServerVersion:     "SSH-2.0-podsh",
			BannerCallback: func(conn ssh.ConnMetadata) string {
				return `   __   __  __       __             __
  / /__/ /_/ /  ____/ /__  __ _____/ /
 /  '_/ __/ _ \/ __/ / _ \/ // / _  / 
/_/\_\\__/_//_/\__/_/\___/\_,_/\_,_/  
                                      
`
			},
		}
		scfg.AddHostKey(s.hostSigner)
		s.connector = NewConnectorImpl(s.ctx, s.logger, scfg, cfg.Handler2)
	}

	return s
}

func (s *Server) Validate() (err error) {
	if s.ctx == nil {
		err = errors.Join(ErrNoContext, err)
	}
	if s.logger == nil {
		err = errors.Join(ErrNoLogger, err)
	}
	if s.hostSigner == nil {
		err = errors.Join(ErrNoHostSigner, err)
	}
	if s.connector == nil {
		err = errors.Join(ErrNoSessionHandler, err)
	}
	if s.auth == nil {
		err = errors.Join(ErrNoPublicKeyAuthenticator, err)
	}
	if s.limiter != nil && s.hasher == nil {
		err = errors.Join(ErrRateLimiterNoHasher, err)
	}
	return
}
