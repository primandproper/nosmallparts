package stages

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
)

// ClientMessageType represents the type of client message.
type ClientMessageType string

const (
	ClientMessageTypeJoin             ClientMessageType = "join"
	ClientMessageTypeToggleCastMember ClientMessageType = "toggleCastMember"
	ClientMessageTypeSetDirector      ClientMessageType = "setDirector"
	ClientMessageTypeSetAudience      ClientMessageType = "setAudience"
	ClientMessageTypeSelectPlay       ClientMessageType = "selectPlay"
	ClientMessageTypeSetReady         ClientMessageType = "setReady"
	ClientMessageTypeSendChat         ClientMessageType = "sendChat"
	ClientMessageTypeStartPlay        ClientMessageType = "startPlay"
	ClientMessageTypeNextLine         ClientMessageType = "nextLine"
	ClientMessageTypeResetPlay        ClientMessageType = "resetPlay"
	ClientMessageTypeHeartbeat        ClientMessageType = "heartbeat"
)

// ServerMessageType represents the type of server message.
type ServerMessageType string

const (
	ServerMessageTypeStageState         ServerMessageType = "stageState"
	ServerMessageTypeJoined             ServerMessageType = "joined"
	ServerMessageTypeError              ServerMessageType = "error"
	ServerMessageTypeAvailablePlays     ServerMessageType = "availablePlays"
	ServerMessageTypeCountdownStart     ServerMessageType = "countdownStart"
	ServerMessageTypeCastMembersUpdated ServerMessageType = "castMembersUpdated"
	ServerMessageTypeChatMessage        ServerMessageType = "chatMessage"
	ServerMessageTypePlayStarted        ServerMessageType = "playStarted"
	ServerMessageTypeLineAdvanced       ServerMessageType = "lineAdvanced"
	ServerMessageTypePlayEnded          ServerMessageType = "playEnded"
)

// ClientMessage represents a message from the client to the server.
type ClientMessage struct {
	Type            ClientMessageType `json:"type"`
	StageID         string            `json:"stageId,omitempty"`
	ParticipantName string            `json:"participantName,omitempty"`
	ParticipantID   string            `json:"participantId,omitempty"`
	CastMemberID    string            `json:"castMemberId,omitempty"`
	PlayID          string            `json:"playId,omitempty"`
	IsDirector      *bool             `json:"isDirector,omitempty"`
	IsReady         *bool             `json:"isReady,omitempty"`
	IsAudience      *bool             `json:"isAudience,omitempty"`
	Text            string            `json:"text,omitempty"`
	CastOnly        bool              `json:"castOnly,omitempty"`
}

// ServerMessage represents a message from the server to the client.
type ServerMessage struct {
	State            *StageState               `json:"state,omitempty"`
	StageState       *StageState               `json:"stageState,omitempty"`
	Message          *ChatMessage              `json:"message,omitempty"`
	NextPlay         *playscripts.PlayScript   `json:"nextPlay,omitempty"`
	Type             ServerMessageType         `json:"type"`
	ParticipantID    string                    `json:"participantId,omitempty"`
	Error            string                    `json:"error,omitempty"`
	Plays            []*playscripts.PlayScript `json:"plays,omitempty"`
	Participants     []*StageParticipant       `json:"participants,omitempty"`
	CurrentLineIndex int                       `json:"currentLineIndex,omitempty"`
	Duration         int                       `json:"duration,omitempty"`
}

// StageState represents the current state of a stage for client consumption.
type StageState struct {
	SelectedPlay     *playscripts.PlayScript `json:"selectedPlay,omitempty"`
	ID               string                  `json:"id"`
	Code             string                  `json:"code,omitempty"`
	Phase            StagePhase              `json:"phase"`
	Participants     []*StageParticipant     `json:"participants"`
	CurrentLineIndex int                     `json:"currentLineIndex"`
}

// ChatMessage represents a chat message in a stage.
type ChatMessage struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderId"`
	SenderName string `json:"senderName"`
	Text       string `json:"text"`
	CastOnly   bool   `json:"castOnly"`
	Timestamp  int64  `json:"timestamp"`
}
