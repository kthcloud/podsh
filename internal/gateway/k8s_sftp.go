package gateway

import (
	"context"
	"io"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type KubernetesSFTP struct {
	Client kubernetes.Interface
	Config *rest.Config
}

func NewK8sSFTP(c kubernetes.Interface, cfg *rest.Config) SFTP {
	return &KubernetesSFTP{
		Client: c,
		Config: cfg,
	}
}

func (k *KubernetesSFTP) Exec(
	ctx context.Context,
	t *Target,
	in io.Reader,
	out io.Writer,
) error {
	if err := ensureHelperContainer(
		ctx,
		k.Client,
		t.Namespace,
		t.Pod,
		t.Container,
	); err != nil {
		log.Println("error while ensuring helper:", err)
		return err
	}

	log.Println("helper exists")

	if err := waitForHelperReady(
		ctx,
		k.Client,
		t.Namespace,
		t.Pod,
	); err != nil {
		log.Println("error while waiting for helper:", err)
		return err
	}

	log.Println("helper ready")

	req := k.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(t.Pod).
		Namespace(t.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "podsh-helper",
			Command:   []string{"/usr/bin/podsftp"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(
		k.Config,
		"POST",
		req.URL(),
	)
	if err != nil {
		log.Println("error creating spdy exec:", err)
		return err
	}

	log.Println("spdy success")

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  in,
		Stdout: out,
		Stderr: io.Discard,
		Tty:    false,
	})

	log.Println("stream has (nil = ok) err=", err)

	return err
}
