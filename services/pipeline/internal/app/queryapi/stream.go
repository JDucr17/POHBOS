package queryapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
	"github.com/JDucr17/streamline/services/pipeline/internal/httpapi"
)

// heartbeatInterval keeps idle SSE connections alive through proxies and doubles
// as a liveness probe: a failed heartbeat write reveals a dead connection.
const heartbeatInterval = 20 * time.Second

type streamConfig struct {
	backfillLimit int
	allowedOrigin string
}

// streamRequest is the validated, accepted shape of a stream request, resolved
// before any capacity slot is taken.
type streamRequest struct {
	sourceID      string
	backfill      int
	matchedOrigin string
}

// HandleStream serves the live decision feed over SSE. Every deterministic
// rejection happens before Subscribe, so a rejected request never consumes a
// capacity slot. Once the stream is opened (headers + first flush) the handler
// can no longer send an HTTP status, so later failures end the stream by
// returning nil and letting the deferred Unsubscribe clean up.
func (a *API) HandleStream(w http.ResponseWriter, r *http.Request) *httpapi.Error {
	request, errResp := a.resolveStreamRequest(r)
	if errResp != nil {
		return errResp
	}

	subscription, ok := a.hub.Subscribe(request.sourceID)
	if !ok {
		return tooManyStreamClients(w)
	}
	defer a.hub.Unsubscribe(subscription)

	streamController := http.NewResponseController(w)
	if err := disableStreamWriteDeadline(streamController); err != nil {
		return &httpapi.Error{
			Err:     err,
			Message: "streaming not supported",
			Status:  http.StatusInternalServerError,
		}
	}

	writeStreamHeaders(w, request.matchedOrigin)

	if err := writeAndFlushSSEComment(w, streamController, "connected"); err != nil {
		return nil
	}

	if err := a.writeBackfill(w, streamController, request); err != nil {
		return nil
	}

	a.writeLiveStream(w, r, streamController, subscription)
	return nil
}

func (a *API) resolveStreamRequest(request *http.Request) (streamRequest, *httpapi.Error) {
	backfill, errResp := parseBackfill(request.URL.Query().Get("backfill"), a.stream.backfillLimit)
	if errResp != nil {
		return streamRequest{}, errResp
	}

	matchedOrigin, errResp := a.matchAllowedOrigin(request)
	if errResp != nil {
		return streamRequest{}, errResp
	}

	return streamRequest{
		sourceID:      strings.TrimSpace(request.URL.Query().Get("source_id")),
		backfill:      backfill,
		matchedOrigin: matchedOrigin,
	}, nil
}

// matchAllowedOrigin applies CORS only when the request carries an Origin: curl,
// same-origin.
func (a *API) matchAllowedOrigin(request *http.Request) (string, *httpapi.Error) {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return "", nil
	}

	if origin != a.stream.allowedOrigin {
		return "", &httpapi.Error{
			Err:     errors.New("origin not allowed"),
			Message: "origin not allowed",
			Status:  http.StatusForbidden,
		}
	}

	return origin, nil
}

func tooManyStreamClients(w http.ResponseWriter) *httpapi.Error {
	w.Header().Set("Retry-After", "5")
	return &httpapi.Error{
		Err:     errors.New("sse client capacity reached"),
		Message: "too many connections",
		Status:  http.StatusTooManyRequests,
	}
}

// disableStreamWriteDeadline clears the deadline before the first byte: SSE is
// long-lived and must not inherit the finite per-response write timeout.
func disableStreamWriteDeadline(streamController *http.ResponseController) error {
	return streamController.SetWriteDeadline(time.Time{})
}

// writeStreamHeaders sets the proxy-safe SSE headers, without them proxies buffer
// the stream and the browser sees nothing.
func writeStreamHeaders(w http.ResponseWriter, matchedOrigin string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	if matchedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", matchedOrigin)
		// Add, not Set, so a Vary set elsewhere is not clobbered.
		w.Header().Add("Vary", "Origin")
	}
}

func (a *API) writeBackfill(writer io.Writer, streamController *http.ResponseController, request streamRequest) error {
	for _, decision := range a.store.RecentForSourceInArrivalOrder(request.sourceID, request.backfill) {
		if err := writeAndFlushSSEDecision(writer, streamController, decision); err != nil {
			return err
		}
	}
	return nil
}

func (a *API) writeLiveStream(writer io.Writer, request *http.Request, streamController *http.ResponseController, subscription *subscription) {
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case decision, open := <-subscription.decisions:
			if !open {
				// PublishBatch closed us as a slow client, reconnect re-backfills.
				return
			}
			if err := writeAndFlushSSEDecision(writer, streamController, decision); err != nil {
				return
			}

		case <-heartbeat.C:
			if err := writeAndFlushSSEComment(writer, streamController, "heartbeat"); err != nil {
				return
			}

		case <-request.Context().Done():
			return
		}
	}
}

// writeSSEDecision writes one SSE frame: a single JSON line in the data field,
// terminated by a blank line. json.Marshal (not Encoder.Encode) keeps the payload
// single-line so the framing stays exact.
func writeSSEDecision(writer io.Writer, decision envelope.DecisionEnvelope) error {
	payload, err := json.Marshal(decision)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(
		writer,
		"id: %s\nevent: decision\ndata: %s\n\n",
		decision.DecisionID,
		payload,
	)
	return err
}

func writeAndFlushSSEDecision(writer io.Writer, streamController *http.ResponseController, decision envelope.DecisionEnvelope) error {
	if err := writeSSEDecision(writer, decision); err != nil {
		return err
	}
	return streamController.Flush()
}

func writeAndFlushSSEComment(writer io.Writer, streamController *http.ResponseController, comment string) error {
	if _, err := fmt.Fprintf(writer, ": %s\n\n", comment); err != nil {
		return err
	}
	return streamController.Flush()
}

// parseBackfill resolves the backfill count: missing uses the configured limit, a
// value above it clamps, an explicit 0 means no backfill, and a negative value is
// malformed input rather than a request for none.
func parseBackfill(raw string, limit int) (int, *httpapi.Error) {
	if raw == "" {
		return limit, nil
	}

	count, err := strconv.Atoi(raw)
	if err != nil {
		return 0, httpapi.BadRequest("backfill must be an integer")
	}
	if count < 0 {
		return 0, httpapi.BadRequest("backfill must not be negative")
	}
	if count > limit {
		return limit, nil
	}

	return count, nil
}
