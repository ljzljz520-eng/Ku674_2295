package processing

import (
	"testing"
	"time"

	"farm-sensor-platform/internal/model"
)

func TestAggregatorCountsStatuses(t *testing.T) {
	a := NewAggregator()
	a.Begin("b")
	a.Add("b", model.StatusAccepted)
	a.Add("b", model.StatusPending)
	s := a.Snapshot("b")
	if s.Total != 2 || !s.Failed {
		t.Fatalf("unexpected summary %+v", s)
	}
}

func TestReportsWindowsAndGrouping(t *testing.T) {
	now := time.Unix(100, 0)
	rows := []model.SensorReading{{ID: "a", FieldID: "f", Vendor: model.VendorAgriSense, CapturedAt: time.Unix(10, 0), Status: model.StatusAccepted}, {ID: "b", FieldID: "f", Vendor: model.VendorFieldLink, CapturedAt: time.Unix(20, 0), Status: model.StatusPending}}
	report := BuildReport([]model.SensorBatch{{ID: "b", Readings: rows}}, now)
	if report.Readings != 2 || report.Status() != "partial" || report.Complete() {
		t.Fatalf("report: %+v", report)
	}
	if report.VendorLine() == "" || report.FieldLine() != "f" {
		t.Fatalf("report lines: %+v", report)
	}
	window, ok := NewWindow(time.Unix(0, 0), time.Unix(30, 0))
	if !ok || len(FilterReadings(rows, window)) != 2 {
		t.Fatal("window filtering failed")
	}
	if len(GroupByField(rows)["f"]) != 2 || CountPending(rows) != 1 || len(LatestByField(rows)) != 1 {
		t.Fatal("grouping failed")
	}
}
