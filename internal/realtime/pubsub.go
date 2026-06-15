package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const channelPrefix = "cosm_"

func channelName(projectID string) string {
	return channelPrefix + projectID
}

type Publisher struct {
	pool *pgxpool.Pool
}

func NewPublisher(pool *pgxpool.Pool) *Publisher {
	return &Publisher{pool: pool}
}

func (p *Publisher) Publish(event *Event) {
	go func() {
		data, err := json.Marshal(event)
		if err != nil {
			slog.Error("realtime: marshal event", "error", err)
			return
		}

		_, err = p.pool.Exec(context.Background(),
			"SELECT pg_notify($1, $2)",
			channelName(event.ProjectID), string(data),
		)
		if err != nil {
			slog.Error("realtime: publish failed",
				"project", event.ProjectID,
				"resource", event.Resource,
				"action", event.Action,
				"error", err,
			)
		}
	}()
}

type Subscriber struct {
	pool    *pgxpool.Pool
	eventCh chan<- *Event
	addCh   chan string
	removeCh chan string
	done    chan struct{}
}

func NewSubscriber(pool *pgxpool.Pool, eventCh chan<- *Event) *Subscriber {
	return &Subscriber{
		pool:     pool,
		eventCh:  eventCh,
		addCh:    make(chan string, 64),
		removeCh: make(chan string, 64),
		done:     make(chan struct{}),
	}
}

func (s *Subscriber) Start(ctx context.Context) {
	go s.listenLoop(ctx)
}

func (s *Subscriber) Stop() {
	close(s.done)
}

func (s *Subscriber) AddChannel(projectID string) {
	select {
	case s.addCh <- projectID:
	default:
	}
}

func (s *Subscriber) RemoveChannel(projectID string) {
	select {
	case s.removeCh <- projectID:
	default:
	}
}

func (s *Subscriber) listenLoop(ctx context.Context) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		slog.Error("realtime: acquire dedicated conn", "error", err)
		return
	}
	defer conn.Release()

	pgxConn := conn.Conn()
	active := make(map[string]bool)

	pendingAdd := func() {
		for {
			select {
			case ch := <-s.addCh:
				name := channelName(ch)
				if !active[name] {
					_, err := pgxConn.Exec(ctx, "LISTEN "+name)
					if err != nil {
						slog.Error("realtime: LISTEN failed", "channel", name, "error", err)
						continue
					}
					active[name] = true
					slog.Debug("realtime: LISTEN", "channel", name)
				}
			default:
				return
			}
		}
	}

	pendingRemove := func() {
		for {
			select {
			case ch := <-s.removeCh:
				name := channelName(ch)
				if active[name] {
					_, err := pgxConn.Exec(ctx, "UNLISTEN "+name)
					if err != nil {
						slog.Error("realtime: UNLISTEN failed", "channel", name, "error", err)
					}
					delete(active, name)
					slog.Debug("realtime: UNLISTEN", "channel", name)
				}
			default:
				return
			}
		}
	}

	for {
		select {
		case <-s.done:
			return
		default:
		}

		pendingAdd()
		pendingRemove()

		notifCtx, notifCancel := context.WithTimeout(ctx, 5*time.Second)
		notif, err := pgxConn.WaitForNotification(notifCtx)
		notifCancel()

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		var event Event
		if err := json.Unmarshal([]byte(notif.Payload), &event); err != nil {
			slog.Error("realtime: parse notification", "error", err)
			continue
		}

		select {
		case s.eventCh <- &event:
		default:
		}
	}
}
