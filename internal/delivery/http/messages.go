package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
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

func newMessageHandler(ms service.MessageService, validate *validator.Validate) *messageHandler {
	logger := logging.GetLogger()

	return &messageHandler{
		baseHandler: &baseHandler{
			logger:   logger,
			validate: validate,
		},
		msgService: ms,
		logger:     logger,
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
	tsRaw := req.URL.Query().Get("timestamp")

	ts, err := time.Parse(time.RFC3339Nano, tsRaw)
	if err != nil {
		h.logger.WithError(err).Debugf("Failed to convert to time.Time timestamp %s", tsRaw)
		respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: "failed to parse timestamp"})

		return
	}

	ps := httprouter.ParamsFromContext(req.Context())
	userID := domain.UserIDFromContext(req.Context())
	chatID := ps.ByName("chat_id")

	messages, err := h.msgService.List(req.Context(), chatID, userID, ts)
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

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	dto.ActionID = domain.MessageSendAction
	dto.SenderID = domain.UserIDFromContext(req.Context())

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
