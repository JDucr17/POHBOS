package hbos

import (
	"fmt"
	"sort"
)

// Model is a validated runtime HBOS model compiled from a stored Baseline: a
// Scorer plus the quantiles needed to normalize a raw score.
type Model struct {
	sourceID          string
	runID             int64
	scorer            *Scorer
	scoreQuantiles    []float64
}

// Compile validates a stored Baseline and compiles it into a scoreable Model,
// rejecting any artifact the detector cannot score.
func Compile(b *Baseline) (*Model, error) {
	scorer, err := CompileScorer(b.Histograms)
	if err != nil {
		return nil, err
	}
	if err := validateScoreQuantiles(b.ScoreQuantiles); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnscoreable, err)
	}

	return &Model{
		sourceID:          b.SourceID,
		runID:             b.BaselineRunID,
		scorer:            scorer,
		scoreQuantiles:    b.ScoreQuantiles,
	}, nil
}

func (m *Model) SourceID() string { return m.sourceID }
func (m *Model) RunID() int64     { return m.runID }

// Score computes the raw and normalized HBOS scores for one feature row.
func (m *Model) Score(row []float64) (raw, normalized float64, err error) {
	raw, err = m.scorer.RawScore(row)
	if err != nil {
		return 0, 0, err
	}

	normalized = normalizeScore(raw, m.scoreQuantiles)
	if !finite(normalized) {
		return 0, 0, fmt.Errorf("%w: non-finite normalized score", ErrUnscoreable)
	}
	return raw, normalized, nil
}

// normalizeScore maps a raw score to its baseline percentile.
func normalizeScore(raw float64, scoreQuantiles []float64) float64 {
	beaten := sort.Search(len(scoreQuantiles), func(i int) bool {
		return scoreQuantiles[i] > raw
	})
	return float64(beaten) / float64(len(scoreQuantiles))
}