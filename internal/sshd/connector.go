package sshd

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/kthcloud/podsh/pkg/ssh/requests"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNotPermitted       = errors.New("connection is not permitted")
	ErrGoroutinesExceeded = errors.New("goroutine count (per connection) exeeded for connection")
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

type Connector interface {
	Handle(conn net.Conn) error
}

type ConnectorImpl struct {
	ctx                 context.Context
	logger              *slog.Logger
	config              *ssh.ServerConfig
	perConnGoroutineCap int
}

func NewConnectorImpl(ctx context.Context, logger *slog.Logger, config *ssh.ServerConfig) *ConnectorImpl {
	return &ConnectorImpl{
		ctx:                 ctx,
		logger:              logger,
		config:              config,
		perConnGoroutineCap: 10,
	}
}

func (ci *ConnectorImpl) Handle(conn net.Conn) error {
	log := ci.logger.With("remoteAddr", conn.RemoteAddr().String())

	log.Debug("[TRACE] Handle")
	defer log.Debug("[TRACE] Handle exit")
	defer conn.Close()

	connection, chans, reqs, err := ssh.NewServerConn(conn, ci.config)
	if err != nil {
		return errors.Join(err, ErrNotPermitted)
	}

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
		return ErrNotPermitted
	}

	if identity.RequestedHostname == "" {
		return ErrNotPermitted
	}

	log = log.With("UserID", identity.UserID, "RequestedHostname", identity.RequestedHostname)

	var eg errgroup.Group
	eg.SetLimit(ci.perConnGoroutineCap)

	if !eg.TryGo(func() error {
		return handleGlobalReqs(reqs, log)
	}) {
		return ErrGoroutinesExceeded
	}

	for channel := range chans {
		switch channel.ChannelType() {
		case ChannelTypeSession:
			ch, chReqs, err := channel.Accept()
			if err != nil {
				log.Error("error accepting channel", "error", err)
				continue
			}

			if !eg.TryGo(func() error {
				defer ch.Close()
				if err := handleSessionCh(chReqs, log); err != nil {
					log.Error("Session exit", "error", err)
					return err
				}
				return nil
			}) {
				ch.Close()
				log.Error("goroutines cap exceeded by cap, cannot handle", "channelType", channel.ChannelType())
				continue
			}
		case ChannelTypeDirectTCPIP:
			var fwRq requests.DirectTCPIP
			if err := ssh.Unmarshal(channel.ExtraData(), &fwRq); err != nil {
				channel.Reject(ssh.Prohibited, "Bad Request")
				continue
			}

			ch, chReqs, err := channel.Accept()
			if err != nil {
				log.Error("error accepting channel", "error", err)
				continue
			}

			if eg.TryGo(func() error {
				defer ch.Close()
				if err := handleDirectTCPIP(chReqs, fwRq, log); err != nil {
					log.Error("Forward exit", "error", err)
					return err
				}
				return nil
			}) {
				ch.Close()
				log.Error("goroutines cap exceeded by cap, cannot handle", "channelType", channel.ChannelType())
				continue
			}

		default:
			channel.Reject(ssh.UnknownChannelType, "Unsupported SSH channel")
			continue
		}
	}

	return eg.Wait()
}

type SessionV2 struct {
	pty   *requests.PTYReq
	shell chan struct{}
}

func NewSession() *SessionV2 {
	return &SessionV2{
		shell: make(chan struct{}),
	}
}

func handleSessionCh(reqs <-chan *ssh.Request, logger *slog.Logger) (err error) {
	logger.Info("handleSession")
	defer logger.Debug("[TRACE] handleSession exit")

	sess := NewSession()
	for req := range reqs {
		switch req.Type {
		case RequestTypePtyReq:
			var ptyReq requests.PTYReq
			if err := ssh.Unmarshal(req.Payload, &ptyReq); err != nil {
				logger.Error("Failed to unmarshal pty req", "error", err)
				_ = req.Reply(false, nil)
				continue
			}
			sess.pty = &ptyReq
			logger.Info("Got PTY", "ptyReq", ptyReq)
			_ = req.Reply(true, nil)
		case RequestTypeShell:
			if sess.pty == nil {
				_ = req.Reply(false, nil)
			} else {
				select {
				case <-sess.shell:
					logger.Warn("shell already received for channel but received a new shell request, SSH spec only allows one shell to get ok per channel")
					_ = req.Reply(false, nil)
				default:
					close(sess.shell)
					_ = req.Reply(true, nil)
				}
			}
		default:
			logger.Debug("ignoring unknown request", "type", req.Type)
			_ = req.Reply(false, nil)
		}
	}

	return
}

func handleDirectTCPIP(reqs <-chan *ssh.Request, forwardRequest requests.DirectTCPIP, logger *slog.Logger) (err error) {
	logger.Info("handleDirectTCPIP", "request", forwardRequest)
	defer logger.Debug("[TRACE] handleDirectTCPIP exit")
	for req := range reqs {
		switch req.Type {
		default:
			logger.Debug("ignoring unknown request", "type", req.Type)
			_ = req.Reply(false, nil)
		}
	}
	return
}

func handleGlobalReqs(reqs <-chan *ssh.Request, logger *slog.Logger) (err error) {
	for req := range reqs {
		switch req.Type {
		case RequestTypeKeepalive:
			_ = req.Reply(true, nil)
		default:
			logger.Debug("ignoring global request", "type", req.Type)
			_ = req.Reply(false, nil)
		}
	}

	return
}
