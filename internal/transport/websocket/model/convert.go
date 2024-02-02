package model

import (
	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewMessageFromEntity(message entity.Message) *Message {
	var deliveredAt *timestamppb.Timestamp
	if message.DeliveredAt != nil {
		deliveredAt = timestamppb.New(*message.DeliveredAt)
	}

	var chatType ChatType
	switch message.ChatID.Type {
	case entity.DialogChatType:
		chatType = ChatType_DIALOG
	case entity.GroupChatType:
		chatType = ChatType_GROUP
	}

	var contentType ContentType
	switch message.ContentType {
	case entity.TextContentType:
		contentType = ContentType_TEXT
	case entity.ImageContentType:
		contentType = ContentType_IMAGE
	}

	return &Message{
		Id:          int64(message.ID),
		ChatId:      int64(message.ChatID.ID),
		ChatType:    chatType,
		SenderId:    int64(message.SenderID),
		Content:     message.Content,
		ContentType: contentType,
		IsService:   message.IsService,
		SentAt:      timestamppb.New(message.SentAt),
		Delivered:   deliveredAt,
	}
}

func (x *MessageCreate) DTO() dto.MessageCreate {
	var chatType entity.ChatType

	switch x.ChatType {
	case ChatType_DIALOG:
		chatType = entity.DialogChatType
	case ChatType_GROUP:
		chatType = entity.GroupChatType
	}

	return dto.MessageCreate{
		ChatID: entity.ChatID{
			ID:   int(x.ChatId),
			Type: chatType,
		},
		Content:     x.Content,
		ContentType: entity.TextContentType,
	}
}
