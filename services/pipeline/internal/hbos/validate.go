package hbos

import (
	"fmt"
	"math"
)

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// validateHistogram rejects histogram shapes the scorer cannot use safely.
// The scorer models each feature as a defined/undefined mixture:
// PUndefined scores NaN values, and 1-PUndefined scales defined values.
// Both branches must have positive finite mass because scoring takes -log(probability).
func validateHistogram(hist FeatureHistogram) error {
	if !finite(hist.PUndefined) || hist.PUndefined <= 0 || hist.PUndefined >= 1 {
		return fmt.Errorf("p_undefined out of range: %v", hist.PUndefined)
	}
	if !finite(hist.OutOfRangeMass) || hist.OutOfRangeMass <= 0 {
		return fmt.Errorf("out_of_range_mass not positive: %v", hist.OutOfRangeMass)
	}
	if !finite(hist.BinMin) {
		return fmt.Errorf("bin_min not finite: %v", hist.BinMin)
	}
	if !finite(hist.BinWidth) || hist.BinWidth < 0 {
		return fmt.Errorf("bin_width invalid: %v", hist.BinWidth)
	}

	switch {
	case hist.BinWidth == 0 && len(hist.BinMasses) > 1:
		return fmt.Errorf("zero-width histogram has %d bin masses", len(hist.BinMasses))
	case hist.BinWidth > 0 && len(hist.BinMasses) == 0:
		return fmt.Errorf("binned histogram has no bin masses")
	}

	for _, mass := range hist.BinMasses {
		if !finite(mass) || mass <= 0 {
			return fmt.Errorf("bin mass not positive: %v", mass)
		}
	}

	return nil
}

// validateScoreQuantiles checks the quantile list normalizeScore searches: it
// must be finite and ascending.
func validateScoreQuantiles(quantiles []float64) error {
	if len(quantiles) == 0 {
		return fmt.Errorf("baseline has no score quantiles")
	}

	for i, q := range quantiles {
		if !finite(q) {
			return fmt.Errorf("score quantile %d not finite: %v", i, q)
		}
		if i > 0 && q < quantiles[i-1] {
			return fmt.Errorf("score quantiles not sorted at %d", i)
		}
	}

	return nil
}