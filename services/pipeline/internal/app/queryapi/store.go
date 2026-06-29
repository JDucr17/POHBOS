package queryapi

import (
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

// Store composes the Query API's in-memory decision views:
// latest decision per visitor and recent decisions in arrival order.
type Store struct {
	visitors *VisitorIndex
	recent   *RecentRing
}

func NewStore(ttl time.Duration, ringCapacity int) (*Store, error) {
	recent, err := NewRecentRing(ringCapacity)
	if err != nil {
		return nil, err
	}

	return &Store{
		visitors: NewVisitorIndex(ttl),
		recent:   recent,
	}, nil
}

func visitorKey(sourceID, visitorID string) string {
	return sourceID + ":" + visitorID
}

// RecordBatch updates both in-memory views with decisions in arrival order.
func (s *Store) RecordBatch(decisions []envelope.DecisionEnvelope) {
	if len(decisions) == 0 {
		return
	}

	s.recent.AppendBatch(decisions)

	for _, env := range decisions {
		s.visitors.Put(visitorKey(env.SourceID, env.VisitorID), env)
	}
}

// GetLatest returns the latest non-expired decision for a visitor.
func (s *Store) GetLatest(sourceID, visitorID string) (envelope.DecisionEnvelope, bool) {
	return s.visitors.Get(visitorKey(sourceID, visitorID))
}

// Recent returns retained decisions newest first, up to limit.
func (s *Store) Recent(limit int) []envelope.DecisionEnvelope {
	return s.recent.Latest(limit)
}

// RecentInArrivalOrder returns recent retained decisions in the order they arrived.
// limit<=0 means no backfill.
func (s *Store) RecentInArrivalOrder(limit int) []envelope.DecisionEnvelope {
	if limit <= 0 {
		return nil
	}

	decisions := s.recent.Snapshot()
	if limit >= len(decisions) {
		return decisions
	}

	return decisions[len(decisions)-limit:]
}

// RecentForSourceInArrivalOrder returns recent retained decisions for one source
// in arrival order, so SSE backfill matches the live stream's source filter. A
// blank sourceID matches every source. limit<=0 means no backfill. It scans the
// ring snapshot at call time rather than keeping a per-source index.
func (s *Store) RecentForSourceInArrivalOrder(sourceID string, limit int) []envelope.DecisionEnvelope {
	if limit <= 0 {
		return nil
	}
	if sourceID == "" {
		return s.RecentInArrivalOrder(limit)
	}

	matching := make([]envelope.DecisionEnvelope, 0, limit)
	for _, decision := range s.recent.Snapshot() {
		if decision.SourceID == sourceID {
			matching = append(matching, decision)
		}
	}

	if limit >= len(matching) {
		return matching
	}

	return matching[len(matching)-limit:]
}

// EvictExpired removes expired visitor entries. Ring entries expire by overwrite.
func (s *Store) EvictExpired() {
	s.visitors.EvictExpired()
}

func (s *Store) VisitorCount() int {
	return s.visitors.Size()
}

func (s *Store) RecentSize() int {
	return s.recent.Size()
}
