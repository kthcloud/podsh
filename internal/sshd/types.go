package sshd

import (
	"context"
	"io"
)

type Identity struct {
	User              string
	UserID            string
	PublicKey         []byte
	Metadata          map[string]string
	RemoteAddr        string
	RequestedHostname string
}

type Pty struct {
	Term   string
	Width  int
	Height int
}

type ResizeEvent struct {
	Width  int
	Height int
}

type Session interface {
	Context() context.Context
	Identity() Identity

	Pty() (Pty, bool)
	Resize() <-chan ResizeEvent

	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer

	Exit(code int) error
}

type SessionHandler interface {
	HandleSession(Session)
}
