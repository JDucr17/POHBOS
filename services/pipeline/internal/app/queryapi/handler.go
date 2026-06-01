package queryapi

import (
	"net/http"
	"strings"

	"github.com/JDucr17/streamline/services/pipeline/internal/httpapi"
)

type API struct {
	store  *Store
	reader *Reader
}

func NewAPI(store *Store, reader *Reader) *API {
	return &API{
		store:  store,
		reader: reader,
	}
}

func (a *API) HandleHealth(w http.ResponseWriter, r *http.Request) *httpapi.Error {
	body := map[string]any{
		"status":        "ok",
		"visitor_count": a.store.VisitorCount(),
		"recent_size":   a.store.RecentSize(),
	}

	return httpapi.EncodeJSON(w, http.StatusOK, body)
}

func (a *API) HandleLatest(w http.ResponseWriter, r *http.Request) *httpapi.Error {
	sourceID := strings.TrimSpace(r.URL.Query().Get("source_id"))
	visitorID := strings.TrimSpace(r.URL.Query().Get("visitor_id"))

	if sourceID == "" || visitorID == "" {
		return httpapi.BadRequest("source_id and visitor_id are required")
	}

	if env, ok := a.store.GetLatest(sourceID, visitorID); ok {
		return httpapi.EncodeJSON(w, http.StatusOK, env)
	}

	env, found, err := a.reader.LatestByVisitor(r.Context(), sourceID, visitorID)
	if err != nil {
		return httpapi.ErrorFor(r, err)
	}
	if !found {
		return httpapi.NotFound("no decision for the given source_id and visitor_id")
	}

	return httpapi.EncodeJSON(w, http.StatusOK, env)
}

func (a *API) HandleList(w http.ResponseWriter, r *http.Request) *httpapi.Error {
	sourceID := strings.TrimSpace(r.URL.Query().Get("source_id"))
	if sourceID == "" {
		return httpapi.BadRequest("source_id is required")
	}

	limit, errResp := httpapi.ParseLimit(r.URL.Query().Get("limit"))
	if errResp != nil {
		return errResp
	}

	decisions, err := a.reader.ListBySource(r.Context(), sourceID, limit)
	if err != nil {
		return httpapi.ErrorFor(r, err)
	}

	return httpapi.EncodeJSON(w, http.StatusOK, decisions)
}