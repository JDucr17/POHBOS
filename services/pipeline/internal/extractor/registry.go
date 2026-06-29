package extractor

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"

	"github.com/JDucr17/streamline/services/pipeline/internal/hbos"
)
// Feature list collected from HTTP traffic.
var HTTPFeatures = []*FeatureDef{
	// Volume
	{Name: "window_request_count", Type: Count, Transform: hbos.TransformLog1p},
	{Name: "window_span_seconds", Type: Numeric, Transform: hbos.TransformLog1p},

	// Timing, cadence navigation is measured so asset burst requests are not contemplated
	{Name: "mean_inter_request_gap_ms", Type: Numeric, Transform: hbos.TransformLog1p, EventFilter: "is_html"},
	{Name: "inter_request_gap_cv", Type: Numeric, Transform: hbos.TransformIdentity, EventFilter: "is_html"},

	// Resource type
	{Name: "distinct_endpoints_hit", Type: DistinctCount, Field: FieldURI, Transform: hbos.TransformLog1p},
	{Name: "html_to_asset_ratio", Type: Ratio, NumeratorPredicate: "is_html", Denominator: "window_request_count", Transform: hbos.TransformIdentity},

	// HTTP semantics
	{Name: "error_4xx_rate", Type: Ratio, NumeratorPredicate: "is_4xx_response", Denominator: "window_request_count", Transform: hbos.TransformIdentity},
	{Name: "method_entropy", Type: Entropy, Field: FieldHTTPMethod, Transform: hbos.TransformIdentity},

	// Headers
	{Name: "has_referrer_ratio", Type: Ratio, NumeratorPredicate: "has_referrer", Denominator: "window_request_count", Transform: hbos.TransformIdentity},
}
// RegistryHash fingerprints the feature set: SHA-256 over each FeatureDef's
// fields written in fixed order, in registry order. The detector rejects any
// baseline whose embedded hash differs, so a model is never scored under a
// feature set it was not fit for.
func RegistryHash() string {
	hasher := sha256.New()
	for _, fdef := range HTTPFeatures {
		writeFeatureHash(hasher, fdef)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func writeFeatureHash(h hash.Hash, f *FeatureDef) {
	writeField(h, f.Name)
	writeField(h, string(f.Type))
	writeField(h, string(f.Transform))
	writeField(h, string(f.Field))
	writeField(h, f.NumeratorPredicate)
	writeField(h, f.Denominator)
	writeField(h, f.EventFilter)
}

// writes len(s) then s, so field boundaries are unambiguous.
func writeField(h hash.Hash, s string) {
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], uint64(len(s)))
	_, _ = h.Write(n[:])
	_, _ = h.Write([]byte(s))
}