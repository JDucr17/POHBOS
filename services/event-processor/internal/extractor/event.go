package extractor

import "time"

// RawEvent represents an HTTP event as it arrives at the event-processor.
type RawEvent struct {
	SourceID        string
	VisitorID       string
	EventTime       time.Time
	HTTPMethod      string
	URI             string
	StatusCode      int
	ResourceType    string
	ReferrerPresent bool
}