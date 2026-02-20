package defaults

import "time"

const (
	DefaultNamespace   = "deploy"
	DefaultBindAddress = "localhost:2222"
)

const (
	DefaultLimitRate  = 1.2
	DefaultLimitBurst = 12
	DefaultLimitTTL   = 30 * time.Second
)
