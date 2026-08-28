package alerts

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/model"
	"farm-sensor-platform/internal/normalize"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type Alert struct {
	ID        string    `json:"id"`
	FieldID   string    `json:"field_id"`
	ReadingID string    `json:"reading_id"`
	Severity  Severity  `json:"severity"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Resolved  bool      `json:"resolved"`
}

type Summary struct {
	Total    int      `json:"total"`
	Info     int      `json:"info"`
	Warning  int      `json:"warning"`
	Critical int      `json:"critical"`
	Fields   []string `json:"fields"`
	Healthy  bool     `json:"healthy"`
}

func Evaluate(readings []model.SensorReading, settings config.Settings, now time.Time) []Alert {
	alerts := make([]Alert, 0)
	for _, reading := range readings {
		issues := normalize.Diagnose(reading)
		for index, issue := range issues {
			severity := SeverityWarning
			if issue.Severity == "critical" {
				severity = SeverityCritical
			}
			alerts = append(alerts, Alert{ID: fmt.Sprintf("alert:%s:%d", reading.ID, index), FieldID: reading.FieldID, ReadingID: reading.ID, Severity: severity, Code: issue.Code, Message: issue.Message, CreatedAt: now.UTC()})
		}
		if reading.Status == model.StatusPending && len(issues) == 0 {
			alerts = append(alerts, Alert{ID: "pending:" + reading.ID, FieldID: reading.FieldID, ReadingID: reading.ID, Severity: SeverityWarning, Code: "pending", Message: "reading is awaiting review", CreatedAt: now.UTC()})
		}
	}
	return alerts
}

func Summarize(alerts []Alert) Summary {
	fields := make(map[string]bool)
	summary := Summary{Total: len(alerts)}
	for _, alert := range alerts {
		if alert.Resolved {
			continue
		}
		fields[alert.FieldID] = true
		switch alert.Severity {
		case SeverityInfo:
			summary.Info++
		case SeverityWarning:
			summary.Warning++
		case SeverityCritical:
			summary.Critical++
		}
	}
	for field := range fields {
		summary.Fields = append(summary.Fields, field)
	}
	sort.Strings(summary.Fields)
	summary.Healthy = summary.Warning == 0 && summary.Critical == 0
	return summary
}

func Resolve(alert Alert, actor string, now time.Time) Alert {
	alert.Resolved = true
	alert.Message = strings.TrimSpace(alert.Message) + "; resolved by " + strings.TrimSpace(actor)
	alert.CreatedAt = now.UTC()
	return alert
}

func Highest(alerts []Alert) Severity {
	highest := SeverityInfo
	for _, alert := range alerts {
		if alert.Resolved {
			continue
		}
		if alert.Severity == SeverityCritical {
			return SeverityCritical
		}
		if alert.Severity == SeverityWarning {
			highest = SeverityWarning
		}
	}
	return highest
}

func GroupByField(alerts []Alert) map[string][]Alert {
	groups := make(map[string][]Alert)
	for _, alert := range alerts {
		groups[alert.FieldID] = append(groups[alert.FieldID], alert)
	}
	return groups
}

func Sort(alerts []Alert) []Alert {
	result := append([]Alert(nil), alerts...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Severity != result[j].Severity {
			return result[i].Severity > result[j].Severity
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}
