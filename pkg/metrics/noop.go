package metrics

import "net/http"

type noopMetrics struct{}

func NewNoop() Metrics {
	return &noopMetrics{}
}

func (n *noopMetrics) RegisterCounter(name, help string) error { return nil }
func (n *noopMetrics) RegisterGauge(name, help string) error   { return nil }
func (n *noopMetrics) RegisterHistogram(name, help string, buckets []float64) error {
	return nil
}

func (n *noopMetrics) Counter(name string) Counter     { return noopCounter{} }
func (n *noopMetrics) Gauge(name string) Gauge         { return noopGauge{} }
func (n *noopMetrics) Histogram(name string) Histogram { return noopHistogram{} }

func (n *noopMetrics) Handler() http.Handler {
	return http.NotFoundHandler()
}

type noopCounter struct{}

func (noopCounter) Inc()        {}
func (noopCounter) Add(float64) {}

type noopGauge struct{}

func (noopGauge) Inc()        {}
func (noopGauge) Dec()        {}
func (noopGauge) Set(float64) {}
func (noopGauge) Add(float64) {}

type noopHistogram struct{}

func (noopHistogram) Observe(float64) {}
