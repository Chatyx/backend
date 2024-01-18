package entity

import "time"

type User struct {
	ID        int
	Username  string
	PwdHash   string
	Email     string
	FirstName string
	LastName  string
	BirthDate time.Time
	Bio       string
	IsActive  bool
	IsDeleted bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
