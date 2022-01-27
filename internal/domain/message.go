package domain

import (
	"time"
)

const (
	MessageSendAction = iota + 1
	MessageJoinAction
	MessageLeaveAction
	MessageKickAction
)

const (
	NewerMessages = "newer"
	OlderMessages = "older"
)

const defaultListMessages = 20

type MessageListDTO struct {
	OffsetDate time.Time
	Direction  string `json:"direction" validate:"oneof=newer older"`
	Limit      int    `json:"limit"     validate:"gte=0,lte=100"`
	Offset     int    `json:"offset"    validate:"gte=0"`
}

func NewMessageListDTO(offsetDate time.Time, direction string, limit, offset int) MessageListDTO {
	dto := MessageListDTO{
		OffsetDate: offsetDate,
		Direction:  direction,
		Limit:      limit,
		Offset:     offset,
	}

	if dto.Direction == "" {
		dto.Direction = NewerMessages
	}

	if dto.Limit == 0 {
		dto.Limit = defaultListMessages
	}

	return dto
}

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

type MessageList struct {
	Total    int
	Messages []Message
}
