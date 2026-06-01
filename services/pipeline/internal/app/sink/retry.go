package sink

import (
	"context"
	"log/slog"
	"time"
)

// Fail fast over optimistic recovery: brief downstream transients should
// resolve within this retry window. Longer failures should not be hidden by
// long sleeps, leave offsets uncommitted so records remain replayable and
// consumer lag exposes that this sink cannot safely make progress.

const (
	SinkRetryFirstDelay  = 50 * time.Millisecond
	SinkRetrySecondDelay = 200 * time.Millisecond
)

// Three attempts total: initial, then waits of 50ms and 200ms
var retryDelays = []time.Duration{
	0,
	SinkRetryFirstDelay,
	SinkRetrySecondDelay,
}

// retry runs fn with bounded backoff. Stops on success or permanent error
func retry(ctx context.Context, fn func(context.Context) error) error {
	var err error

	for attempt, delay := range retryDelays {
		if delay > 0 {
			if err := sleep(ctx, delay); err != nil {
				return err
			}
		}

		err = fn(ctx)
		if err == nil || isPermanent(err) {
			return err
		}

		slog.Warn("transient failure, retrying",
			slog.Int("attempt", attempt+1),
			slog.Any("error", err),
		)
	}

	return err
}

// sleep returns early if ctx is cancelled, so shutdown isn't blocked by a retry wait
func sleep(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}