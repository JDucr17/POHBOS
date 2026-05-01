package constants

import "time"

// Operational constants for the HBOS pipeline.
const (
	// Rolling visitor session window in seconds.
	WindowSeconds = 300

	// Minimum events before a baseline is fitted. Below this cold-start.
	MinHistoryEvents = 1000

	// How long a fitted baseline lives in Redis.
	BaselineTTL = time.Hour

	// Cache TTL for (source, visitor) pair lookup.
	RecencyTTL = 30 * time.Second

	// Laplace smoothing factor.
	HBOSSmoothingAlpha = 0.1

	// Bin count for per-feature histograms.
	HistogramBins = 50
)