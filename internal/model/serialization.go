package model

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type BatchStats struct {
	Total    int            `json:"total"`
	Accepted int            `json:"accepted"`
	Pending  int            `json:"pending"`
	Rejected int            `json:"rejected"`
	ByVendor map[Vendor]int `json:"by_vendor"`
	ByField  map[string]int `json:"by_field"`
}

func (b SensorBatch) Stats() BatchStats {
	stats := BatchStats{ByVendor: make(map[Vendor]int), ByField: make(map[string]int)}
	for _, reading := range b.Readings {
		stats.Total++
		stats.ByVendor[reading.Vendor]++
		stats.ByField[reading.FieldID]++
		switch reading.Status {
		case StatusAccepted:
			stats.Accepted++
		case StatusPending, StatusRetried:
			stats.Pending++
		case StatusRejected:
			stats.Rejected++
		}
	}
	return stats
}

func (b SensorBatch) MarshalSummary() ([]byte, error) {
	stats := b.Stats()
	return json.Marshal(struct {
		BatchID    string     `json:"batch_id"`
		ReceivedAt time.Time  `json:"received_at"`
		Stats      BatchStats `json:"stats"`
		Failed     bool       `json:"failed"`
		Summary    string     `json:"summary"`
	}{b.ID, b.ReceivedAt, stats, b.Failed, b.Summary})
}

func ParseStatus(value string) (ReadingStatus, error) {
	status := ReadingStatus(strings.ToLower(strings.TrimSpace(value)))
	switch status {
	case StatusAccepted, StatusPending, StatusRetried, StatusRejected:
		return status, nil
	default:
		return "", fmt.Errorf("unknown reading status %q", value)
	}
}

func OrderedVendors(readings []SensorReading) []Vendor {
	seen := make(map[Vendor]bool)
	for _, reading := range readings {
		seen[reading.Vendor] = true
	}
	vendors := make([]Vendor, 0, len(seen))
	for vendor := range seen {
		vendors = append(vendors, vendor)
	}
	sort.Slice(vendors, func(i, j int) bool { return vendors[i] < vendors[j] })
	return vendors
}

func (p PendingItem) CanRetry(now time.Time, limit int) bool {
	if p.Status == StatusAccepted || p.Status == StatusRejected {
		return false
	}
	if limit > 0 && p.Attempts >= limit {
		return false
	}
	if now.Before(p.CreatedAt) {
		return false
	}
	return true
}

func (d Dashboard) IsActionable() bool {
	return d.Pending > 0 || d.Health == "empty"
}
