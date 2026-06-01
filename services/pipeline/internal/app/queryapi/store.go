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