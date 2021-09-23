package domain

import (
	"encoding/json"
	"io"
	"time"
)

type CreateUserDTO struct {
	Username   string     `json:"username"   validate:"required,max=50"`
	Password   string     `json:"password"   validate:"required,min=8,max=27"`
	Email      string     `json:"email"      validate:"required,email,max=255"`
	FirstName  string     `json:"first_name" validate:"max=50"`
	LastName   string     `json:"last_name"  validate:"max=50"`
	BirthDate  *time.Time `json:"birth_date"`
	Department string     `json:"department" validate:"max=255"`
}

func (c *CreateUserDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, c)
}

func (c *CreateUserDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

type UpdateUserDTO struct {
	ID         string     `json:"id"         validate:"required"`
	Username   string     `json:"username"   validate:"omitempty,max=50"`
	Password   string     `json:"password"   validate:"omitempty,min=8,max=27"`
	Email      string     `json:"email"      validate:"omitempty,email,max=255"`
	FirstName  string     `json:"first_name" validate:"omitempty,max=50"`
	LastName   string     `json:"last_name"  validate:"omitempty,max=50"`
	BirthDate  *time.Time `json:"birth_date"`
	Department string     `json:"department" validate:"omitempty,max=255"`
}

func (c *UpdateUserDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, c)
}

func (c *UpdateUserDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

type User struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Password   string     `json:"-"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name,omitempty"`
	LastName   string     `json:"last_name,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Department string     `json:"department,omitempty"`
	IsDeleted  bool       `json:"-"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

func (u *User) Encode() ([]byte, error) {
	return json.Marshal(u)
}
