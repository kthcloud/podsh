package gateway

import (
	"context"
	"fmt"
	"io"

	"github.com/kthcloud/podsh/internal/sshd"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type K8sExecutor struct {
	Client *kubernetes.Clientset
	Rest   *rest.Config
}

func NewK8sExecutor(kc *kubernetes.Clientset, rest *rest.Config) *K8sExecutor {
	return &K8sExecutor{
		Client: kc,
		Rest:   rest,
	}
}

func (k *K8sExecutor) Exec(ctx context.Context, t *Target, pty sshd.Pty, s Streams) (int, error) {
	req := k.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(t.Pod).
		Namespace(t.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: t.Container,
			Command:   t.Command,
			Stdin:     s.Stdin != nil,
			Stdout:    s.Stdout != nil,
			Stderr:    s.Stderr != nil,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.Rest, "POST", req.URL())
	if err != nil {
		return -1, fmt.Errorf("failed to create SPDY executor: %w", err)
	}

	// Handle PTY resize events
	/*resizeFn := func(size remotecommand.TerminalSize) {}
	if pty.Term != "" {
		resizeFn = func(size remotecommand.TerminalSize) {
			select {
			case <-ctx.Done():
			case <-s.Resize:
				// We’ll use the latest size in a goroutine
			default:
			}
		}
	}*/

	// Stream stdin/stdout/stderr
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             s.Stdin,
		Stdout:            s.Stdout,
		Stderr:            s.Stderr,
		Tty:               true,
		TerminalSizeQueue: &resizeQueue{Resize: s.Resize},
	})
	if err != nil {
		return -1, fmt.Errorf("k8s exec failed: %w", err)
	}

	return 0, nil
}

func (k *K8sExecutor) ExecRaw(ctx context.Context, t *Target, s Streams) error {
	req := k.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(t.Pod).
		Namespace(t.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: t.Container,
			Command:   t.Command,
			Stdin:     s.Stdin != nil,
			Stdout:    s.Stdout != nil,
			Stderr:    s.Stderr != nil,
			TTY:       false, // IMPORTANT
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.Rest, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  s.Stdin,
		Stdout: s.Stdout,
		Stderr: s.Stderr,
		Tty:    false,
	})
}

func (k *K8sExecutor) UploadFile(ctx context.Context, t *Target, data io.Reader, remotePath string) error {
	t.Command = []string{"sh", "-c", fmt.Sprintf("cat > %s", remotePath)}
	return k.ExecRaw(ctx, t, Streams{
		Stdin: data,
	})
}

func (k *K8sExecutor) DownloadFile(ctx context.Context, t *Target, remotePath string, out io.Writer) error {
	t.Command = []string{"cat", remotePath}
	return k.ExecRaw(ctx, t, Streams{
		Stdout: out,
	})
}

func (k *K8sExecutor) Mkdir(ctx context.Context, t *Target, remoteDir string) error {
	t.Command = []string{"mkdir", "-p", remoteDir}
	return k.ExecRaw(ctx, t, Streams{})
}

type resizeQueue struct {
	Resize <-chan sshd.ResizeEvent
}

func (r *resizeQueue) Next() *remotecommand.TerminalSize {
	ev, ok := <-r.Resize
	if !ok {
		return nil
	}
	return &remotecommand.TerminalSize{
		Width:  uint16(ev.Width),
		Height: uint16(ev.Height),
	}
}
