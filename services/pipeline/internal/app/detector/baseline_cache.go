package detector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/hbos"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

// Latest succeeded baseline per source.
const selectLatestBaselinesSQL = `
SELECT DISTINCT ON (source_id) source_id, baseline
FROM baseline_runs
WHERE status = 'succeeded'
ORDER BY source_id, id DESC
`

// Latest succeeded baseline for one source. The signal carries a run_id, but
// we load the latest rather than that exact run so a stale or out-of-order
// signal can never pull the cache back to an older model.
const selectLatestBaselineForSourceSQL = `
SELECT baseline
FROM baseline_runs
WHERE source_id = $1 AND status = 'succeeded'
ORDER BY id DESC
LIMIT 1
`

type baselinesBySource = map[string]*hbos.Model

// baselineSnapshot is the immutable map published through the atomic pointer.
type baselineSnapshot struct {
	bySource baselinesBySource
}

// BaselineCache serves the current per-source compiled baseline to the scoring
// path. Postgres is the source of truth, this cache mirrors it in memory,
// refreshed at startup and on each baseline signal.
//
// The snapshot is published through an atomic pointer and never mutated in
// place: readers load it lock-free and never stall, writers build a new
// snapshot and swap the pointer.
type BaselineCache struct {
	db           *postgres.DB
	registryHash string
	servingSpec  hbos.WindowSpec
	current      atomic.Pointer[baselineSnapshot]
}

func NewBaselineCache(db *postgres.DB, servingSpec hbos.WindowSpec) *BaselineCache {
	registryHash := extractor.RegistryHash()

	cache := &BaselineCache{db: db, registryHash: registryHash, servingSpec: servingSpec}
	cache.publish(make(baselinesBySource))
	return cache
}

// Get returns the cached baseline for a source, or nil if none is available
// The returned baseline is read-only.
func (c *BaselineCache) Get(sourceID string) *hbos.Model {
	return c.current.Load().bySource[sourceID]
}

// LoadAll replaces the cache with the latest succeeded baseline for every
// source, so a freshly started detector holds every current model without
// waiting for a signal.
func (c *BaselineCache) LoadAll(ctx context.Context) error {
	rows, err := c.db.Pool.Query(ctx, selectLatestBaselinesSQL)
	if err != nil {
		return postgres.Classify(err)
	}
	defer rows.Close()

	loaded := make(baselinesBySource)
	for rows.Next() {
		var sourceID string
		var stored []byte
		if err := rows.Scan(&sourceID, &stored); err != nil {
			return postgres.Classify(err)
		}

		model, err := c.verify(sourceID, stored)
		if err != nil {
			slog.Error("skipping baseline at startup",
				slog.String("source_id", sourceID),
				slog.Any("error", err),
			)
			continue
		}
		loaded[sourceID] = model
	}
	if err := rows.Err(); err != nil {
		return postgres.Classify(err)
	}

	c.publish(loaded)
	slog.Info("baselines loaded at startup", slog.Int("count", len(loaded)))
	return nil
}

// Refresh swaps one source's latest succeeded baseline into the cache. On any
// failure the published snapshot is left untouched, so a bad refit never
// blanks a live model, the detector keeps scoring against the last valid baseline.
func (c *BaselineCache) Refresh(ctx context.Context, sourceID string) error {
	var stored []byte
	if err := c.db.Pool.QueryRow(ctx, selectLatestBaselineForSourceSQL, sourceID).Scan(&stored); err != nil {
		return postgres.Classify(err)
	}

	model, err := c.verify(sourceID, stored)
	if err != nil {
		return err
	}

	c.swapSnapshot(sourceID, model)
	slog.Info("baseline refreshed",
		slog.String("source_id", sourceID),
		slog.Int64("run_id", model.RunID()),
	)
	return nil
}

func (c *BaselineCache) publish(bySource baselinesBySource) {
	c.current.Store(&baselineSnapshot{bySource: bySource})
}

// swapSnapshot publishes a new snapshot that differs by one source. The current
// snapshot is copied, never mutated, so readers continue their tasks without being stalled
func (c *BaselineCache) swapSnapshot(sourceID string, model *hbos.Model) {
	current := c.current.Load().bySource

	next := make(baselinesBySource, len(current)+1)
	for source, existing := range current {
		next[source] = existing
	}
	next[sourceID] = model

	c.publish(next)
}

// verify decodes a stored baseline, rejects it unless this detector can score
// with it (the source matches the queried row, the run id is real, the registry
// hash matches), and compiles it into a scoring model. The hash match proves
// the feature set including transforms agrees with this detector's registry.
func (c *BaselineCache) verify(sourceID string, stored []byte) (*hbos.Model, error) {
	var b hbos.Baseline
	if err := json.Unmarshal(stored, &b); err != nil {
		return nil, fmt.Errorf("unmarshal baseline: %w", err)
	}

	if b.SourceID != sourceID {
		return nil, fmt.Errorf("source mismatch: row %q, baseline %q", sourceID, b.SourceID)
	}
	if b.BaselineRunID <= 0 {
		return nil, fmt.Errorf("invalid baseline_run_id %d", b.BaselineRunID)
	}
	if b.RegistryHash != c.registryHash {
		return nil, fmt.Errorf(
			"registry hash mismatch: baseline %q, detector %q",
			b.RegistryHash, c.registryHash,
		)
	}

	// Window spec is the training/serving contract: the histograms learned a
	// population defined by these three numbers, so any difference is a different
	// distribution.
	if b.WindowSpec.LengthSeconds != c.servingSpec.LengthSeconds ||
		b.WindowSpec.CadenceSeconds != c.servingSpec.CadenceSeconds ||
		b.WindowSpec.MinEvents != c.servingSpec.MinEvents {
		return nil, fmt.Errorf(
			"window spec mismatch: baseline length=%d cadence=%d min_events=%d, "+
				"detector length=%d cadence=%d min_events=%d",
			b.WindowSpec.LengthSeconds, b.WindowSpec.CadenceSeconds, b.WindowSpec.MinEvents,
			c.servingSpec.LengthSeconds, c.servingSpec.CadenceSeconds, c.servingSpec.MinEvents)
	}

	return hbos.Compile(&b)
}