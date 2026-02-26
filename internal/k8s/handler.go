package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type HandlerImpl struct {
	client   kubernetes.Interface
	config   *rest.Config
	resolver Resolver
}

func New(client kubernetes.Interface, config *rest.Config, resolver Resolver) *HandlerImpl {
	hi := &HandlerImpl{
		client:   client,
		config:   config,
		resolver: resolver,
	}

	return hi
}
