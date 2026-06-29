package baselineworker

import (
	"context"
	"log/slog"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)

// Worker runs one baseline fit cycle: extract per-source windows, then fit,
// store, and signal each source independently.
type Worker struct {
	db          *postgres.DB
	producer    *broker.Producer
	windowDays  int
	spec        extractor.WindowSpec
	minWindows  int
	minVisitors int
}

func NewWorker(db *postgres.DB, producer *broker.Producer, windowDays int, spec extractor.WindowSpec, minWindows, minVisitors int) *Worker {
	return &Worker{
		db:          db,
		producer:    producer,
		windowDays:  windowDays,
		spec:        spec,
		minWindows:  minWindows,
		minVisitors: minVisitors,
	}
}

// Run executes one cycle. A source that fails to fit or store is recorded and
// skipped so it never aborts the rest of the run.
func (w *Worker) Run(ctx context.Context) error {
	extracts, err := Extract(ctx, w.db.Pool, w.windowDays, w.spec)
	if err != nil {
		return err
	}
	if len(extracts) == 0 {
		slog.Info("no sources to fit")
		return nil
	}

	for sourceID, source := range extracts {
		if err := ctx.Err(); err != nil {
			return err
		}
		w.processSource(ctx, sourceID, source)
	}
	return nil
}

func (w *Worker) processSource(ctx context.Context, sourceID string, source *SourceExtract) {
	windowCount := len(source.Matrix)
	job := fitJob{sourceID: sourceID, source: source, spec: w.spec, windowDays: w.windowDays}

	// Both gates apply: enough sampled windows for a stable fit, and enough
	// distinct visitors that a few heavy ones cannot define the site baseline.
	if windowCount < w.minWindows || source.DistinctVisitors < w.minVisitors {
		w.recordTerminal(ctx, job, "insufficient_history")
		slog.Info("source insufficient_history",
			slog.String("source_id", sourceID),
			slog.Int("windows", windowCount),
			slog.Int("visitors", source.DistinctVisitors),
			slog.Int("min_windows", w.minWindows),
			slog.Int("min_visitors", w.minVisitors),
		)
		return
	}

	runID, err := fitAndStore(ctx, w.db, job)
	if err != nil {
		// The transaction rolled back, record the failure as its own row.
		slog.Error("fit or store failed", slog.String("source_id", sourceID), slog.Any("error", err))
		w.recordTerminal(ctx, job, "failed")
		return
	}

	if err := publishSignal(ctx, w.producer, sourceID, runID); err != nil {
		slog.Error("baseline signal publish failed",
			slog.String("source_id", sourceID),
			slog.Int64("run_id", runID),
			slog.Any("error", err),
		)
	}

	slog.Info("source succeeded",
		slog.String("source_id", sourceID),
		slog.Int64("run_id", runID),
		slog.Int("windows", windowCount),
		slog.Int("visitors", source.DistinctVisitors),
		slog.Int("events", source.EventCount),
		slog.Int("window_days", w.windowDays),
	)
}

func (w *Worker) recordTerminal(ctx context.Context, job fitJob, status string) {
	if err := insertTerminal(ctx, w.db, job, status); err != nil {
		slog.Error("terminal row insert failed",
			slog.String("source_id", job.sourceID),
			slog.String("status", status),
			slog.Any("error", err),
		)
	}
}
