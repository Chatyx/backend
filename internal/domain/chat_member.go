package domain

const (
	InChat = iota + 1
	Left
	Kicked
)

type ChatMember struct {
	Username string `json:"username"`
	StatusID int    `json:"-"`
	UserID   string `json:"user_id"`
	ChatID   string `json:"chat_id"`
}
