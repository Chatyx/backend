package domain

import (
	"encoding/json"
	"io"
	"time"
)

const (
	MessageSendType  = "send"
	MessageJoinType  = "join"
	MessageLeaveType = "leave"
	MessageBlockType = "block"
)

type Message struct {
	Type      string `json:"type"`
	Message   string `json:"text"`
	ChatID    string `json:"chat_id"`
	SenderID  string
	CreatedAt *time.Time
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
