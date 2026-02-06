package websocket

import (
	"sync"
)

// ConnectionManager manages active WebSocket connections grouped by group ID and member ID.
type ConnectionManager struct {
	connections map[string]map[string]*ConnectionContext
	mu          sync.RWMutex
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]map[string]*ConnectionContext),
	}
}

// Add adds a connection for a group and member.
func (m *ConnectionManager) Add(groupID, memberID string, ctx *ConnectionContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[groupID] == nil {
		m.connections[groupID] = make(map[string]*ConnectionContext)
	}
	m.connections[groupID][memberID] = ctx
}

// Remove removes a connection.
func (m *ConnectionManager) Remove(groupID, memberID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		delete(groupConnections, memberID)
		if len(groupConnections) == 0 {
			delete(m.connections, groupID)
		}
	}
}

// Get gets a specific connection.
func (m *ConnectionManager) Get(groupID, memberID string) *ConnectionContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		return groupConnections[memberID]
	}
	return nil
}

// GetGroupConnections gets all connections for a group.
func (m *ConnectionManager) GetGroupConnections(groupID string) []*ConnectionContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var connections []*ConnectionContext
	if groupConnections, ok := m.connections[groupID]; ok {
		for _, ctx := range groupConnections {
			connections = append(connections, ctx)
		}
	}
	return connections
}

// BroadcastToGroup sends a message to all connections in a group.
func (m *ConnectionManager) BroadcastToGroup(groupID string, msg interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		for _, ctx := range groupConnections {
			_ = ctx.WriteJSON(msg)
		}
	}
}

// BroadcastToGroupFiltered sends a message to connections in a group for which includeFunc returns true.
func (m *ConnectionManager) BroadcastToGroupFiltered(groupID string, msg interface{}, includeFunc func(memberID string) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		for memberID, ctx := range groupConnections {
			if includeFunc(memberID) {
				_ = ctx.WriteJSON(msg)
			}
		}
	}
}

// SendToMember sends a message to a specific member in a group.
func (m *ConnectionManager) SendToMember(groupID, memberID string, msg interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		if ctx, ok := groupConnections[memberID]; ok {
			return ctx.WriteJSON(msg)
		}
	}
	return nil
}

// GroupHasConnections returns whether a group has any active connections.
func (m *ConnectionManager) GroupHasConnections(groupID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		return len(groupConnections) > 0
	}
	return false
}

// GetConnectionCount returns the number of connections for a group.
func (m *ConnectionManager) GetConnectionCount(groupID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		return len(groupConnections)
	}
	return 0
}
