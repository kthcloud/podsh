package gateway

import (
	"context"
	"fmt"

	"github.com/kthcloud/podsh/internal/sshd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type LabelResolver struct {
	kubeClient *kubernetes.Clientset
	namespace  string
}

func NewLabelResolver(kc *kubernetes.Clientset, namespace string) *LabelResolver {
	lr := &LabelResolver{
		kubeClient: kc,
		namespace:  namespace,
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
		return &Target{
			Namespace: pod.Namespace,
			Pod:       pod.Name,
			Container: pod.Spec.Containers[0].Name,
			// some way to specify shell would be nice
			Command: []string{"/bin/sh"},
		}, nil
	}

	return nil, fmt.Errorf("user %s cannot access pod %s", id.User, hostname)
}
