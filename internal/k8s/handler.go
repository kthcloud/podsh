package k8s

import (
	"fmt"

	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/kthcloud/podsh/pkg/ssh/requests"
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

func (HandlerImpl) HandleSFTP(ctx sshd.Context) error {
	return fmt.Errorf("not impl")
}

func (HandlerImpl) HandleForward(ctx sshd.Context, req requests.DirectTCPIP) error {
	return fmt.Errorf("not impl")
}
