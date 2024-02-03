package dto

import "time"

type UserCreate struct {
	Username  string
	Password  string
	Email     string
	FirstName string
	LastName  string
	BirthDate *time.Time
	Bio       string
}

type UserUpdate struct {
	ID        int
	Username  string
	Email     string
	FirstName string
	LastName  string
	BirthDate *time.Time
	Bio       string
}

type UserUpdatePassword struct {
	UserID      int
	CurPassword string
	NewPassword string
}
