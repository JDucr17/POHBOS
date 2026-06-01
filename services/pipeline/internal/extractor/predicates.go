package extractor

import (
	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
)

// Predicate tests whether a raw HTTP event matches a condition.
// Used by ratio features to filter which events count toward the numerator.
type Predicate func(event domain.RawEvent) bool

// Registry of named predicates referenced by ratio features.
var Predicates = map[string]Predicate{
	"is_html":         func(e domain.RawEvent) bool { return e.ResourceType == "html" },
	"is_4xx_response": func(e domain.RawEvent) bool { return e.StatusCode >= 400 && e.StatusCode < 500 },
	"has_referrer":    func(e domain.RawEvent) bool { return e.ReferrerPresent },
}
