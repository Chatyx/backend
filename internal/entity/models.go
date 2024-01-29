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

type ChatType string

func (ct ChatType) String() string {
	return string(ct)
}

const (
	DialogChatType ChatType = "dialog"
	GroupChatType  ChatType = "group"
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

type ChatID struct {
	ID   int
	Type ChatType
}

type Message struct {
	ID          int
	ChatID      ChatID
	SenderID    int
	Content     []byte
	ContentType string
	IsService   bool
	SentAt      time.Time
	DeliveredAt *time.Time
}
