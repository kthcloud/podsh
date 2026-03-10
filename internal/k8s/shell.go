package k8s

import (
	"errors"
	"strings"

	metricsConstants "github.com/kthcloud/podsh/internal/metrics"
	"github.com/kthcloud/podsh/internal/sshd"
	"k8s.io/client-go/tools/remotecommand"
)

var ErrTargetNil = errors.New("nil target")

func (hi *HandlerImpl) HandleShell(ctx sshd.ShellContext) (err error) {
	target, err := hi.resolver.Resolve(ctx, ctx.Identity())
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNil
	}

	exec, err := remoteExec(hi.client, hi.config, *target, RemoteExecOptions{
		stdin:  ctx.Stdin() != nil,
		stdout: ctx.Stdout() != nil,
		stderr: ctx.Stderr() != nil,
		tty:    true,
	})
	if err != nil {
		return err
	}

	hi.metrics.Gauge(metricsConstants.PodshActiveK8sExecStreams).Inc()
	defer func() {
		hi.metrics.Gauge(metricsConstants.PodshActiveK8sExecStreams).Dec()
	}()

	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             ctx.Stdin(),
		Stdout:            ctx.Stdout(),
		Stderr:            ctx.Stderr(),
		Tty:               true,
		TerminalSizeQueue: &resizeQueue{Resize: ctx.Resize()},
	}); err != nil {
		return err
	}

	return
}

func (hi *HandlerImpl) HandleExec(ctx sshd.Context, command ...string) error {
	target, err := hi.resolver.Resolve(ctx, ctx.Identity())
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNil
	}

	target.Command = make([]string, 0, 10)
	for _, c := range command {
		target.Command = append(target.Command, splitArgs(c)...)
	}

	exec, err := remoteExec(hi.client, hi.config, *target, RemoteExecOptions{
		stdin:  ctx.Stdin() != nil,
		stdout: ctx.Stdout() != nil,
		stderr: ctx.Stderr() != nil,
	})
	if err != nil {
		return err
	}

	hi.metrics.Gauge(metricsConstants.PodshActiveK8sExecStreams).Inc()
	defer func() {
		hi.metrics.Gauge(metricsConstants.PodshActiveK8sExecStreams).Dec()
	}()

	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  ctx.Stdin(),
		Stdout: ctx.Stdout(),
		Stderr: ctx.Stderr(),
	}); err != nil {
		return err
	}

	return nil
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

func buildCommand(command []string) []string {
	var result []string
	for _, c := range command {
		result = append(result, c)
	}
	return result
}

// TODO: do better!
func splitArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		switch ch {
		case '\'':
			inQuotes = !inQuotes // toggle quotes
		case ' ':
			if inQuotes {
				current.WriteByte(ch) // keep spaces inside quotes
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	// Special handling for sh -c: combine everything after -c into one argument
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "sh" && args[i+1] == "-c" && i+2 < len(args) {
			// Combine remaining arguments into one
			combined := strings.Join(args[i+2:], " ")
			args = args[:i+2]             // keep ["sh", "-c"]
			args = append(args, combined) // append combined command
			break
		}
	}

	return args
}
