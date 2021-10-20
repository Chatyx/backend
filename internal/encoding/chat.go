package encoding

import (
	"encoding/json"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type jsonCreateChatDTOUnmarshaler struct {
	dto *domain.CreateChatDTO
}

func NewJSONCreateChatDTOUnmarshaler(dto *domain.CreateChatDTO) Unmarshaler {
	return &jsonCreateChatDTOUnmarshaler{dto: dto}
}

func (um *jsonCreateChatDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonUpdateChatDTOUnmarshaler struct {
	dto *domain.UpdateChatDTO
}

func NewJSONUpdateChatDTOUnmarshaler(dto *domain.UpdateChatDTO) Unmarshaler {
	return &jsonUpdateChatDTOUnmarshaler{dto: dto}
}

func (um *jsonUpdateChatDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonChatMarshaler struct {
	chat domain.Chat
}

func NewJSONChatMarshaler(chat domain.Chat) Marshaler {
	return jsonChatMarshaler{chat: chat}
}

func (m jsonChatMarshaler) Marshal() ([]byte, error) {
	return json.Marshal(m.chat)
}
