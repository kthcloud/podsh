package sshd

import (
	"context"
	"log/slog"

	"golang.org/x/crypto/ssh"
)

func (s *Server) discardRequests(ctx context.Context, logger *slog.Logger, reqs <-chan *ssh.Request) {
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-reqs:
			if !ok {
				return
			}

			switch req.Type {
			case "keepalive@openssh.com":
				_ = req.Reply(true, nil)
			default:
				logger.Debug("ignoring global request", "type", req.Type)
				_ = req.Reply(false, nil)
			}
		}
	}
}
