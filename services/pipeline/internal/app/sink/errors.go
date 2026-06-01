package sink

import (
	"errors"

	"github.com/JDucr17/streamline/services/pipeline/internal/postgres"
)
// Function to determine if postgres error can be classified as a non transient error
func isPermanent(err error) bool {
	return errors.Is(err, postgres.ErrCheckViolation) ||
		errors.Is(err, postgres.ErrNotNullViolation) ||
		errors.Is(err, postgres.ErrInvalidFormat) ||
		errors.Is(err, postgres.ErrForeignKeyViolation) ||
		errors.Is(err, postgres.ErrIntegrityViolation)
}