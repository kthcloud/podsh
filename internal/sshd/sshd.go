package sshd

import (
	"log/slog"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	logger *slog.Logger

	hostSigner ssh.Signer
	auth       PublicKeyAuthenticator
	handler    SessionHandler

	mu sync.RWMutex
}

func New(opts ...Option) *Server {
	s := &Server{
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) HandleSession(h SessionHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handler = h
}
