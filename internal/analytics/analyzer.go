package analytics

import (
	"math"
	"sort"
	"strings"
	"time"

	"farm-sensor-platform/internal/model"
)

type FieldTrend struct {
	FieldID           string    `json:"field_id"`
	Samples           int       `json:"samples"`
	First             time.Time `json:"first"`
	Last              time.Time `json:"last"`
	MoistureChange    float64   `json:"moisture_change"`
	TemperatureChange float64   `json:"temperature_change"`
	LightChange       float64   `json:"light_change"`
	BatteryChange     float64   `json:"battery_change"`
	Direction         string    `json:"direction"`
}

type VendorCoverage struct {
	Vendor   model.Vendor `json:"vendor"`
	Fields   int          `json:"fields"`
	Readings int          `json:"readings"`
	LastSeen time.Time    `json:"last_seen"`
	Complete bool         `json:"complete"`
}

type Freshness struct {
	FieldID string        `json:"field_id"`
	Age     time.Duration `json:"age"`
	Fresh   bool          `json:"fresh"`
}

func Trends(readings []model.SensorReading) []FieldTrend {
	groups := make(map[string][]model.SensorReading)
	for _, reading := range readings {
		if reading.Status == model.StatusAccepted {
			groups[reading.FieldID] = append(groups[reading.FieldID], reading)
		}
	}
	result := make([]FieldTrend, 0, len(groups))
	for field, rows := range groups {
		result = append(result, trend(field, rows))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].FieldID < result[j].FieldID })
	return result
}

func trend(field string, readings []model.SensorReading) FieldTrend {
	sort.Slice(readings, func(i, j int) bool { return readings[i].CapturedAt.Before(readings[j].CapturedAt) })
	result := FieldTrend{FieldID: field, Samples: len(readings)}
	if len(readings) == 0 {
		return result
	}
	first, last := readings[0], readings[len(readings)-1]
	result.First, result.Last = first.CapturedAt, last.CapturedAt
	result.MoistureChange = last.Moisture - first.Moisture
	result.TemperatureChange = last.Temperature - first.Temperature
	result.LightChange = last.Light - first.Light
	result.BatteryChange = last.Battery - first.Battery
	result.Direction = direction(result.MoistureChange)
	return result
}

func direction(change float64) string {
	if math.Abs(change) < 0.01 {
		return "steady"
	}
	if change > 0 {
		return "rising"
	}
	return "falling"
}

func Coverage(readings []model.SensorReading, expectedFields []string) []VendorCoverage {
	expected := make(map[string]bool)
	for _, field := range expectedFields {
		expected[strings.TrimSpace(field)] = true
	}
	byVendor := make(map[model.Vendor][]model.SensorReading)
	for _, reading := range readings {
		byVendor[reading.Vendor] = append(byVendor[reading.Vendor], reading)
	}
	result := make([]VendorCoverage, 0, len(byVendor))
	for vendor, rows := range byVendor {
		fields := make(map[string]bool)
		latest := time.Time{}
		for _, row := range rows {
			fields[row.FieldID] = true
			if row.CapturedAt.After(latest) {
				latest = row.CapturedAt
			}
		}
		complete := len(expected) == 0 || len(fields) >= len(expected)
		result = append(result, VendorCoverage{Vendor: vendor, Fields: len(fields), Readings: len(rows), LastSeen: latest, Complete: complete})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Vendor < result[j].Vendor })
	return result
}

func FreshnessByField(readings []model.SensorReading, now time.Time, staleAfter time.Duration) []Freshness {
	latest := make(map[string]time.Time)
	for _, reading := range readings {
		if reading.CapturedAt.After(latest[reading.FieldID]) {
			latest[reading.FieldID] = reading.CapturedAt
		}
	}
	result := make([]Freshness, 0, len(latest))
	for field, captured := range latest {
		age := now.Sub(captured)
		result = append(result, Freshness{FieldID: field, Age: age, Fresh: age <= staleAfter && age >= 0})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].FieldID < result[j].FieldID })
	return result
}

func Average(readings []model.SensorReading, selector func(model.SensorReading) float64) float64 {
	if len(readings) == 0 {
		return 0
	}
	total := 0.0
	for _, reading := range readings {
		total += selector(reading)
	}
	return total / float64(len(readings))
}

func MoistureAverage(readings []model.SensorReading) float64 {
	return Average(readings, func(reading model.SensorReading) float64 { return reading.Moisture })
}
func TemperatureAverage(readings []model.SensorReading) float64 {
	return Average(readings, func(reading model.SensorReading) float64 { return reading.Temperature })
}
func LightAverage(readings []model.SensorReading) float64 {
	return Average(readings, func(reading model.SensorReading) float64 { return reading.Light })
}
func BatteryAverage(readings []model.SensorReading) float64 {
	return Average(readings, func(reading model.SensorReading) float64 { return reading.Battery })
}

func SortByFreshness(readings []model.SensorReading) []model.SensorReading {
	result := append([]model.SensorReading(nil), readings...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].CapturedAt.After(result[j].CapturedAt) })
	return result
}
