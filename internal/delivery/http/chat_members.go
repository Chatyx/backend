package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/julienschmidt/httprouter"
)

const (
	currentChatMemberURI = "/api/chats/:chat_id/member"
	listChatMembersURI   = "/api/chats/:chat_id/members"
	detailChatMemberURI  = "/api/chats/:chat_id/members/:user_id"
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
}

func newChatMemberHandler(cms service.ChatMemberService) *chatMemberHandler {
	return &chatMemberHandler{
		baseHandler:       &baseHandler{logger: logging.GetLogger()},
		chatMemberService: cms,
	}
}

func (h *chatMemberHandler) register(router *httprouter.Router, authMid Middleware) {
	router.Handler(http.MethodGet, listChatMembersURI, authMid(http.HandlerFunc(h.list)))
	router.Handler(http.MethodPost, listChatMembersURI, authMid(http.HandlerFunc(h.join)))
	router.Handler(http.MethodGet, currentChatMemberURI, authMid(http.HandlerFunc(h.detailCurrent)))
	router.Handler(http.MethodGet, detailChatMemberURI, authMid(http.HandlerFunc(h.detail)))
	router.Handler(http.MethodPatch, currentChatMemberURI, authMid(http.HandlerFunc(h.updateStatus)))
	router.Handler(http.MethodPatch, detailChatMemberURI, authMid(http.HandlerFunc(h.updateStatusByCreator)))
}

// @Summary Get list of chat members
// @Tags Chat Members
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Success 200 {object} MemberListResponse
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/members [get]
func (h *chatMemberHandler) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	chatID := httprouter.ParamsFromContext(ctx).ByName(chatIDParam)

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{"chat_id": chatID})
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.UUIDValidator(chatIDParam, chatID)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: authUser.UserID,
		ChatID: chatID,
	}

	members, err := h.chatMemberService.List(ctx, memberKey)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondErrorRefactored(ctx, w, errChatNotFound.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
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
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)
	chatID, userID := ps.ByName(chatIDParam), req.URL.Query().Get(userIDParam)

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		"chat_id": chatID,
		"user_id": userID,
	})
	ctx = logging.NewContextFromLogger(ctx, logger)

	vl := validator.ChainValidator(
		validator.UUIDValidator(chatIDParam, chatID),
		validator.UUIDValidator(userIDParam, userID),
	)

	if err := h.validate(vl); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: userID,
		ChatID: chatID,
	}

	if err := h.chatMemberService.JoinToChat(ctx, memberKey, authUser); err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondErrorRefactored(ctx, w, errChatNotFound.Wrap(err))
		case errors.Is(err, domain.ErrUserNotFound):
			respondErrorRefactored(ctx, w, errUserNotFound.Wrap(err))
		case errors.Is(err, domain.ErrChatMemberUniqueViolation):
			respondErrorRefactored(ctx, w, errChatMemberUniqueViolation.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}

// @Summary Get chat member info
// @Tags Chat Members
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Param user_id path string true "User id"
// @Success 200 {object} domain.ChatMember
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/members/{user_id} [get]
func (h *chatMemberHandler) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)
	chatID, userID := ps.ByName(chatIDParam), ps.ByName(userIDParam)

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		"chat_id": chatID,
		"user_id": userID,
	})
	ctx = logging.NewContextFromLogger(ctx, logger)

	vl := validator.ChainValidator(
		validator.UUIDValidator(chatIDParam, chatID),
		validator.UUIDValidator(userIDParam, userID),
	)

	if err := h.validate(vl); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: userID,
		ChatID: chatID,
	}

	member, err := h.chatMemberService.GetByKey(ctx, memberKey, authUser)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondErrorRefactored(ctx, w, errChatNotFound.Wrap(err))
		case errors.Is(err, domain.ErrChatMemberNotFound):
			respondErrorRefactored(ctx, w, errChatMemberNotFound.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusOK, w, encoding.NewJSONChatMemberMarshaler(member))
}

// @Summary Get current authenticated chat member info
// @Tags Chat Members
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Success 200 {object} domain.ChatMember
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/member [get]
func (h *chatMemberHandler) detailCurrent(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	chatID := httprouter.ParamsFromContext(ctx).ByName(chatIDParam)

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{"chat_id": chatID})
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.UUIDValidator(chatIDParam, chatID)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: authUser.UserID,
		ChatID: chatID,
	}

	member, err := h.chatMemberService.GetByKey(ctx, memberKey, authUser)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondErrorRefactored(ctx, w, errChatNotFound.Wrap(err))
		case errors.Is(err, domain.ErrChatMemberNotFound):
			respondErrorRefactored(ctx, w, errChatMemberNotFound.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusOK, w, encoding.NewJSONChatMemberMarshaler(member))
}

// @Summary Update current authenticated member status in chat
// @Description Use this endpoint to leave current member from chat (status id=2) or come back to chat (status id=1)
// @Tags Chat Members
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Param input body domain.UpdateChatMemberDTO true "Update body"
// @Success 204 "No Content"
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/member [patch]
func (h *chatMemberHandler) updateStatus(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	dto := domain.UpdateChatMemberDTO{}
	chatID := httprouter.ParamsFromContext(ctx).ByName(chatIDParam)

	if err := h.decodeBody(req.Body, encoding.NewJSONUpdateChaMemberDTOUnmarshaler(&dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		"chat_id":   chatID,
		"status_id": dto.StatusID,
	})
	ctx = logging.NewContextFromLogger(ctx, logger)

	vl := validator.ChainValidator(
		validator.StructValidator(dto),
		validator.UUIDValidator(chatIDParam, chatID),
	)

	if err := h.validate(vl); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	dto.ChatID = chatID
	dto.UserID = authUser.UserID

	if err := h.chatMemberService.UpdateStatus(ctx, dto, authUser); err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondErrorRefactored(ctx, w, errChatNotFound.Wrap(err))
		case errors.Is(err, domain.ErrChatMemberNotFound):
			respondErrorRefactored(ctx, w, errChatMemberNotFound.Wrap(err))
		case errors.Is(err, domain.ErrChatMemberWrongStatusTransit):
			respondErrorRefactored(ctx, w, errChatMemberWrongStatusTransit.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}

// @Summary Update member status in chat by creator
// @Description Use this endpoint to kick member from chat (status id=3) or come back him to chat (status id=1)
// @Tags Chat Members
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Param user_id path string true "User id"
// @Param input body domain.UpdateChatMemberDTO true "Update body"
// @Success 204 "No Content"
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/members/{user_id} [patch]
func (h *chatMemberHandler) updateStatusByCreator(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	dto := domain.UpdateChatMemberDTO{}
	ps := httprouter.ParamsFromContext(ctx)
	chatID, userID := ps.ByName(chatIDParam), ps.ByName(userIDParam)

	if err := h.decodeBody(req.Body, encoding.NewJSONUpdateChaMemberDTOUnmarshaler(&dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		"chat_id":   chatID,
		"user_id":   userID,
		"status_id": dto.StatusID,
	})
	ctx = logging.NewContextFromLogger(ctx, logger)

	vl := validator.ChainValidator(
		validator.StructValidator(dto),
		validator.UUIDValidator(chatIDParam, chatID),
		validator.UUIDValidator(userIDParam, userID),
	)

	if err := h.validate(vl); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	dto.ChatID = chatID
	dto.UserID = userID

	if err := h.chatMemberService.UpdateStatus(ctx, dto, authUser); err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondErrorRefactored(ctx, w, errChatNotFound.Wrap(err))
		case errors.Is(err, domain.ErrChatMemberNotFound):
			respondErrorRefactored(ctx, w, errChatMemberNotFound.Wrap(err))
		case errors.Is(err, domain.ErrChatMemberWrongStatusTransit):
			respondErrorRefactored(ctx, w, errChatMemberWrongStatusTransit.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}
