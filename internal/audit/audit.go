package audit

import (
	"fmt"
	"sync"
	"time"

	"farm-sensor-platform/internal/model"
	"farm-sensor-platform/internal/persistence"
)

type Recorder struct {
	store    *persistence.Store
	mu       sync.Mutex
	sequence int
}

func NewRecorder(store *persistence.Store) *Recorder { return &Recorder{store: store} }

func (r *Recorder) Record(action, subject, actor, detail string, now time.Time) (model.AuditEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sequence++
	entry := model.AuditEntry{ID: fmt.Sprintf("audit-%d", r.sequence), Action: action, SubjectID: subject, Actor: actor, Detail: detail, OccurredAt: now.UTC()}
	return entry, r.store.PutAudit(entry)
}
