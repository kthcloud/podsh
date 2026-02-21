package sshd

import (
	"context"
	"net"

	"golang.org/x/crypto/ssh"
)

func (s *Server) handleConn(parent context.Context, netConn net.Conn) {
	logger := s.logger.With("remote", netConn.RemoteAddr().String())

	defer netConn.Close()

	if s.hostSigner == nil {
		logger.Error("no host signer configured")
		return
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: s.publicKeyCallback(parent, logger),
		ServerVersion:     "SSH-2.0-podsh",
		BannerCallback: func(conn ssh.ConnMetadata) string {
			return `   __   __  __       __             __
  / /__/ /_/ /  ____/ /__  __ _____/ /
 /  '_/ __/ _ \/ __/ / _ \/ // / _  / 
/_/\_\\__/_//_/\__/_/\___/\_,_/\_,_/  
                                      
Connecting to your pod...
`
		},
	}

	config.AddHostKey(s.hostSigner)

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, config)
	if err != nil {
		logger.Debug("ssh handshake failed", "err", err)
		return
	}
	defer sshConn.Close()

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	logger = logger.With(
		"user", sshConn.User(),
		"client_version", string(sshConn.ClientVersion()),
	)

	logger.Info("ssh connection established")

	// global requests (keepalive etc)
	go s.discardRequests(ctx, logger, reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			logger.Debug("rejecting unsupported channel",
				"type", newChan.ChannelType(),
			)
			_ = newChan.Reject(ssh.UnknownChannelType, "only session channels supported")
			continue
		}

		channel, requests, err := newChan.Accept()
		if err != nil {
			logger.Debug("channel accept failed", "err", err)
			continue
		}

		s.mu.RLock()
		handler := s.handler
		s.mu.RUnlock()

		if handler == nil {
			logger.Error("no session handler registered")
			_ = channel.Close()
			continue
		}

		sess := newSession(ctx, cancel, sshConn, channel, requests, logger)

		go func() {
			defer sess.close()

			select {
			case <-ctx.Done():
			case <-sess.ready:
				handler.HandleSession(sess)
			}

			// if handler returns without Exit(), ensure clean shutdown
			_ = sess.Exit(0)
		}()
	}

	logger.Info("ssh connection closed")
}
