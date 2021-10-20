package encoding

import (
	"encoding/json"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type jsonSignInDTOUnmarshaler struct {
	dto *domain.SignInDTO
}

func NewJSONSignInDTOUnmarshaler(dto *domain.SignInDTO) Unmarshaler {
	return &jsonSignInDTOUnmarshaler{dto: dto}
}

func (um *jsonSignInDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonRefreshSessionDTOUnmarshaler struct {
	dto *domain.RefreshSessionDTO
}

func NewJSONRefreshSessionDTOUnmarshaler(dto *domain.RefreshSessionDTO) Unmarshaler {
	return &jsonRefreshSessionDTOUnmarshaler{dto: dto}
}

func (um *jsonRefreshSessionDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonTokenPairMarshaler struct {
	pair domain.JWTPair
}

func NewJSONTokenPairMarshaler(pair domain.JWTPair) Marshaler {
	return jsonTokenPairMarshaler{pair: pair}
}

func (m jsonTokenPairMarshaler) Marshal() ([]byte, error) {
	return json.Marshal(m.pair)
}
