package persistence

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.etcd.io/bbolt"

	"farm-sensor-platform/internal/model"
)

type BatchWrite struct {
	Batch    model.SensorBatch
	Readings []model.SensorReading
	Pending  []model.PendingItem
}

func (s *Store) SaveBatch(write BatchWrite) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		batchData, err := json.Marshal(write.Batch)
		if err != nil {
			return err
		}
		if err := tx.Bucket(bucketBatches).Put([]byte(write.Batch.ID), batchData); err != nil {
			return err
		}
		for _, reading := range write.Readings {
			data, err := json.Marshal(reading)
			if err != nil {
				return fmt.Errorf("marshal reading: %w", err)
			}
			if err := tx.Bucket(bucketReadings).Put([]byte(reading.ID), data); err != nil {
				return err
			}
		}
		for _, item := range write.Pending {
			data, err := json.Marshal(item)
			if err != nil {
				return fmt.Errorf("marshal pending: %w", err)
			}
			if err := tx.Bucket(bucketPending).Put([]byte(item.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) SnapshotBatch(batchID string) (model.SensorBatch, []model.PendingItem, error) {
	batch, err := s.GetBatch(batchID)
	if err != nil {
		return model.SensorBatch{}, nil, err
	}
	items, err := s.ListPending("")
	if err != nil {
		return model.SensorBatch{}, nil, err
	}
	filtered := make([]model.PendingItem, 0)
	for _, item := range items {
		if item.BatchID == batchID {
			filtered = append(filtered, item)
		}
	}
	return batch, filtered, nil
}
