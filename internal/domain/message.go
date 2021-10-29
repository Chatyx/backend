package domain

import (
	"time"
)

const (
	MessageSendAction = iota + 1
	MessageJoinAction
	MessageLeaveAction
	MessageBlockAction
)

type CreateMessageDTO struct {
	ActionID int    `json:"-"`
	Text     string `json:"text"    validate:"required,max=4096"`
	ChatID   string `json:"chat_id" validate:"required,uuid4"`
	SenderID string `json:"-"`
}

type Message struct {
	ID        string     `json:"id"`
	ActionID  int        `json:"action_id"`
	Text      string     `json:"text"`
	ChatID    string     `json:"chat_id"`
	SenderID  string     `json:"sender_id"`
	CreatedAt *time.Time `json:"created_at"`
}
