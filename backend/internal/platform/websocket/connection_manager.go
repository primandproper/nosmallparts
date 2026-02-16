package websocket

import (
	"context"
	"sync"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	observabilityName = "websocket_connection_manager"
)

// ConnectionManager manages active WebSocket connections grouped by group ID and member ID.
type ConnectionManager struct {
	logger      logging.Logger
	tracer      tracing.Tracer
	connections map[string]map[string]*ConnectionContext
	mu          sync.RWMutex
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
) *ConnectionManager {
	return &ConnectionManager{
		logger:      logging.EnsureLogger(logger).WithName(observabilityName),
		tracer:      tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(observabilityName)),
		connections: make(map[string]map[string]*ConnectionContext),
	}
}

// Add adds a connection for a group and member.
func (m *ConnectionManager) Add(ctx context.Context, groupID, memberID string, connCtx *ConnectionContext) {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[groupID] == nil {
		m.connections[groupID] = make(map[string]*ConnectionContext)
	}
	m.connections[groupID][memberID] = connCtx
}

// Remove removes a connection.
func (m *ConnectionManager) Remove(ctx context.Context, groupID, memberID string) {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

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
func (m *ConnectionManager) Get(ctx context.Context, groupID, memberID string) *ConnectionContext {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		return groupConnections[memberID]
	}
	return nil
}

// GetGroupConnections gets all connections for a group.
func (m *ConnectionManager) GetGroupConnections(ctx context.Context, groupID string) []*ConnectionContext {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

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
func (m *ConnectionManager) BroadcastToGroup(ctx context.Context, groupID string, msg interface{}) {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		for _, connCtx := range groupConnections {
			if err := connCtx.WriteJSON(msg); err != nil {
				observability.AcknowledgeError(err, m.logger, span, "writing JSON to connection context")
			}
		}
	}
}

// BroadcastToGroupAsync runs BroadcastToGroup in a goroutine. Use when calling from a
// WebSocket read handler—writing to the same connection from its read goroutine can deadlock.
func (m *ConnectionManager) BroadcastToGroupAsync(ctx context.Context, groupID string, msg interface{}) {
	go m.BroadcastToGroup(ctx, groupID, msg)
}

// BroadcastToGroupFilteredAsync runs BroadcastToGroupFiltered in a goroutine. Use when
// calling from a WebSocket read handler.
func (m *ConnectionManager) BroadcastToGroupFilteredAsync(ctx context.Context, groupID string, msg interface{}, includeFunc func(memberID string) bool) {
	go m.BroadcastToGroupFiltered(ctx, groupID, msg, includeFunc)
}

// BroadcastToGroupFiltered sends a message to connections in a group for which includeFunc returns true.
func (m *ConnectionManager) BroadcastToGroupFiltered(ctx context.Context, groupID string, msg interface{}, includeFunc func(memberID string) bool) {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		for memberID, ctx := range groupConnections {
			if includeFunc(memberID) {
				if err := ctx.WriteJSON(msg); err != nil {
					observability.AcknowledgeError(err, m.logger, span, "writing JSON to connection context")
				}
			}
		}
	}
}

// SendToMember sends a message to a specific member in a group.
func (m *ConnectionManager) SendToMember(ctx context.Context, groupID, memberID string, msg interface{}) error {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		if ctx, ok1 := groupConnections[memberID]; ok1 {
			return ctx.WriteJSON(msg)
		}
	}
	return nil
}

// GroupHasConnections returns whether a group has any active connections.
func (m *ConnectionManager) GroupHasConnections(ctx context.Context, groupID string) bool {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		return len(groupConnections) > 0
	}
	return false
}

// GetConnectionCount returns the number of connections for a group.
func (m *ConnectionManager) GetConnectionCount(ctx context.Context, groupID string) int {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if groupConnections, ok := m.connections[groupID]; ok {
		return len(groupConnections)
	}
	return 0
}
