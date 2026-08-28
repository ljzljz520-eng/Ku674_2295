package adapter

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"farm-sensor-platform/internal/model"
)

type Catalog struct {
	mu       sync.RWMutex
	adapters map[model.Vendor]Adapter
}

func NewCatalog(adapters ...Adapter) *Catalog {
	catalog := &Catalog{adapters: make(map[model.Vendor]Adapter)}
	for _, adapter := range adapters {
		catalog.Register(adapter)
	}
	return catalog
}

func (c *Catalog) Register(source Adapter) error {
	if source == nil {
		return fmt.Errorf("adapter cannot be nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.adapters[source.Vendor()]; exists {
		return fmt.Errorf("vendor %s already registered", source.Vendor())
	}
	c.adapters[source.Vendor()] = source
	return nil
}

func (c *Catalog) Get(vendor model.Vendor) (Adapter, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	source, ok := c.adapters[vendor]
	return source, ok
}

func (c *Catalog) Vendors() []model.Vendor {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]model.Vendor, 0, len(c.adapters))
	for vendor := range c.adapters {
		result = append(result, vendor)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (c *Catalog) FetchAll(ctx context.Context, ids []string) (map[model.Vendor][]RawReading, []error) {
	result := make(map[model.Vendor][]RawReading)
	errorsFound := make([]error, 0)
	for _, vendor := range c.Vendors() {
		source, _ := c.Get(vendor)
		rows, err := source.Fetch(ctx, ids)
		if err != nil {
			errorsFound = append(errorsFound, fmt.Errorf("%s: %w", vendor, err))
			continue
		}
		result[vendor] = rows
	}
	return result, errorsFound
}

func MergeRawReadings(groups map[model.Vendor][]RawReading) []RawReading {
	vendors := make([]model.Vendor, 0, len(groups))
	for vendor := range groups {
		vendors = append(vendors, vendor)
	}
	sort.Slice(vendors, func(i, j int) bool { return vendors[i] < vendors[j] })
	merged := make([]RawReading, 0)
	for _, vendor := range vendors {
		for _, row := range groups[vendor] {
			merged = append(merged, row)
		}
	}
	return merged
}

func FixtureTimestamp() time.Time { return time.Unix(1_700_000_000, 0).UTC() }
