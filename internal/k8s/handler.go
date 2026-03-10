package k8s

import (
	"github.com/kthcloud/podsh/pkg/metrics"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type HandlerImpl struct {
	client   kubernetes.Interface
	config   *rest.Config
	resolver Resolver
	metrics  metrics.Metrics
}

func New(client kubernetes.Interface, config *rest.Config, resolver Resolver, metrics metrics.Metrics) *HandlerImpl {
	hi := &HandlerImpl{
		client:   client,
		config:   config,
		resolver: resolver,
		metrics:  metrics,
	}

	return hi
}
