package service

import "github.com/Mort4lis/scht-backend/internal/domain"

var defaultUser = domain.User{
	ID:        "1",
	Username:  "john1967",
	Password:  "8743b52063cd84097a65d1633f5c74f5",
	Email:     "john1967@gmail.com",
	CreatedAt: &currentTime,
}
