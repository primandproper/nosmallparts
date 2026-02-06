package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionContext represents a WebSocket connection with its associated metadata.
// GroupID and MemberID are generic identifiers for grouping connections (e.g. stage ID and participant ID).
type ConnectionContext struct {
	Conn              *websocket.Conn
	heartbeatInterval *time.Ticker
	done              chan struct{}
	GroupID           string
	MemberID          string
	mu                sync.Mutex
	IsAlive           bool
}

// NewConnectionContext creates a new ConnectionContext.
func NewConnectionContext(conn *websocket.Conn) *ConnectionContext {
	return &ConnectionContext{
		Conn:    conn,
		IsAlive: true,
		done:    make(chan struct{}),
	}
}

// SetGroupAndMember sets the group and member IDs for this connection.
func (c *ConnectionContext) SetGroupAndMember(groupID, memberID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GroupID = groupID
	c.MemberID = memberID
}

// SetHeartbeatTicker sets the heartbeat ticker so it can be stopped when the connection closes.
func (c *ConnectionContext) SetHeartbeatTicker(ticker *time.Ticker) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.heartbeatInterval = ticker
}

// Done returns a channel that is closed when the connection is closed.
func (c *ConnectionContext) Done() <-chan struct{} {
	return c.done
}

// MarkAlive marks the connection as alive (received pong).
func (c *ConnectionContext) MarkAlive() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsAlive = true
}

// MarkDead marks the connection as dead (no pong received).
func (c *ConnectionContext) MarkDead() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsAlive = false
}

// CheckAlive returns whether the connection is alive.
func (c *ConnectionContext) CheckAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.IsAlive
}

// Close closes the connection and stops the heartbeat.
func (c *ConnectionContext) Close() {
	close(c.done)
	c.mu.Lock()
	ticker := c.heartbeatInterval
	c.mu.Unlock()
	if ticker != nil {
		ticker.Stop()
	}

	if err := c.Conn.Close(); err != nil {
		log.Printf("\n\nError closing websocket connection: %v\n\n", err)
	}
}

// WriteJSON writes a JSON message to the connection.
func (c *ConnectionContext) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}
