package metrics

import "net/http"

type Counter interface {
	Inc()
	Add(float64)
}

type Gauge interface {
	Inc()
	Dec()
	Set(float64)
	Add(float64)
}

type Histogram interface {
	Observe(float64)
}

type Metrics interface {
	RegisterCounter(name, help string) error
	RegisterGauge(name, help string) error
	RegisterHistogram(name, help string, buckets []float64) error

	Counter(name string) Counter
	Gauge(name string) Gauge
	Histogram(name string) Histogram

	Handler() http.Handler
}
