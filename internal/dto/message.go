package dto

import "github.com/Chatyx/backend/internal/entity"

type Direction string

const (
	AscDirection  Direction = "asc"
	DescDirection Direction = "desc"
)

type MessageList struct {
	ChatID    entity.ChatID
	IDAfter   int
	Limit     int
	Direction Direction
}

type MessageCreate struct {
	ChatID      entity.ChatID
	Content     []byte
	ContentType string
}
