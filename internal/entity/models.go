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
}

type Group struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

type DialogPartner struct {
	UserID    int
	IsBlocked bool
}

type Dialog struct {
	ID        int
	IsBlocked bool
	Partner   DialogPartner
	CreatedAt time.Time
}
