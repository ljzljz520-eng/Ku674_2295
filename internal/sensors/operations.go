package sensors

import (
	"context"
	"fmt"
	"sort"
	"time"

	"farm-sensor-platform/internal/adapter"
	"farm-sensor-platform/internal/alerts"
	"farm-sensor-platform/internal/analytics"
	"farm-sensor-platform/internal/model"
	"farm-sensor-platform/internal/normalize"
	"farm-sensor-platform/internal/processing"
)

type Operations struct{ service *Service }

func (s *Service) Operations() *Operations { return &Operations{service: s} }

func (o *Operations) BatchReport(ctx context.Context, now time.Time) (processing.Report, error) {
	batches, err := o.service.Batches(ctx)
	if err != nil {
		return processing.Report{}, err
	}
	return processing.BuildReport(batches, now), nil
}

func (o *Operations) Preview(ctx context.Context, batchID string, source adapter.Adapter, raw []adapter.RawReading) ([]model.SensorReading, []normalize.QualityIssue, error) {
	if batchID == "" || source == nil {
		return nil, nil, fmt.Errorf("batch id and adapter are required")
	}
	readings, errorsFound := o.service.mapper.ConvertMany(batchID, source.Vendor(), raw)
	if len(errorsFound) > 0 {
		return readings, nil, fmt.Errorf("preview conversion failed: %v", errorsFound)
	}
	issues := make([]normalize.QualityIssue, 0)
	for _, reading := range readings {
		issues = append(issues, normalize.Diagnose(reading)...)
	}
	return readings, issues, nil
}

func (o *Operations) Fields(ctx context.Context) ([]string, error) {
	readings, err := o.service.store.ListReadings("")
	if err != nil {
		return nil, err
	}
	fields := make(map[string]bool)
	for _, reading := range readings {
		fields[reading.FieldID] = true
	}
	result := make([]string, 0, len(fields))
	for field := range fields {
		result = append(result, field)
	}
	sort.Strings(result)
	return result, nil
}

func (o *Operations) Reconcile(ctx context.Context, batchID string, now time.Time) (model.SensorBatch, error) {
	batch, err := o.service.store.GetBatch(batchID)
	if err != nil {
		return model.SensorBatch{}, err
	}
	batch.Recalculate()
	batch.ReceivedAt = now.UTC()
	if err := o.service.store.PutBatch(batch); err != nil {
		return model.SensorBatch{}, err
	}
	return batch, nil
}

func (o *Operations) Stale(now time.Time, reading model.SensorReading) bool {
	return now.After(reading.CapturedAt.Add(o.service.settings.StaleAfter))
}

func (o *Operations) Trends(ctx context.Context) ([]analytics.FieldTrend, error) {
	readings, err := o.service.store.ListReadings("")
	if err != nil {
		return nil, err
	}
	return analytics.Trends(readings), nil
}

func (o *Operations) Alerts(ctx context.Context, now time.Time) ([]alerts.Alert, alerts.Summary, error) {
	readings, err := o.service.store.ListReadings("")
	if err != nil {
		return nil, alerts.Summary{}, err
	}
	items := alerts.Evaluate(readings, o.service.settings, now)
	return items, alerts.Summarize(items), nil
}
