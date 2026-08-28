package alerts

import (
	"fmt"
	"math"
	"sort"

	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/model"
)

type Policy struct {
	MoistureLow     float64
	MoistureHigh    float64
	TemperatureLow  float64
	TemperatureHigh float64
	BatteryLow      float64
	LightHigh       float64
}

func PolicyFromSettings(settings config.Settings) Policy {
	return Policy{MoistureLow: settings.MoistureMinimum, MoistureHigh: settings.MoistureMaximum, TemperatureLow: settings.TemperatureMinimum, TemperatureHigh: settings.TemperatureMaximum, BatteryLow: 15, LightHigh: settings.LightMaximum}
}

func (p Policy) Check(reading model.SensorReading) []Alert {
	checks := []struct {
		code, field      string
		value, low, high float64
	}{
		{"moisture_range", "moisture", reading.Moisture, p.MoistureLow, p.MoistureHigh},
		{"temperature_range", "temperature", reading.Temperature, p.TemperatureLow, p.TemperatureHigh},
		{"light_range", "light", reading.Light, 0, p.LightHigh},
	}
	result := make([]Alert, 0)
	for _, check := range checks {
		if math.IsNaN(check.value) || math.IsInf(check.value, 0) || check.value < check.low || check.value > check.high {
			result = append(result, Alert{FieldID: reading.FieldID, ReadingID: reading.ID, Severity: SeverityCritical, Code: check.code, Message: fmt.Sprintf("%s %.2f outside %.2f..%.2f", check.field, check.value, check.low, check.high)})
		}
	}
	if reading.Battery >= 0 && reading.Battery < p.BatteryLow {
		result = append(result, Alert{FieldID: reading.FieldID, ReadingID: reading.ID, Severity: SeverityWarning, Code: "battery_low", Message: fmt.Sprintf("battery %.2f below %.2f", reading.Battery, p.BatteryLow)})
	}
	return result
}

func (p Policy) CheckMany(readings []model.SensorReading) []Alert {
	result := make([]Alert, 0)
	for _, reading := range readings {
		result = append(result, p.Check(reading)...)
	}
	return result
}

func SortBySeverity(alerts []Alert) []Alert {
	result := append([]Alert(nil), alerts...)
	sort.SliceStable(result, func(i, j int) bool {
		rank := func(value Severity) int {
			switch value {
			case SeverityCritical:
				return 3
			case SeverityWarning:
				return 2
			default:
				return 1
			}
		}
		return rank(result[i].Severity) > rank(result[j].Severity)
	})
	return result
}

func Merge(primary, secondary []Alert) []Alert {
	merged := append(append([]Alert(nil), primary...), secondary...)
	return SortBySeverity(merged)
}
