package gateway

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kthcloud/podsh/internal/sshd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type LabelResolver struct {
	kubeClient *kubernetes.Clientset
	namespace  string
	logger     *slog.Logger
}

func NewLabelResolver(kc *kubernetes.Clientset, namespace string) *LabelResolver {
	lr := &LabelResolver{
		kubeClient: kc,
		namespace:  namespace,
		logger:     slog.Default(), // TODO: get from args
	}

	return lr
}

func (r *LabelResolver) Resolve(ctx context.Context, hostname string, id sshd.Identity) (*Target, error) {
	pods, err := r.kubeClient.CoreV1().Pods(r.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf(
			"owner-id=%s,app.kubernetes.io/deploy-name=%s",
			id.UserID,
			hostname,
		),
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		var containerName string = pod.Spec.Containers[0].Name
		if len(pod.Spec.Containers) > 1 {
			r.logger.Warn("resolved to pod with more than one container, defaulting to", "containerName", containerName)
		}
		return &Target{
			Namespace: pod.Namespace,
			Pod:       pod.Name,
			Container: containerName,
			// some way to specify shell would be nice
			Command: []string{"/bin/sh"},
		}, nil
	}

	return nil, fmt.Errorf("user %s cannot access pod %s", id.User, hostname)
}
