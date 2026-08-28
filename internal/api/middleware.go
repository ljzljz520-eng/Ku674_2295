package api

import (
	"net/http"
	"time"
)

type responseMetrics struct {
	requests int
	failures int
	total    time.Duration
}

func (h *Handler) withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Service", "farm-sensor-platform")
		next.ServeHTTP(w, r)
		_ = started
	})
}

func (h *Handler) withRequestLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > 2<<20 {
			writeError(w, http.StatusRequestEntityTooLarge, "request body exceeds 2 MiB")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) wrapped() http.Handler { return h.withHeaders(h.withRequestLimit(h.mux)) }
