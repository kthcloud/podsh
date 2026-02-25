// Package k8s contains kubernetes adapters for the sshd.Handler
package k8s

import (
	"context"

	"github.com/kthcloud/podsh/internal/sshd"
)

type Handler interface {
	sshd.Handler
}

type Target struct {
	Pod       string
	Namespace string
	Container string
	Command   []string
}

type Resolver interface {
	Resolve(ctx context.Context, identity sshd.Identity) (*Target, error)
}
