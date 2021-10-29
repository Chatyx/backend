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
	router.Handler(http.MethodGet, listChatMembersURI, authMid(http.HandlerFunc(h.listMembers)))
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
func (h *chatMemberHandler) listMembers(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	userID := domain.UserIDFromContext(ctx)

	members, err := h.chatMemberService.ListMembersWhoBelongToChat(ctx, chatID, userID)
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
