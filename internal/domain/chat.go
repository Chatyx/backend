package domain

import "time"

type CreateChatDTO struct {
	Name        string `json:"name"        validate:"required"`
	Description string `json:"description"`
	CreatorID   string `json:"creator_id"`
}

type UpdateChatDTO struct {
	ID          string
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatorID   string
}

type Chat struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatorID   string     `json:"creator_id"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
