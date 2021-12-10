package http

import (
	"encoding/json"
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
	logger     logging.Logger
}

func newMessageHandler(ms service.MessageService) *messageHandler {
	logger := logging.GetLogger()

	return &messageHandler{
		baseHandler: &baseHandler{logger: logger},
		msgService:  ms,
		logger:      logger,
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
	ps := httprouter.ParamsFromContext(req.Context())
	chatID, tsRaw := ps.ByName(chatIDParam), req.URL.Query().Get("timestamp")

	timeValidator := validator.NewTimeValidator("timestamp", tsRaw, time.RFC3339Nano)
	vl := validator.ChainValidator(
		validator.UUIDValidator(chatIDParam, chatID),
		timeValidator,
	)

	if err := h.validate(vl); err != nil {
		respondError(w, err)
		return
	}

	authUser := domain.AuthUserFromContext(req.Context())
	memberKey := domain.ChatMemberIdentity{
		UserID: authUser.UserID,
		ChatID: chatID,
	}

	messages, err := h.msgService.List(req.Context(), memberKey, timeValidator.Value())
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, err)
		}

		return
	}

	respondSuccess(http.StatusOK, w, MessageListResponse{List: messages})
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
	dto := domain.CreateMessageDTO{}
	if err := h.decodeBody(req.Body, encoding.NewJSONCreateMessageDTOUnmarshaler(&dto)); err != nil {
		respondError(w, err)
		return
	}

	if err := h.validate(validator.StructValidator(dto)); err != nil {
		respondError(w, err)
		return
	}

	authUser := domain.AuthUserFromContext(req.Context())
	dto.ActionID = domain.MessageSendAction
	dto.SenderID = authUser.UserID

	message, err := h.msgService.Create(req.Context(), dto)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, err)
		}

		return
	}

	respondSuccess(http.StatusCreated, w, encoding.NewJSONMessageMarshaler(message))
}
