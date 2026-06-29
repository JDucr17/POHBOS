package extractor

import (
	"math"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
)

// featureDefs indexes the shared registry by name so the filter, predicate, and
// denominator a feature uses are read from the same definition the detector and
// registry hash are built from, rather than duplicated here.
var featureDefs = func() map[string]*FeatureDef {
	defs := make(map[string]*FeatureDef, len(HTTPFeatures))
	for _, fdef := range HTTPFeatures {
		defs[fdef.Name] = fdef
	}
	return defs
}()

// Compute returns one raw feature vector for a window, in registry order.
//
// Raw values only: registry transforms are applied later, at fit and score
// time. Values the window cannot supply
// are NaN so the fitter treats them as undefined rather than as zero. The
// timing features measure document/HTML events only, so they capture
// navigation cadence instead of asset bursts.
func Compute(events []domain.RawEvent) []float64 {
	return []float64{
		float64(len(events)),
		windowSpan(events),
		meanGap(filtered(events, "mean_inter_request_gap_ms")),
		gapCV(filtered(events, "inter_request_gap_cv")),
		distinctEndpoints(events),
		ratio(events, "html_to_asset_ratio"),
		ratio(events, "error_4xx_rate"),
		methodEntropy(events),
		ratio(events, "has_referrer_ratio"),
	}
}

// windowSpan is the wall-clock span of the window in seconds.
func windowSpan(events []domain.RawEvent) float64 {
	return events[len(events)-1].EventTime.Sub(events[0].EventTime).Seconds()
}

// meanGap is the mean inter-event gap in ms, NaN when the window has no gap.
func meanGap(events []domain.RawEvent) float64 {
	gaps := gapsMs(events)
	if len(gaps) == 0 {
		return math.NaN()
	}
	return sum(gaps) / float64(len(gaps))
}

// gapCV is the coefficient of variation of inter-event gaps (population
// stddev / mean). NaN unless there are at least two gaps and a nonzero mean:
// variance needs two samples and a zero mean makes the ratio meaningless.
func gapCV(events []domain.RawEvent) float64 {
	gaps := gapsMs(events)
	if len(gaps) < 2 {
		return math.NaN()
	}

	mean := sum(gaps) / float64(len(gaps))
	if mean == 0.0 {
		return math.NaN()
	}

	var variance float64
	for _, gap := range gaps {
		variance += (gap - mean) * (gap - mean)
	}
	variance /= float64(len(gaps))
	return math.Sqrt(variance) / mean
}

// distinctEndpoints counts distinct query-stripped paths. Endpoint identity
// comes from domain.EndpointPath so /p/1 and /p/2 differ but /p/1?ref=a and
// /p/1?ref=b do not. The detector's scoring side must strip with the same
// domain.EndpointPath when wired, or this feature diverges between fit and score.
func distinctEndpoints(events []domain.RawEvent) float64 {
	paths := make(map[string]struct{}, len(events))
	for _, event := range events {
		paths[domain.EndpointPath(event.URI)] = struct{}{}
	}
	return float64(len(paths))
}

// methodEntropy is the log2 Shannon entropy over the window's HTTP methods.
func methodEntropy(events []domain.RawEvent) float64 {
	methods := make([]string, len(events))
	for i, event := range events {
		methods[i] = event.HTTPMethod
	}
	return entropy(methods)
}

// filtered selects the events a feature measures, per its registry event_filter.
func filtered(events []domain.RawEvent, featureName string) []domain.RawEvent {
	fdef := featureDefs[featureName]
	if fdef.EventFilter == "" {
		return events
	}

	predicate := Predicates[fdef.EventFilter]
	matched := make([]domain.RawEvent, 0, len(events))
	for _, event := range events {
		if predicate(event) {
			matched = append(matched, event)
		}
	}
	return matched
}

// ratio is a registry-declared ratio: predicate count over window_request_count.
// NaN when the denominator is not positive, since the ratio is then undefined
// rather than zero.
func ratio(events []domain.RawEvent, featureName string) float64 {
	fdef := featureDefs[featureName]
	predicate := Predicates[fdef.NumeratorPredicate]

	var numerator int
	for _, event := range events {
		if predicate(event) {
			numerator++
		}
	}

	denominator := len(events)
	if denominator <= 0 {
		return math.NaN()
	}
	return float64(numerator) / float64(denominator)
}

// gapsMs returns inter-event gaps in fractional milliseconds. Seconds()*1000
// keeps sub-millisecond resolution that Milliseconds() would truncate.
func gapsMs(events []domain.RawEvent) []float64 {
	if len(events) < 2 {
		return nil
	}

	gaps := make([]float64, 0, len(events)-1)
	for i := 1; i < len(events); i++ {
		gaps = append(gaps, events[i].EventTime.Sub(events[i-1].EventTime).Seconds()*1000.0)
	}
	return gaps
}

// entropy is the log2 Shannon entropy over a categorical sample, 0.0 for empty.
func entropy(values []string) float64 {
	total := len(values)
	if total == 0 {
		return 0.0
	}

	counts := make(map[string]int)
	for _, value := range values {
		counts[value]++
	}

	var e float64
	for _, count := range counts {
		probability := float64(count) / float64(total)
		e -= probability * math.Log2(probability)
	}
	return e
}

func sum(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total
}