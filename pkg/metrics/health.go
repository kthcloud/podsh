package metrics

import "net/http"

type Check func() error

type Health struct {
	checks []Check
}

func NewHealth(checks ...Check) *Health {
	return &Health{
		checks: checks,
	}
}

func (h *Health) Handler(w http.ResponseWriter, r *http.Request) {
	for _, c := range h.checks {
		if err := c(); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
