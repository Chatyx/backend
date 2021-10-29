package http

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

const (
	listChatUsersURI = "/api/chats/:chat_id/users"
)

type userChatHandler struct {
	*baseHandler
	userChatService service.UserChatService
	logger          logging.Logger
}

func newUserChatHandler(ucs service.UserChatService, validate *validator.Validate) *userChatHandler {
	logger := logging.GetLogger()

	return &userChatHandler{
		baseHandler: &baseHandler{
			logger:   logger,
			validate: validate,
		},
		userChatService: ucs,
		logger:          logger,
	}
}

func (h *userChatHandler) register(router *httprouter.Router, authMid Middleware) {
	router.Handler(http.MethodGet, listChatUsersURI, authMid(http.HandlerFunc(h.listMembers)))
}

// @Summary Get list of users who belong to chat
// @Tags Users_Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Success 200 {object} UserListResponse
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/users [get]
func (h *userChatHandler) listMembers(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	userID := domain.UserIDFromContext(ctx)

	users, err := h.userChatService.ListUsersWhoBelongToChat(ctx, chatID, userID)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, err)
		}

		return
	}

	respondSuccess(http.StatusOK, w, UserListResponse{List: users})
}
