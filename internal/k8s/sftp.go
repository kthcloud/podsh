package k8s

import (
	"io"
	"log"

	metricsConstants "github.com/kthcloud/podsh/internal/metrics"
	"github.com/kthcloud/podsh/internal/sshd"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (hi *HandlerImpl) HandleSFTP(ctx sshd.Context) error {
	target, err := hi.resolver.Resolve(ctx, ctx.Identity())
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNil
	}

	if err := ensureAgentContainer(
		ctx,
		hi.client,
		target.Namespace,
		target.Pod,
		target.Container,
	); err != nil {
		log.Println("error while ensuring helper:", err)
		return err
	}

	if err := waitForAgentReady(
		ctx,
		hi.client,
		target.Namespace,
		target.Pod,
	); err != nil {
		log.Println("error while waiting for agent:", err)
		return err
	}

	req := hi.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(target.Pod).
		Namespace(target.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "podsh-agent",
			Command:   []string{"/usr/bin/podsftp"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(
		hi.config,
		"POST",
		req.URL(),
	)
	if err != nil {
		log.Println("error creating spdy exec:", err)
		return err
	}

	hi.metrics.Gauge(metricsConstants.PodshK8sActiveSFTPStreams).Inc()
	defer func() {
		hi.metrics.Gauge(metricsConstants.PodshK8sActiveSFTPStreams).Dec()
	}()

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  ctx.Stdin(),
		Stdout: ctx.Stdout(),
		Stderr: io.Discard,
		Tty:    false,
	})
}
