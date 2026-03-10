package k8s

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	metricsConstants "github.com/kthcloud/podsh/internal/metrics"
	"github.com/kthcloud/podsh/pkg/metrics"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"
)

// ForwardManager manages multiple port-forward streams over a single SPDY connection.
type ForwardManager struct {
	ctx           context.Context
	conn          httpstream.Connection
	mu            sync.Mutex
	reqID         uint64
	closed        bool
	closeOnce     sync.Once
	activeStreams map[string]httpstream.Stream

	namespace, pod string
	port           int
	client         kubernetes.Interface
	config         *rest.Config

	metrics metrics.Metrics
}

// NewForwardManager creates a new manager for a specific pod.
func NewForwardManager(ctx context.Context, client kubernetes.Interface, config *rest.Config, namespace, pod string, port int, metrics metrics.Metrics) (*ForwardManager, error) {
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, fmt.Errorf("spdy roundtripper: %w", err)
	}

	req := client.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("portforward")

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		"POST",
		req.URL(),
	)

	conn, _, err := dialer.Dial("portforward.k8s.io")
	if err != nil {
		return nil, fmt.Errorf("spdy dial: %w", err)
	}

	metrics.Gauge(metricsConstants.PodshK8sActiveTunnelForwards).Inc()

	return &ForwardManager{
		ctx:           ctx,
		conn:          conn,
		namespace:     namespace,
		pod:           pod,
		port:          port,
		client:        client,
		config:        config,
		activeStreams: make(map[string]httpstream.Stream),
		metrics:       metrics,
	}, nil
}

// Forward sets up a single port forward using new streams for in/out.
func (fm *ForwardManager) Forward(in io.Reader, out io.Writer) error {
	fm.mu.Lock()
	if fm.closed {
		fm.mu.Unlock()
		return fmt.Errorf("forward manager is closed")
	}
	fm.mu.Unlock()

	id := strconv.FormatUint(atomic.AddUint64(&fm.reqID, 1), 10)
	portStr := strconv.Itoa(fm.port)

	// Create error stream
	fm.mu.Lock()
	errHeaders := http.Header{}
	errHeaders.Set("StreamType", "error")
	errHeaders.Set("Port", portStr)
	errHeaders.Set("PortForwardRequestID", id)

	errorStream, err := fm.conn.CreateStream(errHeaders)
	if err != nil {
		fm.mu.Unlock()
		return fmt.Errorf("create error stream: %w", err)
	}
	fm.activeStreams[id+"-error"] = errorStream

	// Create data stream
	dataHeaders := http.Header{}
	dataHeaders.Set("StreamType", "data")
	dataHeaders.Set("Port", portStr)
	dataHeaders.Set("PortForwardRequestID", id)

	dataStream, err := fm.conn.CreateStream(dataHeaders)
	if err != nil {
		errorStream.Close()
		delete(fm.activeStreams, id+"-error")
		fm.mu.Unlock()
		return fmt.Errorf("create data stream: %w", err)
	}
	fm.activeStreams[id+"-data"] = dataStream
	fm.mu.Unlock()

	wait, cancel := context.WithCancel(fm.ctx)
	defer cancel()

	fm.metrics.Gauge(metricsConstants.PodshK8sActiveTunnelStreams).Inc()

	// Cleanup function
	var once sync.Once
	closeAll := func() {
		once.Do(func() {
			defer func() {
				fm.metrics.Gauge(metricsConstants.PodshK8sActiveTunnelStreams).Dec()
			}()
			fm.mu.Lock()
			defer fm.mu.Unlock()
			dataStream.Close()
			errorStream.Close()
			delete(fm.activeStreams, id+"-data")
			delete(fm.activeStreams, id+"-error")
		})
	}

	go func() {
		<-wait.Done()
		closeAll()
	}()

	// Read Kubernetes error stream
	remoteErrCh := make(chan error, 1)
	go func() {
		b, _ := io.ReadAll(errorStream)
		if len(b) > 0 {
			remoteErrCh <- fmt.Errorf("k8s portforward error: %s", string(b))
		} else {
			remoteErrCh <- nil
		}
	}()

	// Copy data
	copyErrCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(dataStream, in)
		copyErrCh <- err
	}()
	go func() {
		_, err := io.Copy(out, dataStream)
		copyErrCh <- err
	}()

	// Wait for completion
	select {
	case err := <-remoteErrCh:
		closeAll()
		return err
	case err := <-copyErrCh:
		closeAll()
		return err
	case <-fm.ctx.Done():
		closeAll()
		return fm.ctx.Err()
	}
}

// Close shuts down all active streams and the underlying SPDY connection.
func (fm *ForwardManager) Close() error {
	fm.closeOnce.Do(func() {
		fm.mu.Lock()
		defer fm.mu.Unlock()
		fm.closed = true

		for _, s := range fm.activeStreams {
			s.Close()
		}
		fm.activeStreams = make(map[string]httpstream.Stream)

		if fm.conn != nil {
			fm.conn.Close()
		}

		fm.metrics.Gauge(metricsConstants.PodshK8sActiveTunnelForwards).Dec()
	})
	return nil
}
