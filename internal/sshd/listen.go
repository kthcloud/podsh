package sshd

import (
	"context"
	"net"
)

// FIXME: respect context
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.logger.Info("ssh server listening", "addr", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			s.logger.Error("accept failed", "err", err)
			continue
		}

		go s.handleConn(ctx, conn)
	}
}
