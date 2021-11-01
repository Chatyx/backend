package domain

const (
	InChat = iota + 1
	Left
	Kicked
)

type UpdateChatMemberDTO struct {
	UserID   string `json:"-"`
	ChatID   string `json:"-"`
	StatusID int    `json:"status_id" validate:"required,oneof=1 2 3"`
}

type ChatMember struct {
	Username  string `json:"username"`
	IsCreator bool   `json:"is_creator"`
	StatusID  int    `json:"-"`
	UserID    string `json:"user_id"`
	ChatID    string `json:"chat_id"`
}
