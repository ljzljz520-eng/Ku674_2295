package persistence

import (
	"errors"
	"time"

	"go.etcd.io/bbolt"

	"farm-sensor-platform/internal/model"
)

type StoreHealth struct {
	Readings  int
	Batches   int
	Pending   int
	Audits    int
	Open      bool
	CheckedAt time.Time
}

func (s *Store) Health(now time.Time) (StoreHealth, error) {
	s.mu.RLock()
	open := s.db != nil
	s.mu.RUnlock()
	if !open {
		return StoreHealth{CheckedAt: now.UTC()}, errors.New("store is closed")
	}
	readings, err := s.Count(bucketReadings)
	if err != nil {
		return StoreHealth{}, err
	}
	batches, err := s.Count(bucketBatches)
	if err != nil {
		return StoreHealth{}, err
	}
	pending, err := s.Count(bucketPending)
	if err != nil {
		return StoreHealth{}, err
	}
	audits, err := s.Count(bucketAudits)
	if err != nil {
		return StoreHealth{}, err
	}
	return StoreHealth{Readings: readings, Batches: batches, Pending: pending, Audits: audits, Open: true, CheckedAt: now.UTC()}, nil
}

func (s *Store) DeletePending(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketPending).Delete([]byte(id)) })
}

func (s *Store) UpdatePendingStatus(id string, status model.ReadingStatus, now time.Time) error {
	item, err := s.pendingByID(id)
	if err != nil {
		return err
	}
	item.Status = status
	item.UpdatedAt = now.UTC()
	return s.PutPending(item)
}

func (s *Store) pendingByID(id string) (model.PendingItem, error) {
	var item model.PendingItem
	err := s.get(bucketPending, id, &item)
	return item, err
}

func (s *Store) PurgeBefore(cutoff time.Time) (int, error) {
	items, err := s.ListPending("")
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, item := range items {
		if item.UpdatedAt.Before(cutoff) {
			if err := s.DeletePending(item.ID); err != nil {
				return removed, err
			}
			removed++
		}
	}
	return removed, nil
}
