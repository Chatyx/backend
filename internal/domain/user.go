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

func (d *CreateUserDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, d)
}

func (d *CreateUserDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
}

type UpdateUserDTO struct {
	ID         string `json:"-"`
	Username   string `json:"username"   validate:"required,max=50"`
	Email      string `json:"email"      validate:"required,email,max=255"`
	FirstName  string `json:"first_name" validate:"max=50"`
	LastName   string `json:"last_name"  validate:"max=50"`
	BirthDate  string `json:"birth_date" validate:"sql-date"`
	Department string `json:"department" validate:"max=255"`
}

func (d *UpdateUserDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, d)
}

func (d *UpdateUserDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
}

type UpdateUserPasswordDTO struct {
	UserID  string `json:"-"`
	New     string `json:"new_password"     validate:"required,min=8,max=27"`
	Current string `json:"current_password" validate:"required,min=8,max=27"`
}

func (d *UpdateUserPasswordDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, d)
}

func (d *UpdateUserPasswordDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
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
