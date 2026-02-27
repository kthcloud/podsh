package metrics

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics interface {
	ListenAndServe(ctx context.Context, address string) error
	IncrCounter(name string)
	IncrGauge(name string)
	DecrGauge(name string)
}

type PrometheusMetrics struct {
	counters map[string]prometheus.Counter
	gauges   map[string]prometheus.Gauge
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		counters: make(map[string]prometheus.Counter),
		gauges:   make(map[string]prometheus.Gauge),
	}
}

// Start runs the HTTP server for Prometheus metrics and blocks
// until the context is canceled.
func (m *PrometheusMetrics) ListenAndServe(ctx context.Context, address string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/readyz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	}))
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	}))

	server := &http.Server{
		Addr:    address,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// IncrCounter increments a counter metric
func (m *PrometheusMetrics) IncrCounter(name string) {
	c, ok := m.counters[name]
	if !ok {
		c = prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: name + " counter",
		})
		prometheus.MustRegister(c)
		m.counters[name] = c
	}
	c.Inc()
}

// IncrGauge increments a gauge metric
func (m *PrometheusMetrics) IncrGauge(name string) {
	g, ok := m.gauges[name]
	if !ok {
		g = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: name + " gauge",
		})
		prometheus.MustRegister(g)
		m.gauges[name] = g
	}
	g.Inc()
}

// DecrGauge decrements a gauge metric
func (m *PrometheusMetrics) DecrGauge(name string) {
	g, ok := m.gauges[name]
	if !ok {
		g = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: name + " gauge",
		})
		prometheus.MustRegister(g)
		m.gauges[name] = g
	}
	g.Dec()
}
