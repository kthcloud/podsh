package sshd

import (
	"context"
	"errors"
	"log/slog"

	"golang.org/x/crypto/ssh"
)

func (s *Server) publicKeyCallback(parent context.Context, logger *slog.Logger) func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	return func(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if s.auth == nil {
			return nil, errors.New("no authenticator configured")
		}

		id, err := s.auth.Authenticate(parent, ConnMetadata{
			User:       meta.User(),
			RemoteAddr: meta.RemoteAddr().String(),
		}, key.Marshal())
		if err != nil {
			logger.Debug("authentication failed",
				"user", meta.User(),
				"err", err,
			)
			return nil, err
		}

		perms := &ssh.Permissions{
			Extensions: map[string]string{
				"identity": encodeIdentity(id),
				// TODO: sanitize
				"requested-host": meta.User(),
			},
		}

		logger.Info("authentication success", "user", meta.User())
		return perms, nil
	}
}
