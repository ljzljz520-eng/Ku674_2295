package normalize

import (
	"testing"
	"time"

	"farm-sensor-platform/internal/adapter"
	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/model"
)

func TestMapperCanonicalizesVendorFields(t *testing.T) {
	m := New(config.Default())
	r, err := m.Convert("b1", model.VendorFieldLink, adapter.RawReading{ExternalID: "x", Field: "north", Timestamp: time.Unix(1, 0), Values: map[string]float64{"moist": 42, "temperature_c": 20, "light_lux": 100, "power": 90}})
	if err != nil || r.Moisture != 42 || r.Status != model.StatusAccepted {
		t.Fatalf("unexpected normalized reading: %+v %v", r, err)
	}
}

func TestMapperMarksOutOfRangeAsPending(t *testing.T) {
	m := New(config.Default())
	r, err := m.Convert("b1", model.VendorAgriSense, adapter.RawReading{ExternalID: "x", Field: "north", Timestamp: time.Unix(1, 0), Values: map[string]float64{"soil_moisture": 142, "temp_c": 20, "lux": 100, "battery_pct": 90}})
	if err != nil || r.Status != model.StatusPending {
		t.Fatalf("expected pending reading: %+v %v", r, err)
	}
}

func TestQualityHelpers(t *testing.T) {
	reading := model.SensorReading{ID: "r", FieldID: "f", CapturedAt: time.Unix(1, 0), Battery: 5, Status: model.StatusAccepted}
	issues := Diagnose(reading)
	if QualityLabel(reading) != "warning" || Explain(issues) == "" || !ShouldPend(reading) {
		t.Fatalf("quality: %+v", issues)
	}
	if CaptureTime(time.Time{}).Unix() != 0 || CaptureTime(time.Unix(2, 0)).Location() != time.UTC {
		t.Fatal("capture time normalization failed")
	}
}
