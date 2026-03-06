package defaults

import "time"

const (
	DefaultNamespace   = "deploy"
	DefaultBindAddress = "127.0.0.1:2222"
	DefaultMetricsAddr = "127.0.0.1:8080"
)

const (
	DefaultLimitRate  = 1.2
	DefaultLimitBurst = 12
	DefaultLimitTTL   = 30 * time.Second
)

const (
	DefaultPodshAgentImage           = "ghcr.io/kthcloud/podsh/agent:latest"
	DefaultPodshAgentImagePullPolicy = "IfNotPresent"
)

const (
	DefaultHostSignerPath = "/etc/podsh/key"
)

const (
	DefaultRedisAddress  = "localhost:6379"
	DefaultRedisDB       = 0
	DefaultRedisPassword = ""
)
