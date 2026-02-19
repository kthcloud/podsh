package sshd

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"sync"

	"golang.org/x/crypto/ssh"
)

type session struct {
	ctx    context.Context
	cancel context.CancelFunc

	conn    *ssh.ServerConn
	channel ssh.Channel
	reqs    <-chan *ssh.Request

	logger *slog.Logger

	identity Identity

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	pty      Pty
	hasPty   bool
	resizeCh chan ResizeEvent

	exitOnce  sync.Once
	closeOnce sync.Once

	ready chan struct{} // closed when shell request arrives
}

func newSession(
	parent context.Context,
	cancel context.CancelFunc,
	conn *ssh.ServerConn,
	ch ssh.Channel,
	reqs <-chan *ssh.Request,
	logger *slog.Logger,
) *session {
	ctx, c := context.WithCancel(parent)

	s := &session{
		ctx:      ctx,
		cancel:   func() { c(); cancel() },
		conn:     conn,
		channel:  ch,
		reqs:     reqs,
		logger:   logger,
		stdin:    ch,
		stdout:   ch,
		stderr:   ch.Stderr(),
		resizeCh: make(chan ResizeEvent, 8),
		ready:    make(chan struct{}),
	}

	// recover identity from auth callback
	if perms := conn.Permissions; perms != nil {
		if raw, ok := perms.Extensions["identity"]; ok {
			s.identity = decodeIdentity(raw)
		}
		if raw, ok := perms.Extensions["requested-host"]; ok {
			s.identity.RequestedHostname = raw
		}
	}

	// default PTY for non-PTY clients
	s.pty = Pty{
		Term:   "xterm",
		Width:  80,
		Height: 24,
	}
	s.hasPty = true

	go s.handleRequests()

	return s
}

func (s *session) handleRequests() {
	defer s.cancel()

	for req := range s.reqs {
		switch req.Type {

		case "pty-req":
			term, w, h, ok := parsePtyReq(req.Payload)
			if ok {
				s.pty = Pty{Term: term, Width: w, Height: h}
				s.hasPty = true
				_ = req.Reply(true, nil)
			} else {
				_ = req.Reply(false, nil)
			}

		case "window-change":
			w, h, ok := parseWinch(req.Payload)
			if ok {
				select {
				case s.resizeCh <- ResizeEvent{Width: w, Height: h}:
				default:
				}
			}

		case "shell":
			_ = req.Reply(true, nil)
			close(s.ready)
			return

		case "exec":
			// not supported in v1
			_ = req.Reply(false, nil)

		case "env":
			// TODO: parse variable name/value
			// log for now
			s.logger.Debug("env", "payload", string(req.Payload))
			/*name, value, ok := parseEnv(req.Payload)
			if ok {
				if s.env == nil {
					s.env = make(map[string]string)
				}
				s.env[name] = value
			}*/
			_ = req.Reply(true, nil) // accept env

		default:
			s.logger.Debug("unknown session request", "type", req.Type)
			_ = req.Reply(false, nil)
		}
	}
}

func (s *session) Context() context.Context { return s.ctx }

func (s *session) Identity() Identity { return s.identity }

func (s *session) Pty() (Pty, bool) { return s.pty, s.hasPty }

func (s *session) Resize() <-chan ResizeEvent { return s.resizeCh }

func (s *session) Stdin() io.Reader { return s.stdin }

func (s *session) Stdout() io.Writer { return s.stdout }

func (s *session) Stderr() io.Writer { return s.stderr }

func (s *session) Exit(code int) error {
	var err error

	s.exitOnce.Do(func() {
		status := struct{ Status uint32 }{uint32(code)}
		_, err = s.channel.SendRequest("exit-status", false, ssh.Marshal(&status))
		_ = s.channel.Close()
		s.cancel()
	})

	return err
}

func (s *session) close() {
	s.closeOnce.Do(func() {
		s.cancel()
		_ = s.channel.Close()
	})
}

func parsePtyReq(b []byte) (term string, w, h int, ok bool) {
	term, b, ok = readString(b)
	if !ok || len(b) < 8 {
		return
	}

	w = int(binary.BigEndian.Uint32(b))
	h = int(binary.BigEndian.Uint32(b[4:]))
	ok = true
	return
}

func parseWinch(b []byte) (w, h int, ok bool) {
	if len(b) < 8 {
		return
	}
	w = int(binary.BigEndian.Uint32(b))
	h = int(binary.BigEndian.Uint32(b[4:]))
	ok = true
	return
}

func readString(b []byte) (string, []byte, bool) {
	if len(b) < 4 {
		return "", nil, false
	}
	l := binary.BigEndian.Uint32(b)
	if uint32(len(b)) < 4+l {
		return "", nil, false
	}
	return string(b[4 : 4+l]), b[4+l:], true
}
