package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type promMetrics struct {
	registry *prometheus.Registry

	counters   map[string]prometheus.Counter
	gauges     map[string]prometheus.Gauge
	histograms map[string]prometheus.Histogram

	mu sync.RWMutex
}

func NewPrometheus() Metrics {
	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
	)
	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &promMetrics{
		registry:   reg,
		counters:   make(map[string]prometheus.Counter),
		gauges:     make(map[string]prometheus.Gauge),
		histograms: make(map[string]prometheus.Histogram),
	}
}

func (m *promMetrics) RegisterCounter(name, help string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.counters[name]; exists {
		return fmt.Errorf("counter exists: %s", name)
	}

	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})

	if err := m.registry.Register(c); err != nil {
		return err
	}

	m.counters[name] = c
	return nil
}

func (m *promMetrics) RegisterGauge(name, help string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.gauges[name]; exists {
		return fmt.Errorf("gauge exists: %s", name)
	}

	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})

	if err := m.registry.Register(g); err != nil {
		return err
	}

	m.gauges[name] = g
	return nil
}

func (m *promMetrics) RegisterHistogram(name, help string, buckets []float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.histograms[name]; exists {
		return fmt.Errorf("histogram exists: %s", name)
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	h := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	})

	if err := m.registry.Register(h); err != nil {
		return err
	}

	m.histograms[name] = h
	return nil
}

type promCounter struct {
	c prometheus.Counter
}

func (p *promCounter) Inc()          { p.c.Inc() }
func (p *promCounter) Add(v float64) { p.c.Add(v) }

type promGauge struct {
	g prometheus.Gauge
}

func (p *promGauge) Inc()          { p.g.Inc() }
func (p *promGauge) Dec()          { p.g.Dec() }
func (p *promGauge) Set(v float64) { p.g.Set(v) }
func (p *promGauge) Add(v float64) { p.g.Add(v) }

type promHistogram struct {
	h prometheus.Histogram
}

func (p *promHistogram) Observe(v float64) { p.h.Observe(v) }

func (m *promMetrics) Counter(name string) Counter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.counters[name]
	if !ok {
		panic("counter not found: " + name)
	}
	return &promCounter{c: c}
}

func (m *promMetrics) Gauge(name string) Gauge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g, ok := m.gauges[name]
	if !ok {
		panic("gauge not found: " + name)
	}
	return &promGauge{g: g}
}

func (m *promMetrics) Histogram(name string) Histogram {
	m.mu.RLock()
	defer m.mu.RUnlock()

	h, ok := m.histograms[name]
	if !ok {
		panic("histogram not found: " + name)
	}
	return &promHistogram{h: h}
}

func (m *promMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
