package sshd

import (
	"context"
	"net"
	"sync"

	ratelimiter "github.com/kthcloud/podsh/internal/ratelimit"
	"golang.org/x/sync/errgroup"
)

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	var closeOnce sync.Once
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer closeOnce.Do(func() {
		ln.Close()
	})

	done := make(chan struct{})
	var eg errgroup.Group
	eg.Go(func() error {
		defer close(done)
		for {
			conn, err := ln.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				s.logger.Error("accept failed", "err", err)
				continue
			}

			// TODO: rate limiting
			if s.limiter != nil {
				key, ok := s.hasher.FromConn(conn)
				if !ok {
					conn.Close()
					continue
				}
				res := s.limiter.Allow(ctx, key)
				if res.Decision == ratelimiter.DecisionDeny {
					if res.RetryAfter > 0 {
						// tell them theyre rate ratelimited
						s.tarpit.Add(conn, res.RetryAfter)
					} else {
						conn.Close()
					}
					continue
				}

			}

			go s.handleConn(ctx, conn)
		}
	})

	s.logger.Info("ssh server listening", "addr", addr)

	select {
	case <-done:
	case <-ctx.Done():
		closeOnce.Do(func() {
			ln.Close()
		})
	}

	return eg.Wait()
}
