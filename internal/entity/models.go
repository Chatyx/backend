package entity

import "time"

type User struct {
	ID        int
	Username  string
	PwdHash   string
	Email     string
	FirstName string
	LastName  string
	BirthDate *time.Time
	Bio       string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
