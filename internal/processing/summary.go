package processing

import (
	"fmt"
	"sort"
	"sync"

	"farm-sensor-platform/internal/model"
)

type Summary struct {
	BatchID  string
	Total    int
	Accepted int
	Pending  int
	Failed   bool
	Message  string
}

type Aggregator struct {
	mu      sync.Mutex
	byBatch map[string]Summary
}

func NewAggregator() *Aggregator { return &Aggregator{byBatch: make(map[string]Summary)} }

func (a *Aggregator) Begin(batchID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.byBatch[batchID] = Summary{BatchID: batchID}
}

func (a *Aggregator) Add(batchID string, status model.ReadingStatus) {
	a.mu.Lock()
	defer a.mu.Unlock()
	summary := a.byBatch[batchID]
	summary.Total++
	if status == model.StatusAccepted {
		summary.Accepted++
	}
	if status == model.StatusPending || status == model.StatusRetried {
		summary.Pending++
	}
	summary.Failed = summary.Pending > 0
	summary.Message = fmt.Sprintf("accepted=%d pending=%d", summary.Accepted, summary.Pending)
	a.byBatch[batchID] = summary
}

func (a *Aggregator) Snapshot(batchID string) Summary {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.byBatch[batchID]
}

func (a *Aggregator) All() []Summary {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]Summary, 0, len(a.byBatch))
	for _, item := range a.byBatch {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].BatchID < result[j].BatchID })
	return result
}
