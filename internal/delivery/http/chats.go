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
	listChatURI   = "/api/chats"
	detailChatURI = "/api/chats/:chat_id"
)

const chatIDParam = "chat_id"

type ChatListResponse struct {
	List []domain.Chat `json:"list"`
}

func (r ChatListResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type chatHandler struct {
	*baseHandler
	chatService service.ChatService
}

func newChatHandler(chatService service.ChatService) *chatHandler {
	return &chatHandler{
		baseHandler: &baseHandler{logger: logging.GetLogger()},
		chatService: chatService,
	}
}

func (h *chatHandler) register(router *httprouter.Router, authMid Middleware) {
	router.Handler(http.MethodGet, listChatURI, authMid(http.HandlerFunc(h.list)))
	router.Handler(http.MethodPost, listChatURI, authMid(http.HandlerFunc(h.create)))
	router.Handler(http.MethodGet, detailChatURI, authMid(http.HandlerFunc(h.detail)))
	router.Handler(http.MethodPut, detailChatURI, authMid(http.HandlerFunc(h.update)))
	router.Handler(http.MethodDelete, detailChatURI, authMid(http.HandlerFunc(h.delete)))
}

// @Summary Get list of chats where user is a member
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Success 200 {object} ChatListResponse
// @Failure 500 {object} ResponseError
// @Router /chats [get]
func (h *chatHandler) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	authUser := domain.AuthUserFromContext(ctx)

	chats, err := h.chatService.List(ctx, authUser.UserID)
	if err != nil {
		respondError(ctx, w, err)
		return
	}

	respondSuccess(ctx, http.StatusOK, w, ChatListResponse{List: chats})
}

// @Summary Create chat
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param input body domain.CreateChatDTO true "Create body"
// @Success 201 {object} domain.Chat
// @Failure 400 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats [post]
func (h *chatHandler) create(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	dto := domain.CreateChatDTO{}

	if err := h.decodeBody(req.Body, encoding.NewJSONCreateChatDTOUnmarshaler(&dto)); err != nil {
		respondError(ctx, w, err)
		return
	}

	logFields := logging.Fields{}
	if dto.Name != "" {
		logFields["name"] = dto.Name
	}

	if dto.Description != "" {
		logFields["description"] = dto.Description
	}

	logger := logging.GetLoggerFromContext(ctx).WithFields(logFields)
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.StructValidator(dto)); err != nil {
		respondError(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	dto.CreatorID = authUser.UserID

	chat, err := h.chatService.Create(ctx, dto)
	if err != nil {
		respondError(ctx, w, err)
		return
	}

	respondSuccess(ctx, http.StatusCreated, w, encoding.NewJSONChatMarshaler(chat))
}

// @Summary Get chat by id where user is a member
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Success 200 {object} domain.Chat
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id} [get]
func (h *chatHandler) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	chatID := httprouter.ParamsFromContext(ctx).ByName(chatIDParam)
	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{"chat_id": chatID})
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.UUIDValidator(chatIDParam, chatID)); err != nil {
		respondError(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: authUser.UserID,
		ChatID: chatID,
	}

	chat, err := h.chatService.Get(ctx, memberKey)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondError(ctx, w, errChatNotFound.Wrap(err))
		default:
			respondError(ctx, w, err)
		}

		return
	}

	respondSuccess(ctx, http.StatusOK, w, encoding.NewJSONChatMarshaler(chat))
}

// @Summary Update chat where authenticated user is creator
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Param input body domain.UpdateChatDTO true "Update body"
// @Success 200 {object} domain.Chat
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id} [put]
func (h *chatHandler) update(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	dto := domain.UpdateChatDTO{}
	chatID := httprouter.ParamsFromContext(ctx).ByName(chatIDParam)

	if err := h.decodeBody(req.Body, encoding.NewJSONUpdateChatDTOUnmarshaler(&dto)); err != nil {
		respondError(ctx, w, err)
		return
	}

	logFields := logging.Fields{"chat_id": chatID}
	if dto.Name != "" {
		logFields["name"] = dto.Name
	}

	if dto.Description != "" {
		logFields["description"] = dto.Description
	}

	logger := logging.GetLoggerFromContext(ctx).WithFields(logFields)
	ctx = logging.NewContextFromLogger(ctx, logger)

	vl := validator.ChainValidator(
		validator.StructValidator(dto),
		validator.UUIDValidator(chatIDParam, chatID),
	)

	if err := h.validate(vl); err != nil {
		respondError(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	dto.ID = chatID
	dto.CreatorID = authUser.UserID

	chat, err := h.chatService.Update(ctx, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondError(ctx, w, errChatNotFound.Wrap(err))
		default:
			respondError(ctx, w, err)
		}

		return
	}

	respondSuccess(ctx, http.StatusOK, w, encoding.NewJSONChatMarshaler(chat))
}

// @Summary Delete chat where authenticated user is creator
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Success 204 "No Content"
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id} [delete]
func (h *chatHandler) delete(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	chatID := httprouter.ParamsFromContext(ctx).ByName(chatIDParam)
	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{"chat_id": chatID})
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.UUIDValidator(chatIDParam, chatID)); err != nil {
		respondError(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: authUser.UserID,
		ChatID: chatID,
	}

	err := h.chatService.Delete(ctx, memberKey)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondError(ctx, w, errChatNotFound.Wrap(err))
		default:
			respondError(ctx, w, err)
		}

		return
	}

	respondSuccess(ctx, http.StatusNoContent, w, nil)
}
