package ws

import (
	"encoding/json"
	"time"
)

const (
	EventTaskCreated   = "task:created"
	EventTaskUpdated   = "task:updated"
	EventTaskDeleted   = "task:deleted"
	EventTaskClaimed   = "task:claimed"
	EventTaskCompleted = "task:completed"

	EventPing = "ping"
	EventPong = "pong"
)

// унифицированное сообщение для вебсокета
type Message struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp int64           `json:"ts"`
}

func NewMessage(msgType string, payload any) (Message, error) {
	var raw json.RawMessage

	if b, ok := payload.([]byte); ok {
		raw = b
	} else {
		if payload != nil {
			b, err := json.Marshal(payload)
			if err != nil {
				return Message{}, err
			}
			raw = b
		}
	}

	return Message{
		Type:      msgType,
		Payload:   raw,
		Timestamp: TimestampNow(),
	}, nil
}

func TimestampNow() int64 {
	return time.Now().Unix()
}
