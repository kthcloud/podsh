package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"
)

type K8sForwarder struct {
	Client kubernetes.Interface
	Config *rest.Config
}

func NewK8sForwarder(c kubernetes.Interface, cfg *rest.Config) *K8sForwarder {
	return &K8sForwarder{
		Client: c,
		Config: cfg,
	}
}

func (k *K8sForwarder) Forward(ctx context.Context, t *Target, in io.Reader, out io.Writer, hostname string, port uint16) error {
	if t == nil {
		return errors.New("target is nil")
	}
	if t.Namespace == "" || t.Pod == "" {
		return errors.New("target namespace/pod required")
	}

	transport, upgrader, err := spdy.RoundTripperFor(k.Config)
	if err != nil {
		return fmt.Errorf("spdy roundtripper: %w", err)
	}

	req := k.Client.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Namespace(t.Namespace).
		Name(t.Pod).
		SubResource("portforward")

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		"POST",
		req.URL(),
	)

	conn, _, err := dialer.Dial("portforward.k8s.io")
	if err != nil {
		return fmt.Errorf("spdy dial: %w", err)
	}
	defer conn.Close()

	portStr := strconv.Itoa(int(port))

	errHeaders := http.Header{}
	errHeaders.Set("StreamType", "error")
	errHeaders.Set("Port", portStr)
	errHeaders.Set("PortForwardRequestID", "1")

	errorStream, err := conn.CreateStream(errHeaders)
	if err != nil {
		return fmt.Errorf("create error stream: %w", err)
	}
	defer errorStream.Close()

	dataHeaders := http.Header{}
	dataHeaders.Set("StreamType", "data")
	dataHeaders.Set("Port", portStr)
	dataHeaders.Set("PortForwardRequestID", "1")

	dataStream, err := conn.CreateStream(dataHeaders)
	if err != nil {
		return fmt.Errorf("create data stream: %w", err)
	}
	defer dataStream.Close()

	var once sync.Once
	closeAll := func() {
		once.Do(func() {
			dataStream.Close()
			errorStream.Close()
			conn.Close()
		})
	}

	go func() {
		select {
		case <-ctx.Done():
			closeAll()
		}
	}()

	remoteErrCh := make(chan error, 1)
	go func() {
		b, _ := io.ReadAll(errorStream)
		if len(b) > 0 {
			remoteErrCh <- fmt.Errorf("k8s portforward error: %s", string(b))
		} else {
			remoteErrCh <- nil
		}
	}()

	copyErrCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(dataStream, in)
		if err != nil && !errors.Is(err, io.EOF) {
			copyErrCh <- err
			return
		}
		copyErrCh <- nil
	}()

	go func() {
		_, err := io.Copy(out, dataStream)
		if err != nil && !errors.Is(err, io.EOF) {
			copyErrCh <- err
			return
		}
		copyErrCh <- nil
	}()

	select {
	case err := <-remoteErrCh:
		closeAll()
		return err

	case err := <-copyErrCh:
		closeAll()
		return err

	case <-ctx.Done():
		closeAll()
		return ctx.Err()
	}
}
