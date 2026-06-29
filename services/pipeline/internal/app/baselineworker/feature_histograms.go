// This file fits an HBOS histogram for every feature in the schema from the
// extracted window-feature matrix. fit.go orchestrates the full baseline, this
// file owns the mechanics of turning a feature's column of values into its
// fitted distribution.

package baselineworker

import (
	"fmt"
	"math"
	"sort"

	"github.com/JDucr17/streamline/services/pipeline/internal/extractor"
	"github.com/JDucr17/streamline/services/pipeline/internal/hbos"
)

const (
	// Laplace pseudo-count added to every histogram bin so no bin has zero
	// density, also the density assigned to out-of-range values at score time.
	smoothingAlpha = 0.1
	histogramBins  = 50
)


// fitHistograms fits one HBOS histogram per registry feature, in registry order.
func fitHistograms(matrix [][]float64) ([]hbos.FeatureHistogram, error) {
	histograms := make([]hbos.FeatureHistogram, len(extractor.HTTPFeatures))

	for i, fdef := range extractor.HTTPFeatures {
		apply, err := hbos.CompileTransform(fdef.Transform)
		if err != nil {
			return nil, fmt.Errorf("feature %s: %w", fdef.Name, err)
		}
		histograms[i] = fitHistogram(columnOf(matrix, i), fdef, apply)
	}

	return histograms, nil
}

// fitHistogram fits one feature column into the HBOS distribution stored in the
// baseline artifact.
func fitHistogram(column []float64, fdef *extractor.FeatureDef, apply hbos.TransformFunc) hbos.FeatureHistogram {
	defined := definedValues(column)

	hist := hbos.FeatureHistogram{
		Name:         fdef.Name,
		Transform:    fdef.Transform,
		PUndefined:   undefinedProbability(len(column), len(defined)),
		DefinedCount: len(defined),
	}

	if len(defined) == 0 {
		// No baseline window produced a numeric value: bin_min/bin_width stay
		// zero, no masses, all mass is out of range.
		hist.OutOfRangeMass = 1.0
		return hist
	}

	transformed := sortedTransform(defined, apply)
	if transformed[len(transformed)-1] == transformed[0] {
		return constantShape(hist, transformed[0])
	}

	return binnedShape(hist, transformed)
}

func definedValues(column []float64) []float64 {
	defined := make([]float64, 0, len(column))
	for _, v := range column {
		if !math.IsNaN(v) {
			defined = append(defined, v)
		}
	}
	return defined
}

// undefinedProbability Laplace-smooths the defined/undefined split so neither
// branch is ever impossible and both -log score paths stay finite.
func undefinedProbability(total, defined int) float64 {
	return (float64(total-defined) + smoothingAlpha) / (float64(total) + 2*smoothingAlpha)
}

// sortedTransform applies the feature transform and sorts ascending, the order
// binning and min/max selection depend on.
func sortedTransform(defined []float64, apply hbos.TransformFunc) []float64 {
	transformed := make([]float64, len(defined))
	for i, v := range defined {
		transformed[i] = apply(v)
	}
	sort.Float64s(transformed)
	return transformed
}

// constantShape is the one-bin model for a feature that never varies: its lone
// value is normal, any other finite value is out of range.
func constantShape(hist hbos.FeatureHistogram, value float64) hbos.FeatureHistogram {
	hist.BinMin = value
	hist.BinMasses = []float64{1.0}
	hist.OutOfRangeMass = outOfRangeMass(hist.DefinedCount)
	return hist
}

// binnedShape is the normal model: fixed-width bins over the fitted range, every
// bin smoothed so none is impossible.
func binnedShape(hist hbos.FeatureHistogram, transformed []float64) hbos.FeatureHistogram {
	minValue := transformed[0]
	maxValue := transformed[len(transformed)-1]
	binWidth := (maxValue - minValue) / float64(histogramBins)

	hist.BinMin = minValue
	hist.BinWidth = binWidth
	hist.BinMasses = smoothedMasses(binCounts(transformed, minValue, binWidth))
	hist.OutOfRangeMass = outOfRangeMass(hist.DefinedCount)
	return hist
}

// binCounts buckets transformed values using the same index-and-clamp rule the
// scorer uses at runtime. Values equal to the fitted maximum clamp into the last
// bin, so every defined baseline value contributes to the fitted distribution.
func binCounts(values []float64, binMin, binWidth float64) []float64 {
	counts := make([]float64, histogramBins)

	for _, v := range values {
		binIndex := int((v - binMin) / binWidth)
		if binIndex < 0 {
			binIndex = 0
		}
		if binIndex > histogramBins-1 {
			binIndex = histogramBins - 1
		}
		counts[binIndex]++
	}

	return counts
}

// smoothedMasses turns raw bin counts into a probability distribution, adding
// the Laplace pseudo-count so empty bins stay rare but not impossible.
func smoothedMasses(counts []float64) []float64 {
	smoothed := make([]float64, len(counts))

	var total float64
	for i, c := range counts {
		smoothed[i] = c + smoothingAlpha
		total += smoothed[i]
	}

	masses := make([]float64, len(smoothed))
	for i, s := range smoothed {
		masses[i] = s / total
	}

	return masses
}

// outOfRangeMass is the empty-bin mass given to finite values beyond the fitted
// range: finite, but maximally surprising for this feature.
func outOfRangeMass(definedCount int) float64 {
	return smoothingAlpha / (float64(definedCount) + smoothingAlpha*histogramBins)
}

// columnOf extracts feature column j from the window-feature matrix.
func columnOf(matrix [][]float64, j int) []float64 {
	column := make([]float64, len(matrix))
	for i := range matrix {
		column[i] = matrix[i][j]
	}
	return column
}