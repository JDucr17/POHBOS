package queryapi

import (
	"time"

	"github.com/puzpuzpuz/xsync/v4"

	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

type VisitorIndex struct {
	entries *xsync.Map[string, visitorEntry]
	ttl     time.Duration
}

type visitorEntry struct {
	envelope  envelope.DecisionEnvelope
	expiresAt time.Time
}

func NewVisitorIndex(ttl time.Duration) *VisitorIndex {
	return &VisitorIndex{
		entries: xsync.NewMap[string, visitorEntry](),
		ttl:     ttl,
	}
}

func (v *VisitorIndex) Put(key string, env envelope.DecisionEnvelope) {
	v.entries.Store(key, visitorEntry{
		envelope:  env,
		expiresAt: time.Now().Add(v.ttl),
	})
}

func (v *VisitorIndex) Get(key string) (envelope.DecisionEnvelope, bool) {
	entry, ok := v.entries.Load(key)
	if !ok {
		return envelope.DecisionEnvelope{}, false
	}

	if time.Now().After(entry.expiresAt) {
		return envelope.DecisionEnvelope{}, false
	}

	return entry.envelope, true
}

func (v *VisitorIndex) EvictExpired() {
	now := time.Now()

	v.entries.Range(func(key string, entry visitorEntry) bool {
		if now.After(entry.expiresAt) {
			v.entries.Delete(key)
		}
		return true
	})
}

func (v *VisitorIndex) Size() int {
	return v.entries.Size()
}