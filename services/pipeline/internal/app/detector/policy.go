package detector

import (
	"fmt"
	"math"
)

// Band maps an upper-exclusive score range to a risk level and recommended
// action. A nil MaxScore marks the final, open-ended band.
type Band struct {
	RiskLevel string   `json:"risk_level"`
	MaxScore  *float64 `json:"max_score"`
	Action    string   `json:"action"`
}

// Policy is the active score -> risk -> action mapping plus fallbacks for the
// non-scored statuses.
type Policy struct {
	Version         string            `json:"-"`
	Bands           []Band            `json:"bands"`
	FallbackActions map[string]string `json:"fallback_actions"`
}

var requiredFallbackStatuses = [...]string{
	"insufficient_history",
	"no_baseline",
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func isValidRiskLevel(risk string) bool {
	switch risk {
	case "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}

func isValidAction(action string) bool {
	switch action {
	case "allow", "log", "challenge", "block":
		return true
	default:
		return false
	}
}

// ValidatePolicy proves the band/fallback structure at load time so the
// classifier can stay tiny. A loaded policy is trusted during evaluation.
func ValidatePolicy(p Policy) error {
	if err := validateBands(p.Bands); err != nil {
		return err
	}
	return validateFallbackActions(p.FallbackActions)
}

func validateBands(bands []Band) error {
	if len(bands) == 0 {
		return fmt.Errorf("policy has no bands")
	}

	seenRisk := make(map[string]struct{}, len(bands))
	var prevMax float64

	for i, band := range bands {
		if err := validateBandLabels(i, band, seenRisk); err != nil {
			return err
		}
		if err := validateBandBoundary(i, len(bands), band, &prevMax); err != nil {
			return err
		}
	}

	return nil
}

func validateBandLabels(i int, band Band, seenRisk map[string]struct{}) error {
	if !isValidRiskLevel(band.RiskLevel) {
		return fmt.Errorf("band %d: invalid risk_level %q", i, band.RiskLevel)
	}

	if _, ok := seenRisk[band.RiskLevel]; ok {
		return fmt.Errorf("band %d: duplicate risk_level %q", i, band.RiskLevel)
	}
	seenRisk[band.RiskLevel] = struct{}{}

	if !isValidAction(band.Action) {
		return fmt.Errorf("band %d: invalid action %q", i, band.Action)
	}

	return nil
}

func validateBandBoundary(i, bandCount int, band Band, prevMax *float64) error {
	isFinal := i == bandCount-1

	if isFinal {
		if band.MaxScore != nil {
			return fmt.Errorf("final band %d must have null max_score", i)
		}
		return nil
	}

	if band.MaxScore == nil {
		return fmt.Errorf("non-final band %d has null max_score", i)
	}

	maxScore := *band.MaxScore
	if !finite(maxScore) || maxScore <= 0 || maxScore > 1 {
		return fmt.Errorf("band %d: max_score %v out of range (0,1]", i, maxScore)
	}

	if i > 0 && maxScore <= *prevMax {
		return fmt.Errorf("band %d: max_score %v not strictly increasing (prev %v)",
			i, maxScore, *prevMax)
	}

	*prevMax = maxScore
	return nil
}

func validateFallbackActions(actions map[string]string) error {
	for _, status := range requiredFallbackStatuses {
		action, ok := actions[status]
		if !ok {
			return fmt.Errorf("fallback_actions missing %q", status)
		}
		if !isValidAction(action) {
			return fmt.Errorf("fallback_actions[%q]: invalid action %q", status, action)
		}
	}

	return nil
}

// Classify maps a normalized score to its risk level and recommended action.
// Bands are upper-exclusive: a score equal to a boundary belongs to the next
// band. The error paths are unreachable once ValidatePolicy has passed.
func (p Policy) Classify(normalized float64) (riskLevel, action string, err error) {
	if !finite(normalized) || normalized < 0 || normalized > 1 {
		return "", "", fmt.Errorf("invalid normalized score: %v", normalized)
	}
	for _, band := range p.Bands {
		if band.MaxScore == nil || normalized < *band.MaxScore {
			return band.RiskLevel, band.Action, nil
		}
	}
	return "", "", fmt.Errorf("policy has no matching band")
}

// FallbackAction returns the recommended action for a non-scored status, the
// risk level stays null for these decisions.
func (p Policy) FallbackAction(status string) (action string, err error) {
	action, ok := p.FallbackActions[status]
	if !ok {
		return "", fmt.Errorf("no fallback action for status %q", status)
	}
	return action, nil
}
