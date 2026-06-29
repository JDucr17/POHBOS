package baselineworker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/hbos"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

const idSequence = "public.baseline_runs_id_seq"

const reserveRunIDSQL = `SELECT nextval($1::regclass)`

// clock_timestamp() is wall-clock at call time (now() is fixed at tx start).
// Captured after Fit so fit_at means when the artifact was finalized, and
// DB-sourced so it matches the fit_at inside the serialized model.
const databaseTimeSQL = `SELECT clock_timestamp()`

// Run id is drawn first so it can be embedded in the model before the row is
// written, then carried back in with OVERRIDING SYSTEM VALUE. fit_at is explicit
// so it matches the value inside the serialized model.
const insertSucceededSQL = `
	INSERT INTO baseline_runs
		(id, source_id, event_count, window_count, distinct_visitors, registry_hash, features_fit, status, baseline, fit_at)
	OVERRIDING SYSTEM VALUE
	VALUES ($1, $2, $3, $4, $5, $6, $7, 'succeeded', $8, $9)
`

// Terminal rows carry no model, so registry_hash comes from the worker's
// compiled-in registry. id is generated, features_fit empty, fit_at defaults.
const insertTerminalSQL = `
	INSERT INTO baseline_runs
		(source_id, event_count, window_count, distinct_visitors, registry_hash, features_fit, status)
	VALUES ($1, $2, $3, $4, $5, '{}', $6)
`

// fitJob carries one source's extract plus the window configuration it was
// sampled under, so fit and store can record the full provenance.
type fitJob struct {
	sourceID   string
	source     *SourceExtract
	spec       extractor.WindowSpec
	windowDays int
}

// fitAndStore fits one baseline and stores it with its audit row in a single
// committed transaction.
func fitAndStore(ctx context.Context, db *postgres.DB, job fitJob) (int64, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return 0, postgres.Classify(err)
	}
	defer tx.Rollback(ctx)

	runID, err := reserveRunID(ctx, tx)
	if err != nil {
		return 0, err
	}

	model, err := Fit(job.source.Matrix, job.sourceID, runID)
	if err != nil {
		return 0, fmt.Errorf("fit baseline: %w", err)
	}

	fitAt, err := databaseTime(ctx, tx)
	if err != nil {
		return 0, err
	}
	model.FitAt = fitAt
	model.DistinctVisitors = job.source.DistinctVisitors
	model.WindowSpec = artifactWindowSpec(job.spec, job.windowDays)

	encoded, err := json.Marshal(model)
	if err != nil {
		return 0, fmt.Errorf("marshal baseline: %w", err)
	}

	if _, err := tx.Exec(ctx, insertSucceededSQL,
		runID, job.sourceID, job.source.EventCount, model.WindowCount,
		model.DistinctVisitors, model.RegistryHash, featuresFit(model), encoded, model.FitAt,
	); err != nil {
		return 0, postgres.Classify(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, postgres.Classify(err)
	}
	return runID, nil
}

func reserveRunID(ctx context.Context, tx pgx.Tx) (int64, error) {
	var runID int64
	if err := tx.QueryRow(ctx, reserveRunIDSQL, idSequence).Scan(&runID); err != nil {
		return 0, postgres.Classify(err)
	}
	return runID, nil
}

func databaseTime(ctx context.Context, tx pgx.Tx) (time.Time, error) {
	var now time.Time
	if err := tx.QueryRow(ctx, databaseTimeSQL).Scan(&now); err != nil {
		return time.Time{}, postgres.Classify(err)
	}
	return now.UTC(), nil
}

// insertTerminal records a source that did not produce a model. The registry
// hash still comes from the compiled-in registry so terminal rows are
// drift-analyzable alongside succeeded ones.
func insertTerminal(ctx context.Context, db *postgres.DB, job fitJob, status string) error {
	if _, err := db.Pool.Exec(ctx, insertTerminalSQL,
		job.sourceID, job.source.EventCount, len(job.source.Matrix),
		job.source.DistinctVisitors, extractor.RegistryHash(), status,
	); err != nil {
		return postgres.Classify(err)
	}
	return nil
}

// artifactWindowSpec translates the extractor's duration-based spec, plus the
// fit horizon, into the integer-seconds form stored in the baseline.
func artifactWindowSpec(spec extractor.WindowSpec, windowDays int) hbos.WindowSpec {
	return hbos.WindowSpec{
		LengthSeconds:  int(spec.Length.Seconds()),
		CadenceSeconds: int(spec.Cadence.Seconds()),
		MinEvents:      spec.MinEvents,
		HorizonDays:    windowDays,
	}
}

func featuresFit(model hbos.Baseline) []string {
	names := make([]string, len(model.Histograms))
	for i, hist := range model.Histograms {
		names[i] = hist.Name
	}
	return names
}