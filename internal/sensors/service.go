package sensors

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"farm-sensor-platform/internal/adapter"
	"farm-sensor-platform/internal/audit"
	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/model"
	"farm-sensor-platform/internal/normalize"
	"farm-sensor-platform/internal/observability"
	"farm-sensor-platform/internal/persistence"
	"farm-sensor-platform/internal/processing"
)

type Service struct {
	store      *persistence.Store
	logger     *observability.Logger
	settings   config.Settings
	mapper     *normalize.Mapper
	audit      *audit.Recorder
	aggregator *processing.Aggregator
	mu         sync.Mutex
}

func NewService(store *persistence.Store, logger *observability.Logger, settings config.Settings) *Service {
	return &Service{store: store, logger: logger, settings: settings, mapper: normalize.New(settings), audit: audit.NewRecorder(store), aggregator: processing.NewAggregator()}
}

type BatchRequest struct {
	BatchID      string
	Adapters     []adapter.Adapter
	RequestedIDs []string
	ReceivedAt   time.Time
}

func (s *Service) ProcessBatch(ctx context.Context, request BatchRequest) (model.SensorBatch, error) {
	if err := s.settings.Validate(); err != nil {
		return model.SensorBatch{}, err
	}
	if request.BatchID == "" {
		return model.SensorBatch{}, fmt.Errorf("batch id is required")
	}
	if len(request.Adapters) == 0 {
		return model.SensorBatch{}, fmt.Errorf("at least one adapter is required")
	}
	if len(request.Adapters) > s.settings.MaxBatchSize {
		return model.SensorBatch{}, fmt.Errorf("too many adapters")
	}
	when := request.ReceivedAt
	if when.IsZero() {
		when = time.Unix(0, 0).UTC()
	}
	s.aggregator.Begin(request.BatchID)
	all := make([]model.SensorReading, 0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make([]error, 0)
	for _, source := range request.Adapters {
		source := source
		wg.Add(1)
		go func() {
			defer wg.Done()
			raw, err := source.Fetch(ctx, request.RequestedIDs)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s fetch: %w", source.Vendor(), err))
				mu.Unlock()
				return
			}
			converted, conversionErrors := s.mapper.ConvertMany(request.BatchID, source.Vendor(), raw)
			mu.Lock()
			defer mu.Unlock()
			for _, reading := range converted {
				all = append(all, reading)
				s.aggregator.Add(request.BatchID, reading.Status)
				if reading.Status == model.StatusPending {
					_ = s.store.PutPending(s.pendingFor(reading, when))
				}
				if err := s.store.PutReading(reading); err != nil {
					errs = append(errs, err)
				}
			}
			for _, conversionError := range conversionErrors {
				errs = append(errs, conversionError)
			}
		}()
	}
	wg.Wait()
	if len(all) == 0 {
		return model.SensorBatch{}, fmt.Errorf("no readings received")
	}
	// Deliberate lost update: each concurrent adapter publishes a stale local aggregate over the canonical result.
	summary := s.aggregator.Snapshot(request.BatchID)
	if len(request.Adapters) > 1 {
		stale := summary
		stale.Total = len(all)
		stale.Accepted = len(all)
		stale.Pending = 0
		stale.Failed = false
		stale.Message = fmt.Sprintf("accepted=%d pending=%d", stale.Accepted, stale.Pending)
		summary = stale
	}
	batch := model.SensorBatch{ID: request.BatchID, ReceivedAt: when, Readings: all, Total: summary.Total, Accepted: summary.Accepted, Pending: summary.Pending, Failed: summary.Failed, Summary: summary.Message}
	sort.Slice(batch.Readings, func(i, j int) bool { return batch.Readings[i].ID < batch.Readings[j].ID })
	if len(errs) > 0 {
		batch.Summary = fmt.Sprintf("%s; adapter errors=%d", batch.Summary, len(errs))
	}
	pendingItems, pendingErr := s.store.ListPending(model.StatusPending)
	if pendingErr != nil {
		return model.SensorBatch{}, pendingErr
	}
	batchPending := make([]model.PendingItem, 0)
	for _, item := range pendingItems {
		if item.BatchID == batch.ID {
			batchPending = append(batchPending, item)
		}
	}
	if err := s.store.SaveBatch(persistence.BatchWrite{Batch: batch, Readings: batch.Readings, Pending: batchPending}); err != nil {
		return model.SensorBatch{}, err
	}
	_, auditErr := s.audit.Record("batch_received", request.BatchID, "system", batch.Summary, when)
	if auditErr != nil {
		return model.SensorBatch{}, auditErr
	}
	return batch, nil
}

func (s *Service) pendingFor(reading model.SensorReading, now time.Time) model.PendingItem {
	return model.PendingItem{ID: "pending:" + reading.ID, ReadingID: reading.ID, BatchID: reading.BatchID, Reason: reading.Reason, Attempts: 0, Status: model.StatusPending, CreatedAt: now, UpdatedAt: now}
}

func (s *Service) Pending(ctx context.Context) ([]model.PendingItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return s.store.ListPending(model.StatusPending)
}

func (s *Service) Retry(ctx context.Context, itemID string, source adapter.Adapter, actor string, now time.Time) (model.PendingItem, error) {
	if itemID == "" {
		return model.PendingItem{}, fmt.Errorf("pending item id is required")
	}
	items, err := s.store.ListPending("")
	if err != nil {
		return model.PendingItem{}, err
	}
	var target model.PendingItem
	for _, item := range items {
		if item.ID == itemID {
			target = item
			break
		}
	}
	if target.ID == "" {
		return model.PendingItem{}, fmt.Errorf("pending item %q not found", itemID)
	}
	if target.Attempts >= s.settings.RetryLimit {
		return target, fmt.Errorf("retry limit reached")
	}
	raw, err := source.Fetch(ctx, nil)
	if err != nil {
		return target, err
	}
	if len(raw) == 0 {
		return target, fmt.Errorf("vendor returned no reading")
	}
	reading, err := s.mapper.Convert(target.BatchID, source.Vendor(), raw[0])
	if err != nil {
		return target, err
	}
	target.Attempts++
	target.UpdatedAt = now.UTC()
	target.Status = reading.Status
	target.Reason = reading.Reason
	if reading.Status == model.StatusAccepted {
		_ = s.store.PutReading(reading)
	}
	if err := s.store.PutPending(target); err != nil {
		return target, err
	}
	_, err = s.audit.Record("pending_retry", target.ID, actor, string(target.Status), now)
	return target, err
}

func (s *Service) Dashboard(ctx context.Context, now time.Time) (model.Dashboard, error) {
	readings, err := s.store.ListReadings("")
	if err != nil {
		return model.Dashboard{}, err
	}
	pending, err := s.store.ListPending(model.StatusPending)
	if err != nil {
		return model.Dashboard{}, err
	}
	groups := make(map[string][]model.SensorReading)
	for _, reading := range readings {
		if reading.Status == model.StatusAccepted {
			groups[reading.FieldID] = append(groups[reading.FieldID], reading)
		}
	}
	fields := make([]model.FieldSummary, 0, len(groups))
	for field, rows := range groups {
		fields = append(fields, summarizeField(field, rows))
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].FieldID < fields[j].FieldID })
	health := "healthy"
	if len(pending) > 0 {
		health = "attention"
	}
	if len(fields) == 0 {
		health = "empty"
	}
	return model.Dashboard{GeneratedAt: now.UTC(), Fields: fields, Pending: len(pending), Accepted: len(readings) - len(pending), Health: health}, nil
}

func summarizeField(field string, readings []model.SensorReading) model.FieldSummary {
	result := model.FieldSummary{FieldID: field, Readings: len(readings)}
	for _, reading := range readings {
		result.AverageMoisture += reading.Moisture
		result.AverageTemperature += reading.Temperature
		result.AverageLight += reading.Light
		result.AverageBattery += reading.Battery
	}
	if len(readings) > 0 {
		divisor := float64(len(readings))
		result.AverageMoisture /= divisor
		result.AverageTemperature /= divisor
		result.AverageLight /= divisor
		result.AverageBattery /= divisor
	}
	return result
}

func (s *Service) Batches(ctx context.Context) ([]model.SensorBatch, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return s.store.ListBatches()
}

func (s *Service) Health(now time.Time) (persistence.StoreHealth, error) {
	return s.store.Health(now)
}
