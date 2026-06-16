package realtime

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupHubTest(t *testing.T) (*pgxpool.Pool, *Hub) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	hub := NewHub(pool)
	hub.Start(context.Background())

	t.Cleanup(func() {
		hub.Stop()
	})

	return pool, hub
}

func newTestClient(hub *Hub, projectID, tenantID, userID string) *Client {
	return &Client{
		hub:       hub,
		send:      make(chan *Event, 64),
		UserID:    userID,
		ProjectID: projectID,
		TenantID:  tenantID,
		done:      make(chan struct{}),
	}
}

func TestHub_RegisterAndPublish(t *testing.T) {
	_, hub := setupHubTest(t)

	client := newTestClient(hub, "p1", "t1", "u1")
	hub.Register(client)
	defer hub.Unregister(client)

	hub.Publisher().Publish(&Event{
		ProjectID: "p1",
		TenantID:  "t1",
		Resource:  "files",
		Action:    "create",
	})

	select {
	case event := <-client.send:
		assert.Equal(t, "files", event.Resource)
		assert.Equal(t, "create", event.Action)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestHub_FanOut_OnlyMatchingTenant(t *testing.T) {
	_, hub := setupHubTest(t)

	c1 := newTestClient(hub, "p1", "t1", "u1")
	c2 := newTestClient(hub, "p1", "t2", "u2")

	hub.Register(c1)
	hub.Register(c2)
	defer hub.Unregister(c1)
	defer hub.Unregister(c2)

	hub.Publisher().Publish(&Event{
		ProjectID: "p1",
		TenantID:  "t1",
		Resource:  "tenants",
		Action:    "update",
	})

	select {
	case <-c1.send:
		// ok — c1 matches tenant t1
	case <-time.After(time.Second):
		t.Fatal("c1 should receive event for tenant t1")
	}

	select {
	case <-c2.send:
		t.Fatal("c2 should NOT receive event for tenant t1")
	default:
		// ok — c2 should not receive
	}
}

func TestHub_Unregister_StopsReceiving(t *testing.T) {
	_, hub := setupHubTest(t)

	client := newTestClient(hub, "p1", "t1", "u1")
	hub.Register(client)
	hub.Unregister(client)

	hub.Publisher().Publish(&Event{
		ProjectID: "p1",
		TenantID:  "t1",
		Resource:  "files",
		Action:    "delete",
	})

	select {
	case <-client.send:
		t.Fatal("unregistered client should not receive events")
	default:
		// ok
	}
}

func TestHub_MultipleClientsSameProject(t *testing.T) {
	_, hub := setupHubTest(t)

	c1 := newTestClient(hub, "p1", "t1", "u1")
	c2 := newTestClient(hub, "p1", "t1", "u2")

	hub.Register(c1)
	hub.Register(c2)

	hub.Publisher().Publish(&Event{
		ProjectID: "p1",
		TenantID:  "t1",
		Resource:  "tenants",
		Action:    "create",
	})

	select {
	case <-c1.send:
	case <-time.After(time.Second):
		t.Fatal("c1 should receive")
	}

	select {
	case <-c2.send:
	case <-time.After(time.Second):
		t.Fatal("c2 should also receive")
	}

	hub.Unregister(c1)
	hub.Unregister(c2)
}
