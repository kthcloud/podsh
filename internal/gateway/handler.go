package gateway

import (
	"context"
	"errors"
	"log/slog"

	"github.com/kthcloud/podsh/internal/sshd"
)

type Handler struct {
	log      *slog.Logger
	resolver Resolver
	exec     Executor
}

func NewHandler(log *slog.Logger, r Resolver, e Executor) *Handler {
	return &Handler{
		log:      log,
		resolver: r,
		exec:     e,
	}
}

func (h *Handler) HandleSession(sess sshd.Session) {
	ctx := sess.Context()
	log := h.log.With(
		"user", sess.Identity().User,
		"userID", sess.Identity().UserID,
		"addr", sess.Identity().RemoteAddr,
		"requestedHostname", sess.Identity().RequestedHostname,
	)

	if ctx.Err() != nil {
		log.Info("Session cancelled", "error", ctx.Err())
		return
	}

	// must be PTY session
	pty, ok := sess.Pty()
	if !ok {
		log.Debug("rejecting non-pty session")
		_ = sess.Exit(1)
		return
	}

	// resolve hostname target
	hostname := sess.Identity().RequestedHostname
	if hostname == "" {
		log.Debug("missing hostname metadata")
		_ = sess.Exit(1)
		return
	}

	target, err := h.resolver.Resolve(ctx, hostname, sess.Identity())
	if err != nil {
		log.Info("resolve failed", "err", err)
		_ = sess.Exit(1)
		return
	}

	log = log.With(
		"ns", target.Namespace,
		"pod", target.Pod,
		"container", target.Container,
	)

	log.Info("session start")

	// 3 — run exec
	code, err := h.exec.Exec(ctx, target, pty, Streams{
		Stdin:  sess.Stdin(),
		Stdout: sess.Stdout(),
		Stderr: sess.Stderr(),
		Resize: sess.Resize(),
	})

	if err != nil && !errors.Is(err, context.Canceled) {
		log.Error("exec failed", "err", err)
		_ = sess.Exit(1)
		return
	}

	log.Info("session end", "code", code)
	_ = sess.Exit(code)
}
