package sshd_test

import (
	"bytes"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/kthcloud/podsh/pkg/metrics"
	"golang.org/x/crypto/ssh"
)

// FakeConnection is a minimal in-memory net.Conn for testing
type FakeConnection struct {
	readBuffer  bytes.Buffer
	writeBuffer bytes.Buffer
	closed      bool
}

func NewFakeConnection() *FakeConnection {
	var rBuf [1024]byte
	var wBuf [1024]byte
	return &FakeConnection{
		readBuffer:  *bytes.NewBuffer(rBuf[:]),
		writeBuffer: *bytes.NewBuffer(wBuf[:]),
	}
}

var _ net.Conn = &FakeConnection{}

func (fc *FakeConnection) Read(b []byte) (n int, err error) {
	if fc.closed {
		return 0, net.ErrClosed
	}
	return fc.readBuffer.Read(b)
}

func (fc *FakeConnection) Write(b []byte) (n int, err error) {
	if fc.closed {
		return 0, net.ErrClosed
	}
	return fc.writeBuffer.Write(b)
}

func (fc *FakeConnection) Close() error {
	fc.closed = true
	return nil
}

func (fc *FakeConnection) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 22}
}

func (fc *FakeConnection) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(192, 168, 1, 100), Port: 12345}
}

func (fc *FakeConnection) SetDeadline(t time.Time) error {
	return nil
}

func (fc *FakeConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (fc *FakeConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

// Helper to preload data for Read
func (fc *FakeConnection) PreloadRead(data []byte) {
	fc.readBuffer.Write(data)
}

// Helper to get what was written
func (fc *FakeConnection) WrittenData() []byte {
	return fc.writeBuffer.Bytes()
}

func TestConnector(t *testing.T) {
	serverConfig := sshd.NewTestServerConfig()
	clientConfig := sshd.NewTestClientConfig("testuser")

	serverConn, clientConn := sshd.NewPipePair()

	go func() {
		sshConn, chans, reqs, err := ssh.NewClientConn(clientConn, "", clientConfig)
		if err != nil {
			return
		}
		client := ssh.NewClient(sshConn, chans, reqs)
		defer client.Close()
	}()

	connector := sshd.NewConnectorImpl(t.Context(), slog.Default(), serverConfig, nil, metrics.NewNoop())

	if err := connector.Handle(serverConn); err == nil {
		t.Fail()
		t.Fatal(err)
	}
}
