package extractor

import (
	pb "github.com/JDucr17/streamline/services/pipeline/internal/extractor/proto"
)

// Operational constants for feature extraction
const (
	// rolling visitor session window in seconds
	WindowSeconds = 300

	// minimum number of session feature vectors a
	// source must accumulate before its HBOS histograms are fitted, below this
	// the source is flagged as insufficient_history.
	MinBaselineSessions = 3000
)

// Aliases for less verbose registry literals.
const (
	// Feature types
	Count       = pb.FeatureType_FEATURE_TYPE_COUNT
	Numeric     = pb.FeatureType_FEATURE_TYPE_NUMERIC
	Ratio       = pb.FeatureType_FEATURE_TYPE_RATIO
	HLLDistinct = pb.FeatureType_FEATURE_TYPE_HLL_DISTINCT
	Entropy     = pb.FeatureType_FEATURE_TYPE_ENTROPY
	TimeGap     = pb.FeatureType_FEATURE_TYPE_TIME_GAP

	//Transformations
	Log1p = pb.Transform_TRANSFORM_LOG1P

	//Aggregations
	Mean = pb.Aggregation_AGGREGATION_MEAN
	CV   = pb.Aggregation_AGGREGATION_CV
	Last = pb.Aggregation_AGGREGATION_LAST
	Max  = pb.Aggregation_AGGREGATION_MAX

	//Fields
	FieldURI        = pb.Field_FIELD_URI
	FieldHTTPMethod = pb.Field_FIELD_HTTP_METHOD
)

// Feature list collected from HTTP traffic.
var HTTPFeatures = []*pb.FeatureDef{
	// Volume
	{Name: "session_request_count", Type: Count, Transform: Log1p},
	{Name: "session_duration_seconds", Type: Numeric, Aggregation: Last},

	// Timing
	{Name: "mean_inter_request_gap_ms", Type: Numeric, Aggregation: Mean},
	{Name: "inter_request_gap_cv", Type: Numeric, Aggregation: CV},

	// Resource type
	{Name: "distinct_endpoints_hit", Type: HLLDistinct, Field: FieldURI, Transform: Log1p},
	{Name: "html_to_asset_ratio", Type: Ratio, NumeratorPredicate: "is_html", Denominator: "session_request_count"},

	// HTTP semantics
	{Name: "error_4xx_rate", Type: Ratio, NumeratorPredicate: "is_4xx_response", Denominator: "session_request_count"},
	{Name: "method_entropy", Type: Entropy, Field: FieldHTTPMethod},

	// Headers
	{Name: "has_referrer_ratio", Type: Ratio, NumeratorPredicate: "has_referrer", Denominator: "session_request_count", DenominatorOffset: -1},
}
