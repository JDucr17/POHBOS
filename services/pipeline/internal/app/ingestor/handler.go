package ingestor

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JDucr17/streamline/services/pipeline/internal/broker"
	"github.com/JDucr17/streamline/services/pipeline/internal/domain"
	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
	"github.com/JDucr17/streamline/services/pipeline/internal/httpapi"
)

const healthTimeout = 2 * time.Second

// API holds the dependencies the ingestor handlers need
type API struct {
	Producer *broker.Producer
}

// HandleEvent decodes a RawEvent, builds an EventEnvelope, publishes it to
// the broker, and returns 202 Accepted. The decision itself is produced
// asynchronously by the detector and made available via the query-api
func (a *API) HandleEvent(w http.ResponseWriter, r *http.Request) *httpapi.Error {
	input, errResp := httpapi.DecodeAndValidate[domain.RawEventInput](w, r, httpapi.MaxEventBodyBytes)
	if errResp != nil {
		return errResp
	}

	id, err := uuid.NewV7()
	if err != nil {
		return httpapi.ErrorFor(r, err)
	}

	event := domain.NewRawEvent(input)

	env := envelope.EventEnvelope{
		EventID:       id.String(),
		IngestedAt:    time.Now().UnixMilli(),
		SchemaVersion: envelope.SchemaVersion,
		SourceID:      event.SourceID,
		VisitorID:     event.VisitorID,
		Payload:       event,
	}

	body, err := json.Marshal(env)
	if err != nil {
		return httpapi.ErrorFor(r, err)
	}

	key := []byte(event.SourceID + ":" + event.VisitorID)

	if err := a.Producer.Publish(r.Context(), key, body); err != nil {
		return httpapi.ErrorFor(r, err)
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}

// HandleHealth pings the broker. Returns 200 when reachable, 503 with a
// JSON detail body otherwise
func (a *API) HandleHealth(w http.ResponseWriter, r *http.Request) *httpapi.Error {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()

	if err := a.Producer.Ping(ctx); err != nil {
		body := map[string]string{"broker": "unavailable"}
		return httpapi.EncodeJSON(w, http.StatusServiceUnavailable, body)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}