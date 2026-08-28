package alerts

import (
	"testing"
	"time"

	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/model"
)

func alertReading() model.SensorReading {
	return model.SensorReading{ID: "r", FieldID: "field", CapturedAt: time.Unix(10, 0), Moisture: 40, Temperature: 20, Light: 200, Battery: 8, Status: model.StatusAccepted}
}

func TestAlertPoliciesAndSummary(t *testing.T) {
	settings := config.Default()
	reading := alertReading()
	policy := PolicyFromSettings(settings)
	policyAlerts := policy.Check(reading)
	if len(policyAlerts) != 1 || policyAlerts[0].Code != "battery_low" {
		t.Fatalf("policy alerts: %+v", policyAlerts)
	}
	if len(policy.CheckMany([]model.SensorReading{reading})) != 1 {
		t.Fatal("expected check many result")
	}
	items := Evaluate([]model.SensorReading{reading}, settings, time.Unix(20, 0))
	summary := Summarize(items)
	if summary.Total != 1 || summary.Warning != 1 || summary.Healthy {
		t.Fatalf("summary: %+v %+v", items, summary)
	}
	if Highest(items) != SeverityWarning || len(GroupByField(items)["field"]) != 1 {
		t.Fatal("unexpected alert grouping")
	}
	resolved := Resolve(items[0], "admin", time.Unix(21, 0))
	if !resolved.Resolved {
		t.Fatal("expected resolved alert")
	}
	if !Summarize([]Alert{resolved}).Healthy {
		t.Fatal("resolved alert should be healthy")
	}
}

func TestAlertOrderingAndMerge(t *testing.T) {
	input := []Alert{{ID: "i", Severity: SeverityInfo}, {ID: "c", Severity: SeverityCritical}, {ID: "w", Severity: SeverityWarning}}
	if len(Sort(input)) != 3 {
		t.Fatal("sort changed alert count")
	}
	if SortBySeverity(input)[0].Severity != SeverityCritical {
		t.Fatal("severity sort did not prioritize critical")
	}
	if len(Merge(input[:1], input[1:])) != 3 {
		t.Fatal("merge lost alerts")
	}
}
