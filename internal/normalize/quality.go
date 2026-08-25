package normalize

import (
	"fmt"
	"math"
	"strings"

	"farm-sensor-platform/internal/model"
)

type QualityIssue struct {
	ReadingID string
	Field     string
	Code      string
	Message   string
	Severity  string
}

func Diagnose(reading model.SensorReading) []QualityIssue {
	issues := make([]QualityIssue, 0)
	checks := []struct {
		name  string
		value float64
	}{{"moisture", reading.Moisture}, {"temperature", reading.Temperature}, {"light", reading.Light}, {"battery", reading.Battery}}
	for _, check := range checks {
		if math.IsNaN(check.value) || math.IsInf(check.value, 0) {
			issues = append(issues, QualityIssue{reading.ID, check.name, "not_finite", "value is not finite", "critical"})
		}
	}
	if strings.TrimSpace(reading.FieldID) == "" {
		issues = append(issues, QualityIssue{reading.ID, "field", "missing", "field identifier is missing", "critical"})
	}
	if reading.Battery >= 0 && reading.Battery < 10 {
		issues = append(issues, QualityIssue{reading.ID, "battery", "low", fmt.Sprintf("battery %.1f%% is low", reading.Battery), "warning"})
	}
	if reading.CapturedAt.IsZero() {
		issues = append(issues, QualityIssue{reading.ID, "captured_at", "missing", "capture time is missing", "critical"})
	}
	return issues
}

func QualityLabel(reading model.SensorReading) string {
	issues := Diagnose(reading)
	for _, issue := range issues {
		if issue.Severity == "critical" {
			return "invalid"
		}
	}
	if len(issues) > 0 {
		return "warning"
	}
	return "good"
}

func Explain(issues []QualityIssue) string {
	if len(issues) == 0 {
		return "no quality issues"
	}
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		parts = append(parts, issue.Field+": "+issue.Message)
	}
	return strings.Join(parts, "; ")
}

func ShouldPend(reading model.SensorReading) bool {
	return reading.Status == model.StatusPending || len(Diagnose(reading)) > 0
}
