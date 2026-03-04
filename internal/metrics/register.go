package metrics

import "github.com/kthcloud/podsh/pkg/metrics"

const (
	PodshSuccessfulAuth = "podsh_successful_auth"
	PodshFailedAuth     = "podsh_failed_auth"
)

func RegisterSSHdMetrics(m metrics.Metrics) {
	counters := map[string]string{
		PodshSuccessfulAuth: "Number of successful auth attempts",
		PodshFailedAuth:     "Number of failed auth attempts",
	}
	for name, help := range counters {
		m.RegisterCounter(name, help)
	}

	gauges := map[string]string{}
	for name, help := range gauges {
		m.RegisterGauge(name, help)
	}

	histograms := map[string]string{}
	for name, help := range histograms {
		m.RegisterHistogram(name, help, nil)
	}
}
