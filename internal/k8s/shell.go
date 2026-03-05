package k8s

import (
	"errors"

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

	target.Command = append(target.Command, command...)

	exec, err := remoteExec(hi.client, hi.config, *target, RemoteExecOptions{
		stdin:  ctx.Stdin() != nil,
		stdout: ctx.Stdout() != nil,
		stderr: ctx.Stderr() != nil,
	})
	if err != nil {
		return err
	}

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
	// log.Println("resize term", "witdh", ev.Width, "height", ev.Height)
	return &remotecommand.TerminalSize{
		Width:  uint16(ev.Width),
		Height: uint16(ev.Height),
	}
}
