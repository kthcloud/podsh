package worker

import "context"

type Worker interface {
	// Start begins the worker loop.
	// It must return when ctx is cancelled.
	Start(ctx context.Context) error

	// Name is useful for logging/metrics.
	Name() string
}
