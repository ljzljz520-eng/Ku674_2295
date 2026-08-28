package normalize

import (
	"fmt"
	"math"
	"strings"
	"time"

	"farm-sensor-platform/internal/adapter"
	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/model"
)

type Mapper struct {
	settings config.Settings
	aliases  map[model.Vendor]map[string]string
}

func New(settings config.Settings) *Mapper {
	return &Mapper{settings: settings, aliases: map[model.Vendor]map[string]string{
		model.VendorAgriSense: {"soil_moisture": "moisture", "temp_c": "temperature", "lux": "light", "battery_pct": "battery"},
		model.VendorFieldLink: {"moist": "moisture", "temperature_c": "temperature", "light_lux": "light", "power": "battery"},
		model.VendorTerraNode: {"m": "moisture", "t": "temperature", "l": "light", "b": "battery"},
	}}
}

func (m *Mapper) Convert(batchID string, vendor model.Vendor, raw adapter.RawReading) (model.SensorReading, error) {
	if err := adapter.CheckRawReading(raw); err != nil {
		return model.SensorReading{}, err
	}
	aliases, ok := m.aliases[vendor]
	if !ok {
		return model.SensorReading{}, fmt.Errorf("unsupported vendor %q", vendor)
	}
	values := map[string]float64{}
	for key, value := range raw.Values {
		canonical := aliases[adapter.NormalizeKey(key)]
		if canonical != "" {
			values[canonical] = value
		}
	}
	reading := model.SensorReading{ID: batchID + ":" + string(vendor) + ":" + raw.ExternalID, BatchID: batchID, Vendor: vendor, ExternalID: raw.ExternalID, FieldID: strings.TrimSpace(raw.Field), CapturedAt: raw.Timestamp, Moisture: values["moisture"], Temperature: values["temperature"], Light: values["light"], Battery: values["battery"], Status: model.StatusAccepted}
	if err := reading.Validate(); err != nil {
		return model.SensorReading{}, err
	}
	if err := m.ValidateValues(reading); err != nil {
		reading.Status = model.StatusPending
		reading.Reason = err.Error()
	}
	return reading, nil
}

func (m *Mapper) ValidateValues(r model.SensorReading) error {
	checks := []struct {
		name             string
		value, low, high float64
	}{
		{"moisture", r.Moisture, m.settings.MoistureMinimum, m.settings.MoistureMaximum},
		{"temperature", r.Temperature, m.settings.TemperatureMinimum, m.settings.TemperatureMaximum},
		{"light", r.Light, 0, m.settings.LightMaximum},
		{"battery", r.Battery, m.settings.BatteryMinimum, 100},
	}
	for _, check := range checks {
		if math.IsNaN(check.value) || math.IsInf(check.value, 0) || check.value < check.low || check.value > check.high {
			return fmt.Errorf("%s value %.2f outside permitted range", check.name, check.value)
		}
	}
	return nil
}

func (m *Mapper) ConvertMany(batchID string, vendor model.Vendor, raw []adapter.RawReading) ([]model.SensorReading, []error) {
	result := make([]model.SensorReading, 0, len(raw))
	errors := make([]error, 0)
	for _, item := range raw {
		reading, err := m.Convert(batchID, vendor, item)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		result = append(result, reading)
	}
	return result, errors
}

func CaptureTime(now time.Time) time.Time {
	if now.IsZero() {
		return time.Unix(0, 0).UTC()
	}
	return now.UTC()
}
