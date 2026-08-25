package model

import (
	"fmt"
	"strings"
	"time"
)

type Vendor string

const (
	VendorAgriSense Vendor = "agrisense"
	VendorFieldLink Vendor = "fieldlink"
	VendorTerraNode Vendor = "terranode"
)

type ReadingStatus string

const (
	StatusAccepted ReadingStatus = "accepted"
	StatusPending  ReadingStatus = "pending"
	StatusRetried  ReadingStatus = "retried"
	StatusRejected ReadingStatus = "rejected"
)

type SensorReading struct {
	ID          string        `json:"id"`
	BatchID     string        `json:"batch_id"`
	Vendor      Vendor        `json:"vendor"`
	ExternalID  string        `json:"external_id"`
	FieldID     string        `json:"field_id"`
	CapturedAt  time.Time     `json:"captured_at"`
	Moisture    float64       `json:"moisture"`
	Temperature float64       `json:"temperature"`
	Light       float64       `json:"light"`
	Battery     float64       `json:"battery"`
	Status      ReadingStatus `json:"status"`
	Reason      string        `json:"reason,omitempty"`
}

type SensorBatch struct {
	ID         string          `json:"id"`
	ReceivedAt time.Time       `json:"received_at"`
	Readings   []SensorReading `json:"readings"`
	Total      int             `json:"total"`
	Accepted   int             `json:"accepted"`
	Pending    int             `json:"pending"`
	Failed     bool            `json:"failed"`
	Summary    string          `json:"summary"`
}

type PendingItem struct {
	ID        string        `json:"id"`
	ReadingID string        `json:"reading_id"`
	BatchID   string        `json:"batch_id"`
	Reason    string        `json:"reason"`
	Attempts  int           `json:"attempts"`
	Status    ReadingStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type AuditEntry struct {
	ID         string    `json:"id"`
	Action     string    `json:"action"`
	SubjectID  string    `json:"subject_id"`
	Actor      string    `json:"actor"`
	Detail     string    `json:"detail"`
	OccurredAt time.Time `json:"occurred_at"`
}

type Dashboard struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Fields      []FieldSummary `json:"fields"`
	Pending     int            `json:"pending"`
	Accepted    int            `json:"accepted"`
	Health      string         `json:"health"`
}

type FieldSummary struct {
	FieldID            string  `json:"field_id"`
	Readings           int     `json:"readings"`
	AverageMoisture    float64 `json:"average_moisture"`
	AverageTemperature float64 `json:"average_temperature"`
	AverageLight       float64 `json:"average_light"`
	AverageBattery     float64 `json:"average_battery"`
}

func (r SensorReading) Validate() error {
	if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.FieldID) == "" {
		return fmt.Errorf("reading id and field id are required")
	}
	if r.CapturedAt.IsZero() {
		return fmt.Errorf("captured_at is required")
	}
	if r.Vendor == "" || r.ExternalID == "" {
		return fmt.Errorf("vendor and external id are required")
	}
	return nil
}

func (b *SensorBatch) Recalculate() {
	b.Total = len(b.Readings)
	b.Accepted, b.Pending = 0, 0
	for _, reading := range b.Readings {
		if reading.Status == StatusAccepted {
			b.Accepted++
		}
		if reading.Status == StatusPending || reading.Status == StatusRetried {
			b.Pending++
		}
	}
	b.Failed = b.Pending > 0
	if b.Failed {
		b.Summary = fmt.Sprintf("partial failure: %d of %d pending", b.Pending, b.Total)
	} else {
		b.Summary = fmt.Sprintf("all %d readings accepted", b.Total)
	}
}
