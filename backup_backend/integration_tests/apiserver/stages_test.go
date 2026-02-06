package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wsTestTimeout = 10 * time.Second
)

// wsTestClient is a test helper for WebSocket connections.
type wsTestClient struct {
	t              *testing.T
	conn           *websocket.Conn
	availablePlays *stages.ServerMessage
	messages       chan stages.ServerMessage
	done           chan struct{}
	stageID        string
	participantID  string
	mu             sync.Mutex
}

// newWSTestClient creates a new WebSocket test client.
func newWSTestClient(t *testing.T) *wsTestClient {
	t.Helper()

	wsURL := fmt.Sprintf("ws://localhost:%d/ws/stages", apiServiceConfig.HTTPServer.Port)
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err, "failed to connect to WebSocket")
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	client := &wsTestClient{
		t:        t,
		conn:     conn,
		messages: make(chan stages.ServerMessage, 100),
		done:     make(chan struct{}),
	}

	// Start message reader
	go client.readMessages()

	return client
}

// readMessages reads messages from the WebSocket connection.
func (c *wsTestClient) readMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				// Connection closed, exit
				return
			}

			var serverMsg stages.ServerMessage
			if err = json.Unmarshal(msg, &serverMsg); err != nil {
				c.t.Logf("failed to unmarshal message: %v", err)
				continue
			}

			c.messages <- serverMsg
		}
	}
}

// sendMessage sends a client message.
func (c *wsTestClient) sendMessage(msg *stages.ClientMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(msg)
}

// waitForMessage waits for a message of a specific type.
// Non-matching messages are discarded (not put back) to avoid race conditions.
func (c *wsTestClient) waitForMessage(msgType stages.ServerMessageType, timeout time.Duration) (*stages.ServerMessage, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case msg := <-c.messages:
			if msg.Type == msgType {
				return &msg, nil
			}
			// Non-matching messages are discarded to avoid race conditions
			// Tests should be designed to consume messages in expected order
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for message type %s", msgType)
		}
	}
}

// waitForMessageWithFilter waits for a message that matches a filter.
func (c *wsTestClient) waitForMessageWithFilter(filter func(msg *stages.ServerMessage) bool, timeout time.Duration) (*stages.ServerMessage, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case msg := <-c.messages:
			if filter(&msg) {
				return &msg, nil
			}
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for filtered message")
		}
	}
}

// close closes the WebSocket connection.
func (c *wsTestClient) close() {
	close(c.done)
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close()
}

// joinStage sends a join message and waits for confirmation.
func (c *wsTestClient) joinStage(stageID, participantName string) (*stages.StageState, error) {
	err := c.sendMessage(&stages.ClientMessage{
		Type:            stages.ClientMessageTypeJoin,
		StageID:         stageID,
		ParticipantName: participantName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send join message: %w", err)
	}

	// The server sends messages in this order after join:
	// 1. joined (with stageState)
	// 2. availablePlays
	// 3. stageState broadcast
	// We need to consume them all to avoid race conditions in subsequent tests

	// Collect all expected messages with a short timeout
	timeout := 2 * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var joinedMsg *stages.ServerMessage
	messagesCollected := 0
	expectedMessages := 3 // joined, availablePlays, stageState

	for messagesCollected < expectedMessages {
		select {
		case msg := <-c.messages:
			switch msg.Type {
			case stages.ServerMessageTypeJoined:
				msgCopy := msg
				joinedMsg = &msgCopy
				c.participantID = msg.ParticipantID
				if msg.StageState != nil {
					c.stageID = msg.StageState.ID
				}
			case stages.ServerMessageTypeAvailablePlays:
				msgCopy := msg
				c.availablePlays = &msgCopy
			case stages.ServerMessageTypeStageState:
				// Just consume it
			}
			messagesCollected++
		case <-timer.C:
			// Timeout, return what we have
			if joinedMsg != nil && joinedMsg.StageState != nil {
				return joinedMsg.StageState, nil
			}
			return nil, fmt.Errorf("timeout waiting for join messages, got %d of %d", messagesCollected, expectedMessages)
		}
	}

	if joinedMsg != nil && joinedMsg.StageState != nil {
		return joinedMsg.StageState, nil
	}

	return nil, fmt.Errorf("no stage state in joined message")
}

func TestStages_CreateAndJoin(T *testing.T) {
	T.Parallel()

	T.Run("create new stage by joining with empty ID", func(t *testing.T) {
		t.Parallel()

		client := newWSTestClient(t)
		defer client.close()

		// Join with empty stage ID to create a new stage
		state, err := client.joinStage("", "TestPlayer1")
		require.NoError(t, err)
		require.NotNil(t, state)

		assert.NotEmpty(t, state.ID, "stage should have an ID")
		assert.NotEmpty(t, state.Code, "stage should have a code")
		assert.Len(t, state.Participants, 1, "stage should have one participant")
		assert.Equal(t, "TestPlayer1", state.Participants[0].DisplayName)
		assert.Equal(t, stages.StagePhaseLobby, state.Phase)
	})

	T.Run("join existing stage by code", func(t *testing.T) {
		t.Parallel()

		// First client creates a stage
		client1 := newWSTestClient(t)
		defer client1.close()

		state1, err := client1.joinStage("", "Player1")
		require.NoError(t, err)
		require.NotNil(t, state1)

		stageCode := state1.Code
		require.NotEmpty(t, stageCode)

		// Second client joins the same stage
		client2 := newWSTestClient(t)
		defer client2.close()

		state2, err := client2.joinStage(stageCode, "Player2")
		require.NoError(t, err)
		require.NotNil(t, state2)

		assert.Equal(t, state1.ID, state2.ID, "both clients should be in the same stage")
		assert.Len(t, state2.Participants, 2, "stage should have two participants")
	})
}

func TestStages_ParticipantUpdates(T *testing.T) {
	T.Parallel()

	T.Run("set participant as director", func(t *testing.T) {
		t.Parallel()

		client := newWSTestClient(t)
		defer client.close()

		state, err := client.joinStage("", "Director")
		require.NoError(t, err)
		require.NotNil(t, state)

		participantID := client.participantID
		require.NotEmpty(t, participantID)

		// Set as director
		isDirector := true
		err = client.sendMessage(&stages.ClientMessage{
			Type:          stages.ClientMessageTypeSetDirector,
			ParticipantID: participantID,
			IsDirector:    &isDirector,
		})
		require.NoError(t, err)

		// Wait for state update
		msg, err := client.waitForMessage(stages.ServerMessageTypeStageState, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg.State)

		// Find our participant and verify director status
		var found bool
		for _, p := range msg.State.Participants {
			if p.ID == participantID {
				assert.True(t, p.IsDirector, "participant should be director")
				found = true
				break
			}
		}
		assert.True(t, found, "participant should be in state")
	})

	T.Run("set participant as audience", func(t *testing.T) {
		t.Parallel()

		client := newWSTestClient(t)
		defer client.close()

		state, err := client.joinStage("", "AudienceMember")
		require.NoError(t, err)
		require.NotNil(t, state)

		participantID := client.participantID
		require.NotEmpty(t, participantID)

		// Set as audience
		isAudience := true
		err = client.sendMessage(&stages.ClientMessage{
			Type:          stages.ClientMessageTypeSetAudience,
			ParticipantID: participantID,
			IsAudience:    &isAudience,
		})
		require.NoError(t, err)

		// Wait for state update
		msg, err := client.waitForMessage(stages.ServerMessageTypeStageState, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg.State)

		// Find our participant and verify audience status
		var found bool
		for _, p := range msg.State.Participants {
			if p.ID == participantID {
				assert.True(t, p.IsAudience, "participant should be audience")
				found = true
				break
			}
		}
		assert.True(t, found, "participant should be in state")
	})

	T.Run("set participant ready", func(t *testing.T) {
		t.Parallel()

		client := newWSTestClient(t)
		defer client.close()

		state, err := client.joinStage("", "ReadyPlayer")
		require.NoError(t, err)
		require.NotNil(t, state)

		participantID := client.participantID
		require.NotEmpty(t, participantID)

		// Set as ready
		isReady := true
		err = client.sendMessage(&stages.ClientMessage{
			Type:          stages.ClientMessageTypeSetReady,
			ParticipantID: participantID,
			IsReady:       &isReady,
		})
		require.NoError(t, err)

		// Wait for state update
		msg, err := client.waitForMessage(stages.ServerMessageTypeStageState, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg.State)

		// Find our participant and verify ready status
		var found bool
		for _, p := range msg.State.Participants {
			if p.ID == participantID {
				assert.True(t, p.IsReady, "participant should be ready")
				found = true
				break
			}
		}
		assert.True(t, found, "participant should be in state")
	})
}

func TestStages_MultiClientBroadcast(T *testing.T) {
	T.Parallel()

	T.Run("state changes broadcast to all participants", func(t *testing.T) {
		t.Parallel()

		// First client creates a stage
		client1 := newWSTestClient(t)
		defer client1.close()

		state1, err := client1.joinStage("", "Player1")
		require.NoError(t, err)
		require.NotNil(t, state1)

		stageCode := state1.Code

		// Second client joins
		client2 := newWSTestClient(t)
		defer client2.close()

		_, err = client2.joinStage(stageCode, "Player2")
		require.NoError(t, err)

		// Client1 should receive a stageState update when client2 joins
		msg, err := client1.waitForMessage(stages.ServerMessageTypeStageState, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg.State)
		assert.Len(t, msg.State.Participants, 2, "state should show 2 participants")

		// Client1 sets themselves as director
		isDirector := true
		err = client1.sendMessage(&stages.ClientMessage{
			Type:          stages.ClientMessageTypeSetDirector,
			ParticipantID: client1.participantID,
			IsDirector:    &isDirector,
		})
		require.NoError(t, err)

		// Client2 should receive the state update
		msg2, err := client2.waitForMessage(stages.ServerMessageTypeStageState, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg2.State)

		// Verify client1 is marked as director in the state client2 received
		var found bool
		for _, p := range msg2.State.Participants {
			if p.ID == client1.participantID {
				assert.True(t, p.IsDirector, "player1 should be director in player2's view")
				found = true
				break
			}
		}
		assert.True(t, found, "player1 should be in player2's state")
	})

	T.Run("participant disconnect updates remaining participants", func(t *testing.T) {
		t.Parallel()

		// First client creates a stage
		client1 := newWSTestClient(t)
		defer client1.close()

		state1, err := client1.joinStage("", "Stayer")
		require.NoError(t, err)
		require.NotNil(t, state1)

		stageCode := state1.Code

		// Second client joins
		client2 := newWSTestClient(t)

		_, err = client2.joinStage(stageCode, "Leaver")
		require.NoError(t, err)

		// Client1 should receive state update showing 2 participants
		msg1, err := client1.waitForMessage(stages.ServerMessageTypeStageState, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg1.State)
		assert.Len(t, msg1.State.Participants, 2)

		// Client2 disconnects
		client2.close()

		// Client1 should receive state update showing 1 participant
		msg2, err := client1.waitForMessageWithFilter(func(msg *stages.ServerMessage) bool {
			return msg.Type == stages.ServerMessageTypeStageState &&
				msg.State != nil &&
				len(msg.State.Participants) == 1
		}, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg2.State)
		assert.Len(t, msg2.State.Participants, 1, "should only have 1 participant after disconnect")
	})
}

func TestStages_ChatMessages(T *testing.T) {
	T.Parallel()

	T.Run("chat messages broadcast to all participants", func(t *testing.T) {
		t.Parallel()

		// Create stage with two participants
		client1 := newWSTestClient(t)
		defer client1.close()

		state1, err := client1.joinStage("", "Chatter1")
		require.NoError(t, err)

		client2 := newWSTestClient(t)
		defer client2.close()

		_, err = client2.joinStage(state1.Code, "Chatter2")
		require.NoError(t, err)

		// Drain the join state updates
		_, _ = client1.waitForMessage(stages.ServerMessageTypeStageState, 2*time.Second)

		// Client1 sends a chat message
		err = client1.sendMessage(&stages.ClientMessage{
			Type:          stages.ClientMessageTypeSendChat,
			ParticipantID: client1.participantID,
			Text:          "Hello from Player1!",
			CastOnly:      false,
		})
		require.NoError(t, err)

		// Client2 should receive the chat message
		msg, err := client2.waitForMessage(stages.ServerMessageTypeChatMessage, wsTestTimeout)
		require.NoError(t, err)
		require.NotNil(t, msg.Message)

		assert.Equal(t, "Hello from Player1!", msg.Message.Text)
		assert.Equal(t, "Chatter1", msg.Message.SenderName)
		assert.False(t, msg.Message.CastOnly)
	})
}

func TestStages_AvailablePlays(T *testing.T) {
	T.Parallel()

	T.Run("receives available plays on join", func(t *testing.T) {
		t.Parallel()

		client := newWSTestClient(t)
		defer client.close()

		_, err := client.joinStage("", "PlaySelector")
		require.NoError(t, err)

		// The availablePlays message is captured during joinStage
		require.NotNil(t, client.availablePlays, "should have received availablePlays message")
		assert.Equal(t, stages.ServerMessageTypeAvailablePlays, client.availablePlays.Type)

		// Plays list may be empty or nil if none are public, which is acceptable
		// The important thing is that the message was received
	})
}
