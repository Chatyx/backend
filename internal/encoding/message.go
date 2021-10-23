package encoding

import (
	"encoding/json"

	"github.com/Mort4lis/scht-backend/internal/domain"
	protogen "github.com/Mort4lis/scht-backend/internal/encoding/proto-gen"
	prototypes "github.com/gogo/protobuf/types"
)

type jsonCreateMessageDTOUnmarshaler struct {
	dto *domain.CreateMessageDTO
}

func NewJSONCreateMessageDTOUnmarshaler(dto *domain.CreateMessageDTO) Unmarshaler {
	return &jsonCreateMessageDTOUnmarshaler{dto: dto}
}

func (um *jsonCreateMessageDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonMessageMarshaler struct {
	message domain.Message
}

func NewJSONMessageMarshaler(message domain.Message) Marshaler {
	return jsonMessageMarshaler{message: message}
}

func (m jsonMessageMarshaler) Marshal() ([]byte, error) {
	return json.Marshal(m.message)
}

//go:generate protoc --gofast_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types:. ./proto/message.proto

type protobufCreateDTOMessageMarshaler struct {
	dto domain.CreateMessageDTO
}

func NewProtobufCreateDTOMessageMarshaler(dto domain.CreateMessageDTO) Marshaler {
	return protobufCreateDTOMessageMarshaler{dto: dto}
}

func (m protobufCreateDTOMessageMarshaler) Marshal() ([]byte, error) {
	dto := &protogen.CreateMessageDTO{
		Text:   m.dto.Text,
		ChatId: m.dto.ChatID,
	}

	return dto.Marshal()
}

type protobufCreateDTOMessageUnmarshaler struct {
	dto *domain.CreateMessageDTO
}

func NewProtobufCreateDTOMessageUnmarshaler(dto *domain.CreateMessageDTO) Unmarshaler {
	return &protobufCreateDTOMessageUnmarshaler{dto: dto}
}

func (p *protobufCreateDTOMessageUnmarshaler) Unmarshal(payload []byte) error {
	dto := new(protogen.CreateMessageDTO)
	if err := dto.Unmarshal(payload); err != nil {
		return err
	}

	p.dto.Text = dto.Text
	p.dto.ChatID = dto.ChatId

	return nil
}

type protobufMessageMarshaler struct {
	message domain.Message
}

func NewProtobufMessageMarshaler(message domain.Message) Marshaler {
	return protobufMessageMarshaler{message: message}
}

func (m protobufMessageMarshaler) Marshal() ([]byte, error) {
	createdAt, err := prototypes.TimestampProto(*m.message.CreatedAt)
	if err != nil {
		return nil, err
	}

	msg := &protogen.Message{
		Id:        m.message.ID,
		Action:    protogen.Message_Action(m.message.Action),
		Text:      m.message.Text,
		ChatId:    m.message.ChatID,
		SenderId:  m.message.SenderID,
		CreatedAt: createdAt,
	}

	return msg.Marshal()
}

type protobufMessageUnmarshaler struct {
	message *domain.Message
}

func NewProtobufMessageUnmarshaler(message *domain.Message) Unmarshaler {
	return &protobufMessageUnmarshaler{message: message}
}

func (um *protobufMessageUnmarshaler) Unmarshal(payload []byte) error {
	msg := new(protogen.Message)
	if err := msg.Unmarshal(payload); err != nil {
		return err
	}

	um.message.ID = msg.Id
	um.message.Action = int(msg.Action)
	um.message.Text = msg.Text
	um.message.ChatID = msg.ChatId
	um.message.SenderID = msg.SenderId

	createdAt, err := prototypes.TimestampFromProto(msg.CreatedAt)
	if err != nil {
		return err
	}

	createdAt = createdAt.Local()
	um.message.CreatedAt = &createdAt

	return nil
}
