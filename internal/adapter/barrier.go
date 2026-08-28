package adapter

import (
	"context"

	"farm-sensor-platform/internal/model"
)

type BarrierAdapter struct {
	source  Adapter
	ready   chan<- struct{}
	release <-chan struct{}
}

func NewBarrierAdapter(source Adapter, ready chan<- struct{}, release <-chan struct{}) Adapter {
	return BarrierAdapter{source: source, ready: ready, release: release}
}

func (a BarrierAdapter) Vendor() model.Vendor { return a.source.Vendor() }

func (a BarrierAdapter) Fetch(ctx context.Context, ids []string) ([]RawReading, error) {
	select {
	case a.ready <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case <-a.release:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return a.source.Fetch(ctx, ids)
}
