package k8s

import (
	"context"

	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/kthcloud/podsh/pkg/ssh/requests"
)

func (hi *HandlerImpl) OpenTunnel(ctx context.Context, identity sshd.Identity, req requests.DirectTCPIP) (sshd.Forwarder, error) {
	target, err := hi.resolver.Resolve(ctx, identity)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, ErrTargetNil
	}
	return NewForwardManager(ctx, hi.client, hi.config, target.Namespace, target.Pod, int(req.DestPort), hi.metrics)
}
