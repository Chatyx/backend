package encoding

import (
	"encoding/json"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type jsonCreateUserDTOUnmarshaler struct {
	dto *domain.CreateUserDTO
}

func NewJSONCreateUserDTOUnmarshaler(dto *domain.CreateUserDTO) Unmarshaler {
	return &jsonCreateUserDTOUnmarshaler{dto: dto}
}

func (um *jsonCreateUserDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonUpdateUserDTOUnmarshaler struct {
	dto *domain.UpdateUserDTO
}

func NewJSONUpdateUserDTOUnmarshaler(dto *domain.UpdateUserDTO) Unmarshaler {
	return &jsonUpdateUserDTOUnmarshaler{dto: dto}
}

func (um *jsonUpdateUserDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonUpdateUserPasswordDTOUnmarshaler struct {
	dto *domain.UpdateUserPasswordDTO
}

func NewJSONUpdateUserPasswordDTOUnmarshaler(dto *domain.UpdateUserPasswordDTO) Unmarshaler {
	return &jsonUpdateUserPasswordDTOUnmarshaler{dto: dto}
}

func (um *jsonUpdateUserPasswordDTOUnmarshaler) Unmarshal(payload []byte) error {
	return json.Unmarshal(payload, um.dto)
}

type jsonUserMarshaler struct {
	user domain.User
}

func NewJSONUserMarshaler(user domain.User) Marshaler {
	return jsonUserMarshaler{user: user}
}

func (m jsonUserMarshaler) Marshal() ([]byte, error) {
	return json.Marshal(m.user)
}
