package baselineworker

import (
	"fmt"
	"sort"

	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/hbos"

	"gonum.org/v1/gonum/stat"
)

const (
	// Raw scores from every baseline window are summarized at these many
	// evenly spaced quantiles, the calibration grid normalizeScore searches.
	scoreQuantileSteps = 1001
)

// Fit fits a source baseline from sampled window feature rows.
func Fit(matrix [][]float64, sourceID string, baselineRunID int64) (hbos.Baseline, error) {
	if err := validateMatrix(matrix, len(extractor.HTTPFeatures)); err != nil {
		return hbos.Baseline{}, err
	}

	histograms, err := fitHistograms(matrix)
	if err != nil {
		return hbos.Baseline{}, err
	}

	quantiles, err := calibrateQuantiles(matrix, histograms)
	if err != nil {
		return hbos.Baseline{}, err
	}

	return hbos.Baseline{
		BaselineRunID:  baselineRunID,
		SourceID:       sourceID,
		WindowCount:    len(matrix),
		HistogramBins:  histogramBins,
		SmoothingAlpha: smoothingAlpha,
		Histograms:     histograms,
		ScoreQuantiles: quantiles,
		RegistryHash:   extractor.RegistryHash(),
	}, nil
}

// validateMatrix checks that every extracted window row matches the active
// feature registry width.
func validateMatrix(matrix [][]float64, width int) error {
	if len(matrix) == 0 {
		return fmt.Errorf("empty feature matrix")
	}

	for i, row := range matrix {
		if len(row) != width {
			return fmt.Errorf("row %d width %d, want %d", i, len(row), width)
		}
	}

	return nil
}

// calibrateQuantiles scores every baseline window through the fitted histograms,
// then stores evenly spaced raw-score quantiles for detector normalization.
func calibrateQuantiles(matrix [][]float64, histograms []hbos.FeatureHistogram) ([]float64, error) {
	scorer, err := hbos.CompileScorer(histograms)
	if err != nil {
		return nil, fmt.Errorf("compile scorer: %w", err)
	}

	rawScores := make([]float64, len(matrix))
	for i, row := range matrix {
		raw, err := scorer.RawScore(row)
		if err != nil {
			return nil, fmt.Errorf("calibrate window %d: %w", i, err)
		}
		rawScores[i] = raw
	}

	sort.Float64s(rawScores)

	quantiles := make([]float64, scoreQuantileSteps)
	for i := range quantiles {
		p := float64(i) / float64(scoreQuantileSteps-1)
		quantiles[i] = stat.Quantile(p, stat.LinInterp, rawScores, nil)
	}
	enforceMonotonic(quantiles)

	return quantiles, nil
}

// enforceMonotonic restores the quantile invariant after interpolation. Across
// flat score regions, floating-point interpolation can dip by a ULP even though
// quantiles are non-decreasing by definition.
func enforceMonotonic(values []float64) {
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			values[i] = values[i-1]
		}
	}
}