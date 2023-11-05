package v1

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	messageListPath = "/messages"
)

type Message struct {
	ID          int        `json:"id"`
	ChatID      int        `json:"chat_id"`
	ChatType    string     `json:"chat_type"`
	SenderID    int        `json:"sender_id"`
	Content     []byte     `json:"content"`
	ContentType string     `json:"content_type"`
	IsService   bool       `json:"is_service"`
	SentAt      time.Time  `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	SeenAt      *time.Time `json:"seen_at"`
}

type MessageList struct {
	Total int       `json:"total"`
	Data  []Message `json:"data"`
}

type MessageCreate struct {
	ChatID      int    `json:"chat_id"      validate:"required"`
	ChatType    string `json:"chat_type"    validate:"required,oneof=conversation group"`
	Content     []byte `json:"content"      validate:"required,max=100000"`
	ContentType string `json:"content_type" validate:"required,oneof=text file"`
}

type MessageController struct {
}

func (mc *MessageController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodGet, messageListPath, mc.list)
	mux.HandlerFunc(http.MethodPost, messageListPath, mc.create)
}

// list lists messages for a specified chat
//
//	@Summary	List messages for a specified chat
//	@Tags		messages
//	@Accept		json
//	@Produce	json
//	@Param		chat_id		query		int		true	"Chat identity"
//	@Param		chat_type	query		string	true	"Chat type (conversation, group)"
//	@Success	200			{object}	MessageList
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Router		/messages  [get]
func (mc *MessageController) list(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// create sends message to the specified chat
//
//	@Summary	Send message to the specified chat
//	@Tags		messages
//	@Accept		json
//	@Produce	json
//	@Param		input	body		MessageCreate	true	"Body to create"
//	@Success	201		{object}	Message
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/messages  [post]
func (mc *MessageController) create(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}
