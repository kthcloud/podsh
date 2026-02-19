package sshd

import (
	"log/slog"

	"golang.org/x/crypto/ssh"
)

type Option func(*Server)

func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

func WithHostSigner(signer ssh.Signer) Option {
	return func(s *Server) {
		s.hostSigner = signer
	}
}

func WithPublicKeyAuth(a PublicKeyAuthenticator) Option {
	return func(s *Server) {
		s.auth = a
	}
}
