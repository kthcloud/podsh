package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kthcloud/podsh/internal/defaults"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The ephemeral container we start will start pkg/sftp server as standalone
// and listen on os.Stdin and output to os.Stdout
// We should spawn and connect to it once ready. So we can forward our sftp commands to its stdin

func ensureAgentContainer(
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
		if ec.Name == "podsh-agent" {
			return nil
		}
	}

	agent := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name: "podsh-agent",
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
		agent,
	)

	patchBytes, err := json.Marshal(map[string]any{
		"spec": map[string]any{
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

func waitForAgentReady(ctx context.Context, client kubernetes.Interface, namespace, podName string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			running, terminatedReason, err := agentRunning(pod)
			if err != nil {
				return err
			}

			if running {
				return nil
			}
			if terminatedReason != "" {
				return fmt.Errorf("ephemeral container exited: %s", terminatedReason)
			}
		}
	}
}

func agentRunning(pod *corev1.Pod) (running bool, terminatedReason string, err error) {
	for _, status := range pod.Status.EphemeralContainerStatuses {
		if status.Name == "podsh-agent" {
			if status.State.Running != nil {
				return true, "", nil
			}
			if status.State.Terminated != nil {
				return false, status.State.Terminated.Reason, nil
			}
		}
	}
	return false, "", nil
}
