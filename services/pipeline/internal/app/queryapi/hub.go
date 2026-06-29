package queryapi

import (
	"log/slog"
	"sync"

	"github.com/JDucr17/streamline/services/pipeline/internal/envelope"
)

type subscriptionScope struct {
	allSources bool
	sourceID   string
}

func allSourceSubscriptions() subscriptionScope {
	return subscriptionScope{allSources: true}
}

func sourceSubscriptions(sourceID string) subscriptionScope {
	return subscriptionScope{sourceID: sourceID}
}

// subscription is one connected SSE client. Its scope determines whether it
// receives every decision or only decisions for one source.
type subscription struct {
	decisions chan envelope.DecisionEnvelope
	scope     subscriptionScope
}

// Hub fans out consumed decisions to SSE clients. Subscriptions are bucketed by
// scope, so "all sources" is structurally distinct from any concrete source ID.
type Hub struct {
	mu sync.Mutex

	subscriptionsByScope map[subscriptionScope]map[*subscription]struct{}
	clientCount          int

	maxClients   int
	clientBuffer int
}

func NewHub(maxClients, clientBuffer int) *Hub {
	return &Hub{
		subscriptionsByScope: make(map[subscriptionScope]map[*subscription]struct{}),
		maxClients:          maxClients,
		clientBuffer:        clientBuffer,
	}
}

// Subscribe registers a client, returning ok=false at capacity so the caller can
// reject the connection before streaming. A blank sourceID subscribes to all
// sources.
func (h *Hub) Subscribe(sourceID string) (*subscription, bool) {
	if sourceID == "" {
		return h.subscribe(allSourceSubscriptions())
	}

	return h.subscribe(sourceSubscriptions(sourceID))
}

func (h *Hub) subscribe(scope subscriptionScope) (*subscription, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clientCount >= h.maxClients {
		return nil, false
	}

	client := &subscription{
		decisions: make(chan envelope.DecisionEnvelope, h.clientBuffer),
		scope:     scope,
	}

	subscribers := h.subscriptionsByScope[scope]
	if subscribers == nil {
		subscribers = make(map[*subscription]struct{})
		h.subscriptionsByScope[scope] = subscribers
	}

	subscribers[client] = struct{}{}
	h.clientCount++

	return client, true
}

// Unsubscribe removes and closes a client. It is idempotent: PublishBatch may
// have already dropped a slow client, so removing again is a no-op.
func (h *Hub) Unsubscribe(client *subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.removeSubscription(client)
}

// PublishBatch fans out decisions with non-blocking sends. A blocking send would
// stall Kafka consumption behind one slow browser, so a client whose buffer is
// full is disconnected instead.
func (h *Hub) PublishBatch(decisions []envelope.DecisionEnvelope) {
	slowDisconnects := 0

	h.mu.Lock()
	for _, decision := range decisions {
		slowDisconnects += h.sendDecisionToSubscribers(allSourceSubscriptions(), decision)

		if decision.SourceID != "" {
			slowDisconnects += h.sendDecisionToSubscribers(sourceSubscriptions(decision.SourceID), decision)
		}
	}
	h.mu.Unlock()

	if slowDisconnects > 0 {
		slog.Debug("sse slow clients disconnected", slog.Int("count", slowDisconnects))
	}
}

func (h *Hub) ClientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.clientCount
}

func (h *Hub) MaxClients() int {
	return h.maxClients
}

// The helpers below run while h.mu is held. They mutate subscription buckets,
// clientCount, and decision queues, so callers must keep them inside the hub's
// critical section.
func (h *Hub) removeSubscription(client *subscription) {
	subscribers := h.subscriptionsByScope[client.scope]
	if subscribers == nil {
		return
	}
	if _, present := subscribers[client]; !present {
		return
	}

	delete(subscribers, client)
	if len(subscribers) == 0 {
		delete(h.subscriptionsByScope, client.scope)
	}

	h.clientCount--
	close(client.decisions)
}

func (h *Hub) sendDecisionOrDisconnect(
	client *subscription,
	decision envelope.DecisionEnvelope,
) bool {
	select {
	case client.decisions <- decision:
		return false
	default:
		h.removeSubscription(client)
		return true
	}
}

// sendDecisionToSubscribers sends one decision to one subscriber bucket and
// returns the number of clients disconnected as slow.
func (h *Hub) sendDecisionToSubscribers(scope subscriptionScope, decision envelope.DecisionEnvelope) int {
	slowDisconnects := 0

	subscribers := h.subscriptionsByScope[scope]
	for client := range subscribers {
		if h.sendDecisionOrDisconnect(client, decision) {
			slowDisconnects++
		}
	}

	return slowDisconnects
}