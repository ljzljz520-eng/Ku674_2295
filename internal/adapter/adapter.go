package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"farm-sensor-platform/internal/model"
)

type RawReading struct {
	ExternalID string
	Field      string
	Timestamp  time.Time
	Values     map[string]float64
}

type Adapter interface {
	Vendor() model.Vendor
	Fetch(context.Context, []string) ([]RawReading, error)
}

type FixtureAdapter struct {
	Name       model.Vendor
	Readings   []RawReading
	FetchError error
}

func (a FixtureAdapter) Vendor() model.Vendor { return a.Name }

func (a FixtureAdapter) Fetch(ctx context.Context, ids []string) ([]RawReading, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if a.FetchError != nil {
		return nil, a.FetchError
	}
	if len(ids) == 0 {
		return append([]RawReading(nil), a.Readings...), nil
	}
	allowed := make(map[string]bool, len(ids))
	for _, id := range ids {
		allowed[id] = true
	}
	result := make([]RawReading, 0, len(a.Readings))
	for _, reading := range a.Readings {
		if allowed[reading.ExternalID] {
			result = append(result, reading)
		}
	}
	return result, nil
}

func NewAgriSense(readings []RawReading) Adapter {
	return FixtureAdapter{Name: model.VendorAgriSense, Readings: readings}
}

func NewFieldLink(readings []RawReading) Adapter {
	return FixtureAdapter{Name: model.VendorFieldLink, Readings: readings}
}

func NewTerraNode(readings []RawReading) Adapter {
	return FixtureAdapter{Name: model.VendorTerraNode, Readings: readings}
}

func NormalizeKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"))
}

func CheckRawReading(r RawReading) error {
	if r.ExternalID == "" || r.Field == "" {
		return fmt.Errorf("raw reading identity is incomplete")
	}
	if r.Timestamp.IsZero() {
		return fmt.Errorf("raw reading timestamp is missing")
	}
	if len(r.Values) == 0 {
		return fmt.Errorf("raw reading has no values")
	}
	return nil
}
