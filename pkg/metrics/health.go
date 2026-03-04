package metrics

import (
	"net/http"
	"sync"
)

type Check func() error

type Health struct {
	checks []Check
	mu     sync.RWMutex
}

func NewHealth() *Health {
	return &Health{}
}

func (h *Health) Add(check Check) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks = append(h.checks, check)
}

func (h *Health) Handler(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.checks {
		if err := c(); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
