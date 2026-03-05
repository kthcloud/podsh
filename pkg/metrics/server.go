package metrics

import (
	"context"
	"net/http"
)

type Server interface {
	ListenAndServe(ctx context.Context, addr string) error
}

type Option func(*ServerImpl)

func WithMetrics(m Metrics) Option {
	return func(s *ServerImpl) {
		if m != nil {
			s.mux.Handle("/metrics", m.Handler())
		}
	}
}

func WithHealth(h *Health) Option {
	return func(s *ServerImpl) {
		if h != nil {
			s.mux.HandleFunc("/healthz", h.Handler)
		}
	}
}

func WithReadiness(h *Health) Option {
	return func(s *ServerImpl) {
		if h != nil {
			s.mux.HandleFunc("/readyz", h.Handler)
		}
	}
}

func WithLiveness(h *Health) Option {
	return func(s *ServerImpl) {
		if h != nil {
			s.mux.HandleFunc("/livez", h.Handler)
		}
	}
}

type ServerImpl struct {
	httpServer *http.Server
	mux        *http.ServeMux
}

// Constructor
func NewServer(opts ...Option) Server {
	mux := http.NewServeMux()

	s := &ServerImpl{
		mux: mux,
		httpServer: &http.Server{
			Handler: mux,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *ServerImpl) ListenAndServe(ctx context.Context, addr string) error {
	s.httpServer.Addr = addr

	errCh := make(chan error, 1)

	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
