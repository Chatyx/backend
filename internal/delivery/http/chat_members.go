package http

import (
	"encoding/json"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

const (
	listChatMembersURI = "/api/chats/:chat_id/members"
)

type MemberListResponse struct {
	List []domain.ChatMember `json:"list"`
}

func (r MemberListResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type chatMemberHandler struct {
	*baseHandler
	chatMemberService service.ChatMemberService
	logger            logging.Logger
}

func newChatMemberHandler(cms service.ChatMemberService, validate *validator.Validate) *chatMemberHandler {
	logger := logging.GetLogger()

	return &chatMemberHandler{
		baseHandler: &baseHandler{
			logger:   logger,
			validate: validate,
		},
		chatMemberService: cms,
		logger:            logger,
	}
}

func (h *chatMemberHandler) register(router *httprouter.Router, authMid Middleware) {
	router.Handler(http.MethodGet, listChatMembersURI, authMid(http.HandlerFunc(h.list)))
	router.Handler(http.MethodPost, listChatMembersURI, authMid(http.HandlerFunc(h.join)))
}

// @Summary Get list of chat members
// @Tags Chat Members
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Success 200 {object} MemberListResponse
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/members [get]
func (h *chatMemberHandler) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	userID := domain.UserIDFromContext(ctx)

	members, err := h.chatMemberService.ListMembersInChat(ctx, chatID, userID)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, err)
		}

		return
	}

	respondSuccess(http.StatusOK, w, MemberListResponse{List: members})
}

// @Summary Join member to the chat
// @Tags Chat Members
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Param user_id query string true "User id"
// @Success 204 "No Content"
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/members [post]
func (h *chatMemberHandler) join(w http.ResponseWriter, req *http.Request) {
	userID := req.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "validation error",
			Fields:     ErrorFields{"user_id": "must be required"},
		})

		return
	}

	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	creatorID := domain.UserIDFromContext(ctx)

	if err := h.chatMemberService.JoinMemberToChat(ctx, chatID, creatorID, userID); err != nil {
		switch err {
		case domain.ErrChatNotFound, domain.ErrUserNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		case domain.ErrChatMemberUniqueViolation:
			respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: err.Error()})
		default:
			respondError(w, err)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}