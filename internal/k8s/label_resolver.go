package k8s

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kthcloud/podsh/internal/sshd"
	corev1 "k8s.io/api/core/v1"
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

func (r *LabelResolver) Resolve(ctx context.Context, identity sshd.Identity) (*Target, error) {
	if identity.RequestedHostname == "" || identity.UserID == "" {
		return nil, ErrBadResolveRequest
	}
	pods, err := r.kubeClient.CoreV1().Pods(r.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf(
			"owner-id=%s,app.kubernetes.io/deploy-name=%s",
			identity.UserID,
			identity.RequestedHostname,
		),
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		// Skip pods that are not running
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		// Skip pods that are not ready
		ready := false
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}
		if !ready {
			continue
		}

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

	return nil, fmt.Errorf("user %s cannot access pod %s", identity.User, identity.RequestedHostname)
}
