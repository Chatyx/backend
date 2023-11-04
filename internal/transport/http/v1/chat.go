package v1

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	chatListPath   = "/chats"
	chatDetailPath = "/chats/:chat_id"
)

type ChatDetail struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatList struct {
	Total int          `json:"total"`
	Data  []ChatDetail `json:"data"`
}

type ChatCreate struct {
	ContactID int `json:"contact_id" validate:"required"`
}

type ChatController struct {
}

func (cc *ChatController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodGet, chatListPath, cc.list)
	mux.HandlerFunc(http.MethodGet, chatDetailPath, cc.detail)
	mux.HandlerFunc(http.MethodPost, chatListPath, cc.create)
}

// list lists all one-on-one chats where user is a member
//
//	@Summary	List all one-on-one chats where user is a member
//	@Tags		chats
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	ChatList
//	@Failure	500	{object}	httputil.Error
//	@Router		/chats  [get]
func (cc *ChatController) list(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// detail gets a specified one-on-one chat
//
//	@Summary	Get a specified one-on-one chat
//	@Tags		chats
//	@Accept		json
//	@Produce	json
//	@Param		chat_id	path		string	true	"Chat identity"
//	@Success	200		{object}	ChatDetail
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/chats/{chat_id}  [get]
func (cc *ChatController) detail(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// create creates a one-on-one chat with a specified member
//
//	@Summary	Create a one-on-one chat with a specified member
//	@Tags		chats
//	@Accept		json
//	@Produce	json
//	@Param		input	body		ChatCreate	true	"Body to create"
//	@Success	201		{object}	ChatDetail
//	@Failure	400		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/chats  [post]
func (cc *ChatController) create(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}
