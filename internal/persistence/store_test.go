package persistence

import (
	"path/filepath"
	"testing"
	"time"

	"farm-sensor-platform/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sensor.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	batch := model.SensorBatch{ID: "batch-1", ReceivedAt: time.Unix(2, 0), Summary: "saved"}
	if err := store.PutBatch(batch); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	loaded, err := store.GetBatch("batch-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Summary != batch.Summary {
		t.Fatalf("got %+v", loaded)
	}
}

func TestStoreListsPendingByStatus(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "sensor.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	item := model.PendingItem{ID: "p1", Status: model.StatusPending, CreatedAt: time.Unix(1, 0)}
	if err := store.PutPending(item); err != nil {
		t.Fatal(err)
	}
	items, err := store.ListPending(model.StatusPending)
	if err != nil || len(items) != 1 {
		t.Fatalf("%v %+v", err, items)
	}
}

func TestStoreMaintenanceAndAtomicBatch(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "maintenance.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	reading := model.SensorReading{ID: "r", BatchID: "b", FieldID: "f", Vendor: model.VendorAgriSense, ExternalID: "x", CapturedAt: time.Unix(1, 0), Status: model.StatusAccepted}
	batch := model.SensorBatch{ID: "b", ReceivedAt: time.Unix(1, 0), Readings: []model.SensorReading{reading}}
	pending := model.PendingItem{ID: "p", BatchID: "b", ReadingID: "r", Status: model.StatusPending, CreatedAt: time.Unix(1, 0), UpdatedAt: time.Unix(1, 0)}
	if err := store.SaveBatch(BatchWrite{Batch: batch, Readings: []model.SensorReading{reading}, Pending: []model.PendingItem{pending}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.SnapshotBatch("b"); err != nil {
		t.Fatal(err)
	}
	health, err := store.Health(time.Unix(2, 0))
	if err != nil || !health.Open || health.Readings != 1 {
		t.Fatalf("health: %+v %v", health, err)
	}
	if err := store.UpdatePendingStatus("p", model.StatusRetried, time.Unix(3, 0)); err != nil {
		t.Fatal(err)
	}
	if removed, err := store.PurgeBefore(time.Unix(2, 0)); err != nil || removed != 0 {
		t.Fatalf("purge: %d %v", removed, err)
	}
	if err := store.DeletePending("p"); err != nil {
		t.Fatal(err)
	}
}
