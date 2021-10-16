package domain

import (
	"encoding/json"
	"io"
	"time"
)

const (
	MessageSendAction  = "send"
	MessageJoinAction  = "join"
	MessageLeaveAction = "leave"
	MessageBlockAction = "block"
)

type Message struct {
	Action    string     `json:"action"`
	Message   string     `json:"text"`
	ChatID    string     `json:"chat_id"`
	SenderID  string     `json:"sender_id"`
	CreatedAt *time.Time `json:"created_at"`
}

func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Decode(payload []byte) error {
	return json.Unmarshal(payload, m)
}

func (m *Message) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(m)
}
