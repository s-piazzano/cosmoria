package realtime

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Hub struct {
	publisher  *Publisher
	subscriber *Subscriber

	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
	eventCh chan *Event
}

func NewHub(pool *pgxpool.Pool) *Hub {
	eventCh := make(chan *Event, 1024)
	h := &Hub{
		publisher:  NewPublisher(pool),
		subscriber: NewSubscriber(pool, eventCh),
		clients:    make(map[string]map[*Client]struct{}),
		eventCh:    eventCh,
	}
	return h
}

func (h *Hub) Start(ctx context.Context) {
	h.subscriber.Start(ctx)
	go h.fanOut(ctx)
}

func (h *Hub) Stop() {
	h.subscriber.Stop()
}

func (h *Hub) Publisher() *Publisher {
	return h.publisher
}

func (h *Hub) Register(client *Client) {
	h.subscriber.AddChannel(client.ProjectID)

	h.mu.Lock()
	set, ok := h.clients[client.ProjectID]
	if !ok {
		set = make(map[*Client]struct{})
		h.clients[client.ProjectID] = set
	}
	set[client] = struct{}{}
	h.mu.Unlock()

	slog.Debug("realtime: client registered",
		"project", client.ProjectID,
		"tenant", client.TenantID,
		"user", client.UserID,
	)
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	set, ok := h.clients[client.ProjectID]
	if !ok {
		h.mu.Unlock()
		return
	}
	delete(set, client)
	if len(set) == 0 {
		delete(h.clients, client.ProjectID)
		h.subscriber.RemoveChannel(client.ProjectID)
	}
	h.mu.Unlock()

	slog.Debug("realtime: client unregistered",
		"project", client.ProjectID,
		"tenant", client.TenantID,
		"user", client.UserID,
	)
}

func (h *Hub) fanOut(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-h.eventCh:
			h.mu.RLock()
			set, ok := h.clients[event.ProjectID]
			h.mu.RUnlock()

			if !ok {
				continue
			}
			for client := range set {
				if client.TenantID == event.TenantID {
					select {
					case client.send <- event:
					default:
						slog.Warn("realtime: client send buffer full, dropping",
							"project", event.ProjectID,
							"tenant", event.TenantID,
							"user", client.UserID,
						)
					}
				}
			}
		}
	}
}
