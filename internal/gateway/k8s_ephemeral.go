package gateway

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/kthcloud/podsh/internal/defaults"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EphemeralSFTP wraps a streaming connection to the ephemeral container SFTP server
type EphemeralSFTP struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
}

// StartEphemeralSFTP starts the sftp server in the ephemeral container and returns a stream
func StartEphemeralSFTP(
	ctx context.Context,
	client kubernetes.Interface,
	config *rest.Config,
	namespace, podName, targetContainer string,
) (*EphemeralSFTP, error) {
	// Ensure ephemeral container exists
	if err := ensureHelperContainer(ctx, client, namespace, podName, targetContainer); err != nil {
		return nil, err
	}

	// Wait until it's ready
	if err := waitForHelperReady(ctx, client, namespace, podName); err != nil {
		return nil, err
	}

	// Exec into the ephemeral container, start SFTP server
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "podsh-helper",
			Command:   []string{"/usr/bin/podsftp"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	// Create pipes for stdin/stdout
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	go func() {
		// Stream SFTP process
		err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:  stdinReader,
			Stdout: stdoutWriter,
			Stderr: os.Stderr,
			Tty:    false,
		})
		// Close pipes on exit
		stdinWriter.CloseWithError(err)
		stdoutWriter.CloseWithError(err)
	}()

	return &EphemeralSFTP{
		Stdin:  stdinWriter,
		Stdout: stdoutReader,
	}, nil
}

// The ephemeral container we start will start pkg/sftp server as standalone
// and listen on os.Stdin and output to os.Stdout
// We should spawn and connect to it once ready. So we can forward our sftp commands to its stdin

func ensureHelperContainer(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, podName, targetContainer string,
) error {
	pod, err := client.CoreV1().
		Pods(namespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Check if already exists
	for _, ec := range pod.Spec.EphemeralContainers {
		if ec.Name == "podsh-helper" {
			return nil
		}
	}

	helper := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name: "podsh-helper",
			// We run a pkg/sftp binary and pipe our sftp requests to it
			Image:   defaults.DefaultPodshHelperImage,
			Command: []string{"/usr/bin/sleep", "infinity"},
			Stdin:   true,
			TTY:     false,
		},
		TargetContainerName: targetContainer,
	}

	pod.Spec.EphemeralContainers = append(
		pod.Spec.EphemeralContainers,
		helper,
	)

	patchBytes, err := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{
			"ephemeralContainers": pod.Spec.EphemeralContainers,
		},
	})
	if err != nil {
		return err
	}

	_, err = client.CoreV1().
		Pods(namespace).
		Patch(
			ctx,
			podName,
			types.StrategicMergePatchType,
			patchBytes,
			metav1.PatchOptions{},
			"ephemeralcontainers",
		)

	return err
}

func waitForHelperReady(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, podName string,
) error {
	watcher, err := client.CoreV1().
		Pods(namespace).
		Watch(ctx, metav1.ListOptions{
			FieldSelector: "metadata.name=" + podName,
		})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil // watch closed
			}

			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			for _, status := range pod.Status.EphemeralContainerStatuses {
				if status.Name == "podsh-helper" && status.State.Running != nil {
					return nil
				}
			}
		}
	}
}
