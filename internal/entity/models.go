package entity

import "time"

type GroupParticipantStatus string

func (gps GroupParticipantStatus) String() string {
	return string(gps)
}

const (
	JoinedStatus GroupParticipantStatus = "joined"
	KickedStatus GroupParticipantStatus = "kicked"
	LeftStatus   GroupParticipantStatus = "left"
)

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

type GroupParticipant struct {
	GroupID int
	UserID  int
	IsAdmin bool
	Status  GroupParticipantStatus
}

func (p GroupParticipant) IsInGroup() bool {
	return p.Status == JoinedStatus
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
