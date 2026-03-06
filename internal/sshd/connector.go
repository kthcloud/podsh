package sshd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	metricsConstants "github.com/kthcloud/podsh/internal/metrics"
	"github.com/kthcloud/podsh/pkg/metrics"
	"github.com/kthcloud/podsh/pkg/ssh/requests"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNotPermitted       = errors.New("connection is not permitted")
	ErrGoroutinesExceeded = errors.New("goroutine count (per connection) exeeded for connection")
	ErrSessionTimeout     = errors.New("session timed out")
)

type ChannelType = string

const (
	ChannelTypeSession        ChannelType = "session"
	ChannelTypeDirectTCPIP    ChannelType = "direct-tcpip"
	ChannelTypeForwardedTCPIP ChannelType = "forwarded-tcpip"
)

type RequestType = string

const (
	RequestTypeKeepalive    RequestType = "keepalive@openssh.com"
	RequestTypePtyReq       RequestType = "pty-req"
	RequestTypeWindowChange RequestType = "window-change"
	RequestTypeShell        RequestType = "shell"
	RequestTypeExec         RequestType = "exec"
	RequestTypeEnv          RequestType = "env"
	RequestTypeSubsystem    RequestType = "subsystem"
)

type ResponseType = string

const (
	ResponseTypeExitStatus ResponseType = "exit-status"
)

type SubsystemType = string

const (
	SubsystemTypeSFTP SubsystemType = "sftp"
)

type Connector interface {
	Handle(conn net.Conn) error
}

type Context interface {
	context.Context
	Identity() Identity
	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer
}

type ShellContext interface {
	Context
	Resize() <-chan ResizeEvent
}

type BaseContext struct {
	ctx      context.Context
	identity Identity
	ch       ssh.Channel
}

func NewBaseContext(ctx context.Context, identity Identity, ch ssh.Channel) BaseContext {
	return BaseContext{
		ctx:      ctx,
		identity: identity,
		ch:       ch,
	}
}

func (bc BaseContext) Stdin() io.Reader { return bc.ch }

func (bc BaseContext) Stdout() io.Writer { return bc.ch }

func (bc BaseContext) Stderr() io.Writer { return bc.ch.Stderr() }

func (bc BaseContext) Identity() Identity {
	return bc.identity
}

func (bc BaseContext) Deadline() (deadline time.Time, ok bool) {
	return bc.ctx.Deadline()
}

func (bc BaseContext) Done() <-chan struct{} {
	return bc.ctx.Done()
}

func (bc BaseContext) Err() error {
	return bc.ctx.Err()
}

func (bc BaseContext) Value(key any) any {
	return bc.ctx.Value(key)
}

type ShellHandler interface {
	HandleShell(ctx ShellContext) error
	HandleExec(ctx Context, command ...string) error
	SFTPHandler
}

type Forwarder interface {
	Forward(in io.Reader, out io.Writer) error
	Close() error
}

type ForwardHandler interface {
	OpenTunnel(ctx context.Context, identity Identity, req requests.DirectTCPIP) (Forwarder, error)
}

type SFTPHandler interface {
	HandleSFTP(ctx Context) error
}

type Handler interface {
	ShellHandler
	ForwardHandler
}

type ConnectorImpl struct {
	ctx    context.Context
	logger *slog.Logger
	config *ssh.ServerConfig

	handler Handler
	metrics metrics.Metrics

	perConnGoroutineCap int
}

func NewConnectorImpl(ctx context.Context, logger *slog.Logger, config *ssh.ServerConfig, handler Handler, metrics metrics.Metrics) *ConnectorImpl {
	return &ConnectorImpl{
		ctx:                 ctx,
		logger:              logger,
		config:              config,
		perConnGoroutineCap: 10,
		handler:             handler,
		metrics:             metrics,
	}
}

func (ci *ConnectorImpl) Handle(conn net.Conn) error {
	log := ci.logger.With("remoteAddr", conn.RemoteAddr().String())

	log.Debug("[TRACE] Handle")
	defer log.Debug("[TRACE] Handle exit")
	defer conn.Close()

	deadline := time.Now().Add(10 * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set handshake deadline: %w", err)
	}

	connection, chans, reqs, err := ssh.NewServerConn(conn, ci.config)
	if err != nil {
		ci.metrics.Counter(metricsConstants.PodshFailedAuth).Inc()
		return errors.Join(err, ErrNotPermitted)
	}

	_ = conn.SetDeadline(time.Time{})

	defer connection.Close()

	var identity Identity
	if perms := connection.Permissions; perms != nil {
		if raw, ok := perms.Extensions["identity"]; ok {
			identity = decodeIdentity(raw)
		}
		if raw, ok := perms.Extensions["requested-host"]; ok {
			identity.RequestedHostname = raw
		}
	}

	if identity.UserID == "" {
		log.Warn("authorized user has no user id, denying")
		ci.metrics.Counter(metricsConstants.PodshFailedAuth).Inc()
		return ErrNotPermitted
	}

	if identity.RequestedHostname == "" {
		log.Warn("authorized user has no requested host, denying")
		ci.metrics.Counter(metricsConstants.PodshFailedAuth).Inc()
		return ErrNotPermitted
	}

	ci.metrics.Counter(metricsConstants.PodshSuccessfulAuth).Inc()

	log = log.With("UserID", identity.UserID, "RequestedHostname", identity.RequestedHostname)

	var eg errgroup.Group
	eg.SetLimit(ci.perConnGoroutineCap)

	if !eg.TryGo(func() error {
		return handleGlobalReqs(ci.ctx, reqs, log)
	}) {
		return ErrGoroutinesExceeded
	}

	forwarded := make(map[string]Forwarder)

loop:
	for {
		select {
		case <-ci.ctx.Done():
			break loop
		case channel, ok := <-chans:
			if !ok {
				break loop
			}
			switch channel.ChannelType() {
			case ChannelTypeSession:
				ch, chReqs, err := channel.Accept()
				if err != nil {
					log.Error("error accepting channel", "error", err)
					continue
				}

				bctx := NewBaseContext(ci.ctx, identity, ch)

				if !eg.TryGo(func() error {
					defer ch.Close()
					if err := handleSessionCh(bctx, chReqs, log, ci.handler); err != nil {
						status := struct{ Status uint32 }{1}
						ch.SendRequest(ResponseTypeExitStatus, false, ssh.Marshal(status))
						return err
					}
					status := struct{ Status uint32 }{0}
					ch.SendRequest(ResponseTypeExitStatus, false, ssh.Marshal(status))
					return nil
				}) {
					status := struct{ Status uint32 }{1}
					ch.SendRequest(ResponseTypeExitStatus, false, ssh.Marshal(status))
					ch.Close()
					log.Error("goroutines cap exceeded, cannot handle", "channelType", channel.ChannelType())
					continue
				}
			case ChannelTypeDirectTCPIP:
				var fwRq requests.DirectTCPIP
				if err := ssh.Unmarshal(channel.ExtraData(), &fwRq); err != nil {
					channel.Reject(ssh.Prohibited, "Bad Request")
					continue
				}

				key := fmt.Sprintf("%s:%d", fwRq.DestAddr, fwRq.DestPort)
				fw, exists := forwarded[key]
				if !exists {

					fm, err := ci.handler.OpenTunnel(ci.ctx, identity, fwRq)
					if err != nil {
						log.Error("failed to open k8s tunnel", "key", key)
						channel.Reject(ssh.ConnectionFailed, err.Error())
						continue
					}

					forwarded[key] = fm
					fw = fm
					log.Debug("opended k8s tunnel", "key", key)

				} else {
					log.Debug("re-using open k8s tunnel", "key", key)
				}

				if fw == nil {
					channel.Reject(ssh.ConnectionFailed, "idk")
					log.Error("fw is nil, dropping", "channelType", channel.ChannelType(), "key", key)
					continue
				}

				ch, _, err := channel.Accept()
				if err != nil {
					log.Error("error accepting channel", "error", err)
					continue
				}

				if !eg.TryGo(func() error {
					defer ch.Close()
					// pipe ch (ssh.Channel) => fw (io.ReadWriter)
					if err := fw.Forward(ch, ch); err != nil {
						status := struct{ Status uint32 }{1}
						ch.SendRequest(ResponseTypeExitStatus, false, ssh.Marshal(status))
						return err
					}
					status := struct{ Status uint32 }{0}
					ch.SendRequest(ResponseTypeExitStatus, false, ssh.Marshal(status))
					return nil
				}) {
					status := struct{ Status uint32 }{1}
					ch.SendRequest(ResponseTypeExitStatus, false, ssh.Marshal(status))
					ch.Close()
					log.Error("goroutines cap exceeded, cannot handle", "channelType", channel.ChannelType())
					continue
				}

			default:
				log.Debug("Rejected unsupported channel", "channelType", channel.ChannelType())
				channel.Reject(ssh.UnknownChannelType, "Unsupported SSH channel")
				continue
			}
		}
	}

	for k, fw := range forwarded {
		_ = fw.Close()
		log.Info("closed fw", "key", k)
	}

	return eg.Wait()
}

type SessionV2 struct {
	BaseContext
	shell    chan struct{}
	sftp     chan struct{}
	exec     chan string
	resizeCh chan ResizeEvent
}

func NewSession(ctx BaseContext) *SessionV2 {
	return &SessionV2{
		BaseContext: ctx,
		shell:       make(chan struct{}),
		sftp:        make(chan struct{}),
		exec:        make(chan string, 1),
		resizeCh:    make(chan ResizeEvent, 8),
	}
}

func (s *SessionV2) Resize() <-chan ResizeEvent { return s.resizeCh }

func handleSessionCh(ctx BaseContext, reqs <-chan *ssh.Request, logger *slog.Logger, handler ShellHandler) (err error) {
	logger.Debug("[TRACE] handleSession")
	defer logger.Debug("[TRACE] handleSession exit")

	childCtx, cancel := context.WithCancel(ctx)
	sess := NewSession(ctx)

	var eg errgroup.Group
	eg.Go(func() error {
		defer cancel()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return ErrSessionTimeout
		case <-sess.shell:
			return handler.HandleShell(sess)
		case command, ok := <-sess.exec:
			if !ok {
				return nil
			}
			err := handler.HandleExec(sess, command)
			code := 0
			if err != nil {
				code = 1
			}
			status := struct{ Status uint32 }{uint32(code)}
			_, errs := sess.ch.SendRequest(ResponseTypeExitStatus, false, ssh.Marshal(&status))
			return errors.Join(errs, err)
		case <-sess.sftp:
			if err := handler.HandleSFTP(sess); err != nil {
				logger.Error("SFTP failed", "error", err)
				return err
			}
			return nil
		}
	})

	for {
		select {
		case <-childCtx.Done():
			return errors.Join(eg.Wait(), ctx.Err())
		case req, ok := <-reqs:
			if !ok {
				return eg.Wait()
			}
			switch req.Type {
			case RequestTypePtyReq:
				var ptyReq requests.PTYReq
				if err := ssh.Unmarshal(req.Payload, &ptyReq); err != nil {
					logger.Error("Failed to unmarshal pty req", "error", err)
					_ = req.Reply(false, nil)
					continue
				}
				sess.resizeCh <- ResizeEvent{Width: int(ptyReq.Cols), Height: int(ptyReq.Rows)}
				logger.Debug("Got PTY", "ptyReq", ptyReq)
				_ = req.Reply(true, nil)
			case RequestTypeWindowChange:
				var winchReq requests.WindowChangeRequest
				if err := ssh.Unmarshal(req.Payload, &winchReq); err != nil {
					logger.Error("Failed to unmarshal window-change req", "error", err)
					_ = req.Reply(false, nil)
					continue
				}
				sess.resizeCh <- ResizeEvent{Width: int(winchReq.Cols), Height: int(winchReq.Rows)}
				_ = req.Reply(true, nil)
			case RequestTypeEnv:
				var envReq requests.EnvRequest
				if err := ssh.Unmarshal(req.Payload, &envReq); err != nil {
					logger.Error("Failed to unmarshal env req", "error", err)
					_ = req.Reply(false, nil)
					continue
				}
				logger.Info("got env", "envReq", envReq)
				_ = req.Reply(true, nil)
			case RequestTypeShell:
				select {
				case <-sess.shell:
					logger.Warn("shell already received for channel but received a new shell request, SSH spec only allows one shell to get ok per channel")
					_ = req.Reply(false, nil)
				default:
					close(sess.shell)
					_ = req.Reply(true, nil)
				}
			case RequestTypeExec:
				var execReq requests.ExecRequest
				if err := ssh.Unmarshal(req.Payload, &execReq); err != nil {
					logger.Error("Failed to unmarshal pty req", "error", err)
					_ = req.Reply(false, nil)
					continue
				}
				sess.exec <- execReq.Command
				logger.Info("Got exec", "execReq", execReq)
				_ = req.Reply(true, nil)
			case RequestTypeSubsystem:
				var subsystemReq requests.SubsystemRequest
				if err := ssh.Unmarshal(req.Payload, &subsystemReq); err != nil {
					logger.Error("Failed to unmarshal subsystem req", "error", err)
					_ = req.Reply(false, nil)
					continue
				}
				switch subsystemReq.Subsystem {
				case SubsystemTypeSFTP:
					close(sess.sftp)
					_ = req.Reply(true, nil)
				default:
					logger.Debug("ignoring unsupported subsystem", "subsystem", subsystemReq.Subsystem)
					_ = req.Reply(false, nil)
				}
			default:
				logger.Debug("ignoring unknown request", "type", req.Type)
				_ = req.Reply(false, nil)
			}
		}
	}
}

func handleGlobalReqs(ctx context.Context, reqs <-chan *ssh.Request, logger *slog.Logger) (err error) {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req, ok := <-reqs:
			if !ok {
				return
			}
			switch req.Type {
			case RequestTypeKeepalive:
				_ = req.Reply(true, nil)
			default:
				logger.Debug("ignoring global request", "type", req.Type)
				_ = req.Reply(false, nil)
			}
		}
	}
}
