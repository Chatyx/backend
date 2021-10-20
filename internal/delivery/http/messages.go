package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/Mort4lis/scht-backend/internal/domain"

	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/julienschmidt/httprouter"
)

const (
	listMessageURI = "/api/chats/:id/messages"
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
	router.Handler(http.MethodGet, listMessageURI, authMid(http.HandlerFunc(h.list)))
}

// @Summary Get chat's messages
// @Tags Messages
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param id path string true "Chat id"
// @Param timestamp query string true "Last timestamp of received message"
// @Success 200 {object} MessageListResponse
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{id}/messages [get]
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
	chatID := ps.ByName("id")

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
