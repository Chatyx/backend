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
	StatusID  int    `json:"status_id"`
	UserID    string `json:"user_id"`
	ChatID    string `json:"chat_id"`
}

func (cm ChatMember) IsInChat() bool {
	return cm.StatusID == InChat
}

func (cm ChatMember) HasLeft() bool {
	return cm.StatusID == Left
}

func (cm ChatMember) HasKicked() bool {
	return cm.StatusID == Kicked
}
