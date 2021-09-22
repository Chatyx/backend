package domain

import (
	"encoding/json"
	"io"
	"time"
)

type CreateUserDTO struct {
	Username   string `json:"username"   validate:"required,max=50"`
	Password   string `json:"password"   validate:"required,min=8,max=27"`
	Email      string `json:"email"      validate:"required,email,max=255"`
	FirstName  string `json:"first_name" validate:"max=50"`
	LastName   string `json:"last_name"  validate:"max=50"`
	BirthDate  string `json:"birth_date" validate:"sql-date"`
	Department string `json:"department" validate:"max=255"`
}

func (c *CreateUserDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, c)
}

func (c *CreateUserDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

type UpdateUserDTO struct {
	CreateUserDTO

	ID string `json:"id" validate:"required"`
}

type User struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Password   string     `json:"-"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name,omitempty"`
	LastName   string     `json:"last_name,omitempty"`
	BirthDate  string     `json:"birth_date,omitempty"`
	Department string     `json:"department,omitempty"`
	IsDeleted  bool       `json:"-"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

func (u *User) Encode() ([]byte, error) {
	return json.Marshal(u)
}
