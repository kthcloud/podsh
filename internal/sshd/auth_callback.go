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

		// TODO: sanitize
		deployment := meta.User()

		id, err := s.auth.Authenticate(parent, ConnMetadata{
			Deployment: deployment,
			RemoteAddr: meta.RemoteAddr().String(),
		}, key.Marshal())
		if err != nil {
			logger.Debug("authentication failed",
				"deployment", deployment,
				"remoteAddr", meta.RemoteAddr().String(),
				"err", err,
			)
			return nil, err
		}

		identity := encodeIdentity(id)
		perms := &ssh.Permissions{
			Extensions: map[string]string{
				"identity":       identity,
				"requested-host": deployment,
			},
		}

		logger.Info("authentication success", "deployment", deployment, "user", id.User, "userID", id.UserID, "remoteAddr", id.RemoteAddr)
		return perms, nil
	}
}
