package queryapi

import (
	"fmt"
	"sync"

	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

// RecentRing holds the last N decisions in arrival order.
// Uses a fixed-size circular buffer; capacity must be a power of two.
type RecentRing struct {
	mu       sync.RWMutex
	items    []envelope.DecisionEnvelope
	capacity int
	mask     int
	next     int
	full     bool
}

// NewRecentRing returns a ring with the given capacity. Capacity must be a
// positive power of two so wrap-around can use bitwise AND.
func NewRecentRing(capacity int) (*RecentRing, error) {
	if capacity <= 0 || capacity&(capacity-1) != 0 {
		return nil, fmt.Errorf("recent ring capacity must be a positive power of two, got %d", capacity)
	}

	return &RecentRing{
		items:    make([]envelope.DecisionEnvelope, capacity),
		capacity: capacity,
		mask:     capacity - 1,
	}, nil
}

// AppendBatch writes decisions in order, overwriting the oldest entries
// once the ring is full.
func (r *RecentRing) AppendBatch(decisions []envelope.DecisionEnvelope) {
	if len(decisions) == 0 {
		return
	}

	r.mu.Lock()

	for _, env := range decisions {
		r.items[r.next] = env
		r.next = (r.next + 1) & r.mask
		if !r.full && r.next == 0 {
			r.full = true
		}
	}

	r.mu.Unlock()
}

// Snapshot returns all retained decisions in arrival order, oldest first.
func (r *RecentRing) Snapshot() []envelope.DecisionEnvelope {
	r.mu.RLock()

	size := r.sizeLocked()
	out := make([]envelope.DecisionEnvelope, 0, size)

	if !r.full {
		out = append(out, r.items[:r.next]...)
	} else {
		out = append(out, r.items[r.next:]...)
		out = append(out, r.items[:r.next]...)
	}

	r.mu.RUnlock()

	return out
}

// Latest returns up to limit retained decisions, newest first.
func (r *RecentRing) Latest(limit int) []envelope.DecisionEnvelope {
	r.mu.RLock()

	size := r.sizeLocked()
	if limit <= 0 || limit > size {
		limit = size
	}

	out := make([]envelope.DecisionEnvelope, 0, limit)
	for i := 0; i < limit; i++ {
		idx := (r.next - 1 - i + r.capacity) & r.mask
		out = append(out, r.items[idx])
	}

	r.mu.RUnlock()

	return out
}

// Size returns the current number of retained decisions.
func (r *RecentRing) Size() int {
	r.mu.RLock()
	size := r.sizeLocked()
	r.mu.RUnlock()

	return size
}

func (r *RecentRing) sizeLocked() int {
	if r.full {
		return r.capacity
	}
	return r.next
}