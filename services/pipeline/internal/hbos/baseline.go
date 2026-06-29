// Package hbos defines the HBOS model artifact and pure scoring math.
// The baseline worker produces Baseline values from extracted feature windows
// the detector loads, validates, and scores against them.

package hbos

import "time"

// Baseline is one source's fitted HBOS model, serialized as JSON into
// baseline_runs.baseline.
type Baseline struct {
	BaselineRunID int64     `json:"baseline_run_id"`
	SourceID      string    `json:"source_id"`
	FitAt         time.Time `json:"fit_at"`

	WindowCount      int                `json:"window_count"`
	DistinctVisitors int                `json:"distinct_visitors"`
	HistogramBins    int                `json:"histogram_bins"`
	SmoothingAlpha   float64            `json:"smoothing_alpha"`
	WindowSpec       WindowSpec         `json:"window_spec"`
	Histograms       []FeatureHistogram `json:"histograms"`
	ScoreQuantiles   []float64          `json:"score_quantiles"`
	RegistryHash     string             `json:"registry_hash"`
}

// WindowSpec records the extraction window used to fit the model. The detector
// rejects baselines whose window spec does not match its scoring configuration.
type WindowSpec struct {
	LengthSeconds  int `json:"length_seconds"`
	CadenceSeconds int `json:"cadence_seconds"`
	MinEvents      int `json:"min_events"`
	HorizonDays    int `json:"horizon_days"`
}

// FeatureHistogram is one feature's fitted HBOS distribution.
type FeatureHistogram struct {
	Name           string    `json:"name"`
	Transform      Transform `json:"transform"`
	PUndefined     float64   `json:"p_undefined"`
	DefinedCount   int       `json:"defined_count"`
	BinMin         float64   `json:"bin_min"`
	BinWidth       float64   `json:"bin_width"`
	BinMasses      []float64 `json:"bin_masses"`
	OutOfRangeMass float64   `json:"out_of_range_mass"`
}