package sshd

import (
	"context"
	"net"
	"sync"
	"time"

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
		_ = ln.Close()
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

			if s.limiter != nil {
				key, ok := s.hasher.FromConn(conn)
				if !ok {
					s.logger.Warn("gailed to hash connection")
					_ = conn.Close()
					continue
				}
				res := s.limiter.Allow(ctx, key)
				if res.Decision == ratelimiter.DecisionDeny {
					if res.RetryAfter > 0 {
						s.logger.Debug("tarpitted ratelimited client", "for", res.RetryAfter)
						s.tarpit.Add(conn, res.RetryAfter)
					} else {
						_ = conn.Close()
					}
					continue
				}

			}

			backoff := 1 * time.Millisecond
			for !eg.TryGo(func() error { return s.connector.Handle(conn) }) {
				s.logger.Error("Used up all goroutines and failed to handle new connection, sleeping for", "duration", backoff)
				select {
				case <-time.After(backoff):
					backoff = time.Duration(max(backoff.Milliseconds()*2, 500)) * time.Millisecond
				case <-s.ctx.Done():
					return ctx.Err()
				}
			}

			// go s.handleConn(ctx, conn)
		}
	})

	s.logger.Info("ssh server listening", "addr", "ssh://"+addr)

	select {
	case <-done:
	case <-ctx.Done():
		closeOnce.Do(func() {
			_ = ln.Close()
		})
	}

	return eg.Wait()
}
