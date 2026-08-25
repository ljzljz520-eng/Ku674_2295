package persistence

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"go.etcd.io/bbolt"

	"farm-sensor-platform/internal/model"
)

var (
	bucketReadings = []byte("readings")
	bucketBatches  = []byte("batches")
	bucketPending  = []byte("pending")
	bucketAudits   = []byte("audits")
)

type Store struct {
	db *bbolt.DB
	mu sync.RWMutex
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	db, err := bbolt.Open(path, 0o600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &Store{db: db}
	if err := store.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{bucketReadings, bucketBatches, bucketPending, bucketAudits} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) PutReading(reading model.SensorReading) error {
	return s.put(bucketReadings, reading.ID, reading)
}
func (s *Store) PutBatch(batch model.SensorBatch) error  { return s.put(bucketBatches, batch.ID, batch) }
func (s *Store) PutPending(item model.PendingItem) error { return s.put(bucketPending, item.ID, item) }
func (s *Store) PutAudit(entry model.AuditEntry) error   { return s.put(bucketAudits, entry.ID, entry) }

func (s *Store) put(bucket []byte, key string, value any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucket).Put([]byte(key), data) })
}

func (s *Store) GetReading(id string) (model.SensorReading, error) {
	var v model.SensorReading
	err := s.get(bucketReadings, id, &v)
	return v, err
}
func (s *Store) GetBatch(id string) (model.SensorBatch, error) {
	var v model.SensorBatch
	err := s.get(bucketBatches, id, &v)
	return v, err
}
func (s *Store) get(bucket []byte, key string, target any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucket).Get([]byte(key))
		if value == nil {
			return os.ErrNotExist
		}
		return json.Unmarshal(value, target)
	})
}

func (s *Store) ListPending(status model.ReadingStatus) ([]model.PendingItem, error) {
	return s.listPending(status)
}

func (s *Store) listPending(status model.ReadingStatus) ([]model.PendingItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	items := make([]model.PendingItem, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketPending).ForEach(func(_, value []byte) error {
			var item model.PendingItem
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			if status == "" || item.Status == status {
				items = append(items, item)
			}
			return nil
		})
	})
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, err
}

func (s *Store) ListReadings(fieldID string) ([]model.SensorReading, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	items := make([]model.SensorReading, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketReadings).ForEach(func(_, value []byte) error {
			var item model.SensorReading
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			if fieldID == "" || item.FieldID == fieldID {
				items = append(items, item)
			}
			return nil
		})
	})
	sort.Slice(items, func(i, j int) bool { return items[i].CapturedAt.Before(items[j].CapturedAt) })
	return items, err
}

func (s *Store) ListBatches() ([]model.SensorBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	items := make([]model.SensorBatch, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketBatches).ForEach(func(_, value []byte) error {
			var item model.SensorBatch
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		})
	})
	sort.Slice(items, func(i, j int) bool { return items[i].ReceivedAt.Before(items[j].ReceivedAt) })
	return items, err
}

func (s *Store) Count(bucket []byte) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return 0, errors.New("store is closed")
	}
	count := 0
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).ForEach(func(_, _ []byte) error { count++; return nil })
	})
	return count, err
}
