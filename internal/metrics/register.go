package metrics

import "github.com/kthcloud/podsh/pkg/metrics"

const (
	PodshSuccessfulAuth = "podsh_successful_auth"
	PodshFailedAuth     = "podsh_failed_auth"

	PodshActiveK8sExecStreams    = "podsh_k8s_active_exec_streams"
	PodshK8sActiveTunnelForwards = "podsh_k8s_active_tunnel_forwards"
	PodshK8sActiveTunnelStreams  = "podsh_k8s_active_tunnel_streams"
	PodshK8sActiveSFTPStreams    = "podsh_k8s_active_sftp_streams"
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

func RegisterK8sMetrics(m metrics.Metrics) {
	counters := map[string]string{}
	for name, help := range counters {
		m.RegisterCounter(name, help)
	}

	gauges := map[string]string{
		PodshActiveK8sExecStreams:    "Number of active k8s exec streams, used for shell and command exec",
		PodshK8sActiveTunnelForwards: "Number of active k8s tunnel forwards, a forward has multiple streams",
		PodshK8sActiveTunnelStreams:  "Number of active k8s tunnel streams, a stream has a parent forward",
		PodshK8sActiveSFTPStreams:    "Number of active k8s SFTP (via SPDY) streams, used for scp / sftp",
	}
	for name, help := range gauges {
		m.RegisterGauge(name, help)
	}

	histograms := map[string]string{}
	for name, help := range histograms {
		m.RegisterHistogram(name, help, nil)
	}
}
