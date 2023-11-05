package v1

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	conversationListPath   = "/conversations"
	conversationDetailPath = "/conversations/:conversation_id"
)

type ConversationDetail struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationList struct {
	Total int                  `json:"total"`
	Data  []ConversationDetail `json:"data"`
}

type ConversationCreate struct {
	ContactID int `json:"contact_id" validate:"required"`
}

type ConversationController struct {
}

func (cc *ConversationController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodGet, conversationListPath, cc.list)
	mux.HandlerFunc(http.MethodGet, conversationDetailPath, cc.detail)
	mux.HandlerFunc(http.MethodPost, conversationListPath, cc.create)
}

// list lists all one-on-one conversations
//
//	@Summary	List all one-on-one conversations
//	@Tags		conversations
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	ConversationList
//	@Failure	500	{object}	httputil.Error
//	@Router		/conversations  [get]
func (cc *ConversationController) list(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// detail gets a specified one-on-one conversation
//
//	@Summary	Get a specified one-on-one conversation
//	@Tags		conversations
//	@Accept		json
//	@Produce	json
//	@Param		conversation_id	path		int	true	"Conversation identity"
//	@Success	200				{object}	ConversationDetail
//	@Failure	400				{object}	httputil.Error
//	@Failure	404				{object}	httputil.Error
//	@Failure	500				{object}	httputil.Error
//	@Router		/conversations/{conversation_id}  [get]
func (cc *ConversationController) detail(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// create creates a one-on-one conversation with a specified participant
//
//	@Summary	Create a one-on-one conversation with a specified participant
//	@Tags		conversations
//	@Accept		json
//	@Produce	json
//	@Param		input	body		ConversationCreate	true	"Body to create"
//	@Success	201		{object}	ConversationDetail
//	@Failure	400		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/conversations  [post]
func (cc *ConversationController) create(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}
