package hbos

import (
	"errors"
	"fmt"
	"math"
)

var ErrUnscoreable = errors.New("hbos: unscoreable")

// compiledHistogram pairs a fitted histogram with its resolved transform so the
// two can never be misaligned at score time.
type compiledHistogram struct {
	hist      FeatureHistogram
	transform TransformFunc
}

// Scorer is validated, compiled histograms capable of raw HBOS scoring. The
// fitter builds one to score baseline windows before score quantiles exist.
type Scorer struct {
	histograms []compiledHistogram
}

func CompileScorer(histograms []FeatureHistogram) (*Scorer, error) {
	if len(histograms) == 0 {
		return nil, fmt.Errorf("%w: no histograms", ErrUnscoreable)
	}

	compiled := make([]compiledHistogram, len(histograms))
	for i, hist := range histograms {
		if err := validateHistogram(hist); err != nil {
			return nil, fmt.Errorf("%w: histogram %q: %v", ErrUnscoreable, hist.Name, err)
		}

		transform, err := CompileTransform(hist.Transform)
		if err != nil {
			return nil, fmt.Errorf("histogram %q: %w", hist.Name, err)
		}

		compiled[i] = compiledHistogram{hist: hist, transform: transform}
	}

	return &Scorer{histograms: compiled}, nil
}

// RawScore sums feature contributions in registry order. row must align with the
// histograms the Scorer was compiled from.
func (s *Scorer) RawScore(row []float64) (float64, error) {
	if len(row) != len(s.histograms) {
		return 0, fmt.Errorf("%w: feature width %d, histograms %d",
			ErrUnscoreable, len(row), len(s.histograms))
	}

	var sum float64
	for i, ch := range s.histograms {
		sum += contribution(row[i], ch)
	}

	if !finite(sum) {
		return 0, fmt.Errorf("%w: non-finite raw score", ErrUnscoreable)
	}
	return sum, nil
}

// contribution is one feature's surprise, -log(probability of value): the
// undefined branch for NaN, otherwise the bin mass scaled by the defined rate.
func contribution(value float64, ch compiledHistogram) float64 {
	if math.IsNaN(value) {
		return -math.Log(ch.hist.PUndefined)
	}

	definedFactor := 1.0 - ch.hist.PUndefined
	return -math.Log(definedFactor * binMass(ch.transform(value), ch.hist))
}

// binMass returns the probability mass a transformed value falls into: its bin
// when inside the fitted range, otherwise the out-of-range mass.
func binMass(transformed float64, hist FeatureHistogram) float64 {
	if hist.BinWidth == 0.0 {
		if len(hist.BinMasses) == 1 && transformed == hist.BinMin {
			return hist.BinMasses[0]
		}
		return hist.OutOfRangeMass
	}

	upperEdge := hist.BinMin + hist.BinWidth*float64(len(hist.BinMasses))
	if transformed < hist.BinMin || transformed > upperEdge {
		return hist.OutOfRangeMass
	}

	binIndex := int((transformed - hist.BinMin) / hist.BinWidth)
	if binIndex > len(hist.BinMasses)-1 {
		binIndex = len(hist.BinMasses) - 1
	}
	return hist.BinMasses[binIndex]
}