package sensors

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"farm-sensor-platform/internal/adapter"
	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/model"
	"farm-sensor-platform/internal/observability"
	"farm-sensor-platform/internal/persistence"
)

func serviceFixture(t *testing.T) (*Service, *persistence.Store) {
	t.Helper()
	store, err := persistence.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	return NewService(store, observability.NewLogger(), config.Default()), store
}

func validRaw(id string) adapter.RawReading {
	return adapter.RawReading{ExternalID: id, Field: "north", Timestamp: time.Unix(100, 0), Values: map[string]float64{"soil_moisture": 45, "temp_c": 20, "lux": 300, "battery_pct": 88}}
}

func TestPrimaryWorkflow(t *testing.T) {
	service, store := serviceFixture(t)
	defer store.Close()
	batch, err := service.ProcessBatch(context.Background(), BatchRequest{BatchID: "b-primary", ReceivedAt: time.Unix(200, 0), Adapters: []adapter.Adapter{adapter.NewAgriSense([]adapter.RawReading{validRaw("r1")})}})
	if err != nil || batch.Accepted != 1 || batch.Failed {
		t.Fatalf("unexpected batch: %+v %v", batch, err)
	}
}

func TestSensorBatchKeepsFailures(t *testing.T) {
	service, store := serviceFixture(t)
	defer store.Close()
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	bad := validRaw("bad")
	bad.Values["soil_moisture"] = 180
	good := validRaw("good")
	results := make(chan struct {
		batch model.SensorBatch
		err   error
	}, 1)
	go func() {
		batch, err := service.ProcessBatch(context.Background(), BatchRequest{BatchID: "b-bug", ReceivedAt: time.Unix(300, 0), Adapters: []adapter.Adapter{
			adapter.NewBarrierAdapter(adapter.NewAgriSense([]adapter.RawReading{bad}), ready, release),
			adapter.NewBarrierAdapter(adapter.NewFieldLink([]adapter.RawReading{{ExternalID: "good", Field: "north", Timestamp: good.Timestamp, Values: map[string]float64{"moist": 40, "temperature_c": 20, "light_lux": 100, "power": 90}}}), ready, release),
		}})
		results <- struct {
			batch model.SensorBatch
			err   error
		}{batch: batch, err: err}
	}()
	<-ready
	<-ready
	close(release)
	result := <-results
	batch, err := result.batch, result.err
	if err != nil {
		t.Fatal(err)
	}
	if !batch.Failed || batch.Pending != 1 {
		t.Fatalf("expected partial failure, got %+v", batch)
	}
	items, err := service.Pending(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one pending item, got %d", len(items))
	}
}

func TestSecondaryWorkflow(t *testing.T) {
	service, store := serviceFixture(t)
	defer store.Close()
	bad := validRaw("retry")
	bad.Values["soil_moisture"] = 170
	_, err := service.ProcessBatch(context.Background(), BatchRequest{BatchID: "b-retry", ReceivedAt: time.Unix(300, 0), Adapters: []adapter.Adapter{adapter.NewAgriSense([]adapter.RawReading{bad})}})
	if err != nil {
		t.Fatal(err)
	}
	items, err := service.Pending(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("pending: %v %+v", err, items)
	}
	fixed := validRaw("pending:b-retry:agrisense:retry")
	fixed.ExternalID = "retry"
	fixed.Values["soil_moisture"] = 40
	updated, err := service.Retry(context.Background(), items[0].ID, adapter.NewAgriSense([]adapter.RawReading{fixed}), "admin", time.Unix(400, 0))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != model.StatusAccepted {
		t.Fatalf("expected accepted retry: %+v", updated)
	}
}

func TestTertiaryWorkflow(t *testing.T) {
	service, store := serviceFixture(t)
	defer store.Close()
	_, err := service.ProcessBatch(context.Background(), BatchRequest{BatchID: "b-dashboard", ReceivedAt: time.Unix(300, 0), Adapters: []adapter.Adapter{adapter.NewAgriSense([]adapter.RawReading{validRaw("dash")})}})
	if err != nil {
		t.Fatal(err)
	}
	dashboard, err := service.Dashboard(context.Background(), time.Unix(500, 0))
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Health != "healthy" || len(dashboard.Fields) != 1 {
		t.Fatalf("unexpected dashboard %+v", dashboard)
	}
}
