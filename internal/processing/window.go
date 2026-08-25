package processing

import (
	"sort"
	"time"

	"farm-sensor-platform/internal/model"
)

type TimeWindow struct {
	Start time.Time
	End   time.Time
}

func NewWindow(start, end time.Time) (TimeWindow, bool) {
	start, end = start.UTC(), end.UTC()
	if start.IsZero() || end.IsZero() || !end.After(start) {
		return TimeWindow{}, false
	}
	return TimeWindow{Start: start, End: end}, true
}

func InWindow(reading model.SensorReading, window TimeWindow) bool {
	return !reading.CapturedAt.Before(window.Start) && reading.CapturedAt.Before(window.End)
}

func FilterReadings(readings []model.SensorReading, window TimeWindow) []model.SensorReading {
	filtered := make([]model.SensorReading, 0)
	for _, reading := range readings {
		if InWindow(reading, window) {
			filtered = append(filtered, reading)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CapturedAt.Before(filtered[j].CapturedAt) })
	return filtered
}

func GroupByField(readings []model.SensorReading) map[string][]model.SensorReading {
	groups := make(map[string][]model.SensorReading)
	for _, reading := range readings {
		groups[reading.FieldID] = append(groups[reading.FieldID], reading)
	}
	return groups
}

func CountPending(readings []model.SensorReading) int {
	count := 0
	for _, reading := range readings {
		if reading.Status == model.StatusPending || reading.Status == model.StatusRetried {
			count++
		}
	}
	return count
}

func LatestByField(readings []model.SensorReading) map[string]model.SensorReading {
	latest := make(map[string]model.SensorReading)
	for _, reading := range readings {
		current, exists := latest[reading.FieldID]
		if !exists || current.CapturedAt.Before(reading.CapturedAt) {
			latest[reading.FieldID] = reading
		}
	}
	return latest
}
