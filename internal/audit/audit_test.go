package audit

import (
	"farm-sensor-platform/internal/persistence"
	"path/filepath"
	"testing"
	"time"
)

func TestRecorderPersistsEntry(t *testing.T) {
	store, err := persistence.Open(filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	r := NewRecorder(store)
	if _, err := r.Record("receive", "batch", "admin", "ok", time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
}
