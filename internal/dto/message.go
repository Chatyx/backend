package dto

import "github.com/Chatyx/backend/internal/entity"

type Sort string

func (s Sort) String() string {
	return string(s)
}

const (
	AscSort  Sort = "asc"
	DescSort Sort = "desc"
)

type MessageList struct {
	ChatID  entity.ChatID
	IDAfter int
	Limit   int
	Sort    Sort
}

type MessageCreate struct {
	ChatID      entity.ChatID
	Content     string
	ContentType string
}
