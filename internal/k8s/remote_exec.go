package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	corev1 "k8s.io/api/core/v1"
)

type RemoteExecOptions struct {
	stdin  bool
	stdout bool
	stderr bool
	tty    bool
}

func remoteExec(client kubernetes.Interface, cfg *rest.Config, target Target, opt RemoteExecOptions) (remotecommand.Executor, error) {
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(target.Pod).
		Namespace(target.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: target.Container,
			Command:   target.Command,
			Stdin:     opt.stdin,
			Stdout:    opt.stdout,
			Stderr:    opt.stderr,
			TTY:       opt.tty,
		}, scheme.ParameterCodec)

	return remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
}
