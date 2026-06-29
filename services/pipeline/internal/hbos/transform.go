package hbos

import (
	"fmt"
	"math"
)

// Transform names a stable serialized transform applied before histogram lookup.
type Transform string

const (
	TransformIdentity Transform = "identity"
	TransformLog1p    Transform = "log1p"
)

// TransformFunc applies a transform before histogram lookup.
type TransformFunc func(float64) float64

// CompileTransform resolves a transform into executable behavior. Shared by
// fit-time calibration and detector scoring so both apply identical transforms.
func CompileTransform(transform Transform) (TransformFunc, error) {
	switch transform {
	case TransformIdentity:
		return func(v float64) float64 { return v }, nil
	case TransformLog1p:
		return math.Log1p, nil
	default:
		return nil, fmt.Errorf("unsupported transform %q", transform)
	}
}