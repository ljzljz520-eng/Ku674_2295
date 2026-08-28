package analytics

import (
	"math"
	"sort"
	"time"

	"farm-sensor-platform/internal/model"
	"farm-sensor-platform/internal/processing"
)

type HealthIndex struct {
	Score   float64 `json:"score"`
	Label   string  `json:"label"`
	Pending int     `json:"pending"`
	Stale   int     `json:"stale"`
	Fields  int     `json:"fields"`
}

func ComputeHealth(readings []model.SensorReading, now time.Time, staleAfter time.Duration) HealthIndex {
	fields := make(map[string]bool)
	pending, stale := 0, 0
	for _, reading := range readings {
		fields[reading.FieldID] = true
		if reading.Status == model.StatusPending || reading.Status == model.StatusRetried {
			pending++
		}
		if now.After(reading.CapturedAt.Add(staleAfter)) {
			stale++
		}
	}
	score := 100.0
	score -= float64(pending * 15)
	score -= float64(stale * 10)
	if score < 0 {
		score = 0
	}
	if math.IsNaN(score) {
		score = 0
	}
	label := "healthy"
	if score < 70 {
		label = "attention"
	}
	if score < 40 {
		label = "critical"
	}
	return HealthIndex{Score: score, Label: label, Pending: pending, Stale: stale, Fields: len(fields)}
}

func RankFields(readings []model.SensorReading) []string {
	counts := make(map[string]int)
	for _, reading := range readings {
		if reading.Status == model.StatusAccepted {
			counts[reading.FieldID]++
		}
	}
	fields := make([]string, 0, len(counts))
	for field := range counts {
		fields = append(fields, field)
	}
	sort.Slice(fields, func(i, j int) bool {
		if counts[fields[i]] != counts[fields[j]] {
			return counts[fields[i]] > counts[fields[j]]
		}
		return fields[i] < fields[j]
	})
	return fields
}

func WindowAverage(readings []model.SensorReading, start, end time.Time) (float64, int) {
	window, ok := processing.NewWindow(start, end)
	if !ok {
		return 0, 0
	}
	rows := processing.FilterReadings(readings, window)
	return MoistureAverage(rows), len(rows)
}

func DistinctVendors(readings []model.SensorReading) []model.Vendor {
	seen := make(map[model.Vendor]bool)
	for _, reading := range readings {
		seen[reading.Vendor] = true
	}
	vendors := make([]model.Vendor, 0, len(seen))
	for vendor := range seen {
		vendors = append(vendors, vendor)
	}
	sort.Slice(vendors, func(i, j int) bool { return vendors[i] < vendors[j] })
	return vendors
}
