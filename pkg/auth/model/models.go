package model

import (
	"net"
	"time"
)

type Credentials struct {
	Username    string
	Password    string
	Fingerprint string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type RefreshSession struct {
	RefreshToken string
	Fingerprint  string
}

type Session struct {
	UserID       string
	RefreshToken string
	Fingerprint  string
	IP           net.IP
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
