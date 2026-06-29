package extractor

import "github.com/JDucr17/streamline/services/pipeline/internal/hbos"

// FeatureType is a broad category label for a feature, for documentation and any
// type-level handling at fit/score time. Per-feature computation is dispatched by
// Name
type FeatureType string

const (
	Count         FeatureType = "count"
	Numeric       FeatureType = "numeric"
	Ratio         FeatureType = "ratio"
	DistinctCount FeatureType = "distinct_count"
	Entropy       FeatureType = "entropy"
	TimeGap       FeatureType = "time_gap"
)

// Field names the RawEvent field a feature reads.
type Field string

const (
	FieldURI        Field = "uri"
	FieldHTTPMethod Field = "http_method"
)

// FeatureDef declares one registry feature, in vector order. Fields other than
// Name/Type/Transform are arguments the feature's compute function reads
// (which event field, which predicate, which denominator/filter).
type FeatureDef struct {
	Name               string
	Type               FeatureType
	Transform          hbos.Transform
	Field              Field
	NumeratorPredicate string
	Denominator        string
	EventFilter        string
}