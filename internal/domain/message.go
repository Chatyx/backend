package domain

import (
	"encoding/json"
	"io"
	"time"
)

const (
	MessageSendAction = iota
	MessageJoinAction
	MessageLeaveAction
	MessageBlockAction
)

type CreateMessageDTO struct {
	Text   string `json:"text"    validate:"required,max=4096"`
	ChatID string `json:"chat_id" validate:"required,uuid4"`
}

func (d *CreateMessageDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, d)
}

func (d *CreateMessageDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
}

type Message struct {
	Action    int        `json:"action"`
	Text      string     `json:"text"`
	ChatID    string     `json:"chat_id"`
	SenderID  string     `json:"sender_id"`
	CreatedAt *time.Time `json:"created_at"`
}
