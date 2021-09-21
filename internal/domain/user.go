package domain

import "time"

type User struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"-"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	BirthDate  time.Time `json:"birth_date,omitempty"`
	Department string    `json:"department,omitempty"`
	IsDeleted  bool      `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}
