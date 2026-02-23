package gateway

import (
	"context"
	"fmt"
	"io"

	"github.com/kthcloud/podsh/internal/sshd"
)

type Target struct {
	Namespace string
	Pod       string
	Container string
	Command   []string
}

type Resolver interface {
	Resolve(ctx context.Context, hostname string, id sshd.Identity) (*Target, error)
}

type NopResolver struct{}

func (NopResolver) Resolve(_ context.Context, _ string, _ sshd.Identity) (*Target, error) {
	return nil, fmt.Errorf("not impl")
}

type Executor interface {
	Exec(ctx context.Context, t *Target, pty sshd.Pty, streams Streams) (int, error)
}

type NopExecutor struct{}

func (NopExecutor) Exec(_ context.Context, _ *Target, _ sshd.Pty, _ Streams) (int, error) {
	return -1, fmt.Errorf("not impl")
}

type Streams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Resize <-chan sshd.ResizeEvent
}

type SFTP interface {
	Exec(ctx context.Context, t *Target, in io.Reader, out io.Writer) error
}

type NopSFTP struct{}

func (NopSFTP) Exec(_ context.Context, _ *Target, _ io.ReadCloser, _ io.WriteCloser) error {
	return fmt.Errorf("not impl")
}
