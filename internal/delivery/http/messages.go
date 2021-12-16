package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/julienschmidt/httprouter"
)

const (
	listMessageURI     = "/api/messages"
	listChatMessageURI = "/api/chats/:chat_id/messages"
)

type MessageListResponse struct {
	List []domain.Message `json:"list"`
}

func (r MessageListResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type messageHandler struct {
	*baseHandler
	msgService service.MessageService
}

func newMessageHandler(ms service.MessageService) *messageHandler {
	return &messageHandler{
		baseHandler: &baseHandler{logger: logging.GetLogger()},
		msgService:  ms,
	}
}

func (h *messageHandler) register(router *httprouter.Router, authMid Middleware) {
	router.Handler(http.MethodGet, listChatMessageURI, authMid(http.HandlerFunc(h.list)))
	router.Handler(http.MethodPost, listMessageURI, authMid(http.HandlerFunc(h.create)))
}

// @Summary Get chat's messages
// @Tags Messages
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat id"
// @Param timestamp query string true "Last timestamp of received message"
// @Success 200 {object} MessageListResponse
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/messages [get]
func (h *messageHandler) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)
	chatID, tsRaw := ps.ByName(chatIDParam), req.URL.Query().Get("timestamp")

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		"chat_id":   chatID,
		"timestamp": tsRaw,
	})
	ctx = logging.NewContextFromLogger(ctx, logger)

	timeValidator := validator.NewTimeValidator("timestamp", tsRaw, time.RFC3339Nano)
	vl := validator.ChainValidator(
		validator.UUIDValidator(chatIDParam, chatID),
		timeValidator,
	)

	if err := h.validate(vl); err != nil {
		respondError(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: authUser.UserID,
		ChatID: chatID,
	}

	messages, err := h.msgService.List(ctx, memberKey, timeValidator.Value())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondError(ctx, w, errChatNotFound.Wrap(err))
		default:
			respondError(ctx, w, err)
		}

		return
	}

	respondSuccess(ctx, http.StatusOK, w, MessageListResponse{List: messages})
}

// @Summary Send message to the chat
// @Tags Messages
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param input body domain.CreateMessageDTO true "Create body"
// @Success 201 {object} domain.Message
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /messages [post]
func (h *messageHandler) create(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	dto := domain.CreateMessageDTO{}

	if err := h.decodeBody(req.Body, encoding.NewJSONCreateMessageDTOUnmarshaler(&dto)); err != nil {
		respondError(ctx, w, err)
		return
	}

	logFields := logging.Fields{}
	if dto.ChatID != "" {
		logFields["chat_id"] = dto.ChatID
	}

	logger := logging.GetLoggerFromContext(ctx).WithFields(logFields)
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.StructValidator(dto)); err != nil {
		respondError(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	dto.ActionID = domain.MessageSendAction
	dto.SenderID = authUser.UserID

	message, err := h.msgService.Create(ctx, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondError(ctx, w, errChatNotFound.Wrap(err))
		default:
			respondError(ctx, w, err)
		}

		return
	}

	respondSuccess(ctx, http.StatusCreated, w, encoding.NewJSONMessageMarshaler(message))
}
