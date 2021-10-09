package domain

import (
	"encoding/json"
	"io"
	"time"
)

type CreateChatDTO struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	CreatorID   string `json:"-"`
}

func (c *CreateChatDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, c)
}

func (c *CreateChatDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

type UpdateChatDTO struct {
	ID          string `json:"-"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	CreatorID   string `json:"-"`
}

func (c *UpdateChatDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, c)
}

func (c *UpdateChatDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

type Chat struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatorID   string     `json:"creator_id"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func (c *Chat) Encode() ([]byte, error) {
	return json.Marshal(c)
}
