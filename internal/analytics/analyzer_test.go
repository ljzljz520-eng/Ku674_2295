package analytics

import (
	"testing"
	"time"

	"farm-sensor-platform/internal/model"
)

func analyticReadings() []model.SensorReading {
	return []model.SensorReading{{ID: "a", FieldID: "north", Vendor: model.VendorAgriSense, CapturedAt: time.Unix(10, 0), Moisture: 40, Temperature: 20, Light: 100, Battery: 90, Status: model.StatusAccepted}, {ID: "b", FieldID: "north", Vendor: model.VendorFieldLink, CapturedAt: time.Unix(20, 0), Moisture: 50, Temperature: 22, Light: 120, Battery: 80, Status: model.StatusAccepted}, {ID: "c", FieldID: "south", Vendor: model.VendorFieldLink, CapturedAt: time.Unix(20, 0), Moisture: 50, Temperature: 22, Light: 120, Battery: 80, Status: model.StatusPending}}
}

func TestAnalyticsTrendsCoverageAndFreshness(t *testing.T) {
	readings := analyticReadings()
	trends := Trends(readings)
	if len(trends) != 1 || trends[0].Direction != "rising" {
		t.Fatalf("trends: %+v", trends)
	}
	if len(Coverage(readings, []string{"north", "south"})) != 2 {
		t.Fatal("expected vendor coverage")
	}
	fresh := FreshnessByField(readings, time.Unix(30, 0), 15*time.Second)
	if len(fresh) != 2 || !fresh[0].Fresh {
		t.Fatalf("freshness: %+v", fresh)
	}
	if MoistureAverage(readings) != 140.0/3.0 || TemperatureAverage(readings) != 64.0/3.0 {
		t.Fatal("average mismatch")
	}
	if LightAverage(readings) != 340.0/3.0 || BatteryAverage(readings) != 250.0/3.0 {
		t.Fatal("average mismatch")
	}
	if len(SortByFreshness(readings)) != 3 || len(DistinctVendors(readings)) != 2 {
		t.Fatal("ordering mismatch")
	}
}

func TestAnalyticsHealthAndWindows(t *testing.T) {
	readings := analyticReadings()
	health := ComputeHealth(readings, time.Unix(40, 0), 15*time.Second)
	if health.Label != "attention" || health.Pending != 1 {
		t.Fatalf("health: %+v", health)
	}
	if got, count := WindowAverage(readings, time.Unix(0, 0), time.Unix(30, 0)); got < 46.66 || got > 46.67 || count != 3 {
		t.Fatalf("window average: %v %d", got, count)
	}
	if len(RankFields(readings)) != 1 || RankFields(readings)[0] != "north" {
		t.Fatalf("ranked fields: %v", RankFields(readings))
	}
}
