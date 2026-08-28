package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"farm-sensor-platform/internal/observability"
	"farm-sensor-platform/internal/sensors"
)

type Handler struct {
	service *sensors.Service
	logger  *observability.Logger
	mux     *http.ServeMux
}

func NewHandler(service *sensors.Service, logger *observability.Logger) http.Handler {
	h := &Handler{service: service, logger: logger, mux: http.NewServeMux()}
	h.routes()
	return h
}

func (h *Handler) routes() {
	h.mux.HandleFunc("/health", h.health)
	h.mux.HandleFunc("/v1/batches", h.batches)
	h.mux.HandleFunc("/v1/pending", h.pending)
	h.mux.HandleFunc("/v1/dashboard", h.dashboard)
	h.mux.HandleFunc("/v1/report", h.report)
	h.mux.HandleFunc("/v1/fields", h.fields)
	h.mux.HandleFunc("/v1/trends", h.trends)
	h.mux.HandleFunc("/v1/alerts", h.alerts)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.wrapped().ServeHTTP(w, r) }

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	health, err := h.service.Health(time.Unix(0, 0))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "storage": health})
}

func (h *Handler) batches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	items, err := h.service.Batches(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) pending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	items, err := h.service.Pending(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	dashboard, err := h.service.Dashboard(r.Context(), time.Unix(0, 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	report, err := h.service.Operations().BatchReport(r.Context(), time.Unix(0, 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *Handler) fields(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	fields, err := h.service.Operations().Fields(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, fields)
}

func (h *Handler) trends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	trends, err := h.service.Operations().Trends(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, trends)
}

func (h *Handler) alerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	items, summary, err := h.service.Operations().Alerts(r.Context(), time.Unix(0, 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "summary": summary})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": strings.TrimSpace(message)})
}
