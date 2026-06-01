package domain

import (
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/validation"
)
type RawEventInput struct {
	SourceID        string    `json:"source_id"`
	VisitorID       string    `json:"visitor_id"`
	EventTime       time.Time `json:"event_time"`
	HTTPMethod      string    `json:"http_method"`
	URI             string    `json:"uri"`
	StatusCode      int       `json:"status_code"`
	ReferrerPresent bool      `json:"referrer_present"`
	UserAgent       string    `json:"user_agent,omitempty"`
	Bytes           int64     `json:"bytes,omitempty"`
}

func (e RawEventInput) Validate() error {
	var v validation.Validator

	v.Check(e.SourceID != "", "source_id required")
	v.Check(e.VisitorID != "", "visitor_id required")
	v.Check(!e.EventTime.IsZero(), "event_time required")
	v.Check(e.HTTPMethod != "", "http_method required")
	v.Check(e.URI != "", "uri required")
	v.Checkf(e.StatusCode >= 100 && e.StatusCode <= 599, "status_code out of range: %d", e.StatusCode)
	v.Check(e.Bytes >= 0, "bytes must be non-negative")

	return v.Err()
}

type RawEvent struct {
	RawEventInput

	ResourceType ResourceType `json:"resource_type"`
}

func NewRawEvent(input RawEventInput) RawEvent {
	input.EventTime = input.EventTime.UTC()

	return RawEvent{
		RawEventInput: input,
		ResourceType:  ClassifyResourceType(input.URI),
	}
}