package detector

import (
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
)

// temporary stub
func stubScore(event domain.RawEvent) domain.Decision {
	return domain.Decision{
		DecidedAt:  	time.Now(),
		Status:        "insufficient_history",
		Action:        "log",
		PolicyVersion: "policy-v1-stub",
	}
}