package realtime

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
	sendBuf    = 64
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan *Event
	UserID    string
	ProjectID string
	TenantID  string
	done      chan struct{}
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, projectID, tenantID string) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan *Event, sendBuf),
		UserID:    userID,
		ProjectID: projectID,
		TenantID:  tenantID,
		done:      make(chan struct{}),
	}
}

func (c *Client) Start() {
	c.hub.Register(c)

	go c.writePump()
	go c.readPump()
}

func (c *Client) Close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}

	c.hub.Unregister(c)
	c.conn.Close()
}

func (c *Client) readPump() {
	defer c.Close()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("realtime: read error", "error", err)
			}
			return
		}

		var req struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(msg, &req) != nil {
			continue
		}

		switch req.Type {
		case "ping":
			c.writeJSON(&WSMessage{Type: "pong"})
		default:
			c.writeJSON(&WSMessage{Type: "error", Message: "unknown_message_type"})
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			msg := &WSMessage{Type: "event", Event: event}
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) writeJSON(v any) {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.conn.WriteJSON(v)
}
