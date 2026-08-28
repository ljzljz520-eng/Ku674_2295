package api

import (
	"bytes"
	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/observability"
	"farm-sensor-platform/internal/persistence"
	"farm-sensor-platform/internal/sensors"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	store, err := persistence.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	handler := NewHandler(sensors.NewService(store, observability.NewLogger(), config.Default()), observability.NewLogger())
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/health", nil))
	if recorder.Code != 200 {
		t.Fatalf("status %d", recorder.Code)
	}
}

func TestRequestHelpers(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/batches?since=1970-01-01T00:00:01Z", bytes.NewBufferString(`{"name":"fixture"}`))
	req.Header.Set("X-Request-ID", "request-1")
	var target struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(req, &target); err != nil || target.Name != "fixture" {
		t.Fatalf("decode: %v %+v", err, target)
	}
	if got, err := parseSince(req); err != nil || got != time.Unix(1, 0).UTC() {
		t.Fatalf("since: %v %v", got, err)
	}
	if requestID(req) != "request-1" || !methodAllowed(req, "GET", "POST") {
		t.Fatal("request metadata mismatch")
	}
	recorder := httptest.NewRecorder()
	writeJSONRequestError(recorder, req, 400, contextError("bad"))
	if recorder.Code != 400 {
		t.Fatal("error response status")
	}
}

type contextError string

func (e contextError) Error() string { return string(e) }
