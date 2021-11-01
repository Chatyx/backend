package encoding

import (
	"encoding/json"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type jsonUpdateChatMemberDTOUnmarshaler struct {
	dto *domain.UpdateChatMemberDTO
}

func NewJSONUpdateChaMemberDTOUnmarshaler(dto *domain.UpdateChatMemberDTO) Unmarshaler {
	return &jsonUpdateChatMemberDTOUnmarshaler{dto: dto}
}

func (um *jsonUpdateChatMemberDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}
