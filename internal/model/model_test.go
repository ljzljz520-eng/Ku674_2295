package model

import (
	"testing"
	"time"
)

func TestReadingValidation(t *testing.T) {
	r := SensorReading{ID: "r1", FieldID: "field-1", Vendor: VendorAgriSense, ExternalID: "x", CapturedAt: time.Unix(10, 0)}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	r.ID = ""
	if err := r.Validate(); err == nil {
		t.Fatal("expected missing id error")
	}
}

func TestBatchRecalculate(t *testing.T) {
	b := SensorBatch{Readings: []SensorReading{{Status: StatusAccepted}, {Status: StatusPending}}}
	b.Recalculate()
	if !b.Failed || b.Accepted != 1 || b.Pending != 1 {
		t.Fatalf("unexpected batch summary: %+v", b)
	}
}

func TestModelSummariesAndStatusHelpers(t *testing.T) {
	b := SensorBatch{ID: "batch", ReceivedAt: time.Unix(1, 0), Readings: []SensorReading{{ID: "a", Vendor: VendorFieldLink, FieldID: "f", Status: StatusAccepted}, {ID: "b", Vendor: VendorAgriSense, FieldID: "g", Status: StatusPending}}}
	stats := b.Stats()
	if stats.Total != 2 || stats.Accepted != 1 || stats.Pending != 1 || len(stats.ByVendor) != 2 {
		t.Fatalf("stats: %+v", stats)
	}
	if _, err := b.MarshalSummary(); err != nil {
		t.Fatal(err)
	}
	if parsed, err := ParseStatus(" RETRIED "); err != nil || parsed != StatusRetried {
		t.Fatalf("status: %v %v", parsed, err)
	}
	if _, err := ParseStatus("unknown"); err == nil {
		t.Fatal("expected status error")
	}
	if len(OrderedVendors(b.Readings)) != 2 {
		t.Fatal("expected ordered vendors")
	}
	pending := PendingItem{ID: "p", Status: StatusPending, CreatedAt: time.Unix(1, 0), Attempts: 1}
	if !pending.CanRetry(time.Unix(2, 0), 3) {
		t.Fatal("expected retry")
	}
	if (Dashboard{Health: "empty"}).IsActionable() == false {
		t.Fatal("empty dashboard should be actionable")
	}
}
