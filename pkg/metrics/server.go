package metrics

import (
	"context"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

type ServerOptions struct {
	Addr      string
	Metrics   Metrics
	Health    *Health
	Readiness *Health
	Liveness  *Health
}

func NewServer(opts ServerOptions) *Server {
	mux := http.NewServeMux()

	if opts.Metrics != nil {
		mux.Handle("/metrics", opts.Metrics.Handler())
	}

	if opts.Health != nil {
		mux.HandleFunc("/healthz", opts.Health.Handler)
	}

	if opts.Readiness != nil {
		mux.HandleFunc("/readyz", opts.Readiness.Handler)
	}

	if opts.Liveness != nil {
		mux.HandleFunc("/livez", opts.Liveness.Handler)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    opts.Addr,
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
