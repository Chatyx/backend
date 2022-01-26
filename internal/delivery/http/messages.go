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
	"github.com/Mort4lis/scht-backend/pkg/paginator"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/julienschmidt/httprouter"
)

const (
	listMessageURI     = "/api/messages"
	listChatMessageURI = "/api/chats/:chat_id/messages"
)

const (
	offsetDateQuery = "offset_date"
	directionQuery  = "direction"
	limitQuery      = "limit"
	offsetQuery     = "offset"
)

type MessageListResponse struct {
	Total   int              `json:"total"`
	HasNext bool             `json:"has_next"`
	HasPrev bool             `json:"has_prev"`
	Result  []domain.Message `json:"result"`
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
// @Param offset_date query string true "Date from which pagination will be performed (date format RFC3339Nano)"
// @Param direction query string true "Direction of pagination towards newer or older messages"
// @Param limit query int true "Number of result to return"
// @Param offset query int true "Number of messages to be skipped"
// @Success 200 {object} MessageListResponse
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/messages [get]
func (h *messageHandler) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	query := req.URL.Query()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName(chatIDParam)
	offsetDateStr := query.Get(offsetDateQuery)
	direction := query.Get(directionQuery)
	limitStr := query.Get(limitQuery)
	offsetStr := query.Get(offsetQuery)

	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		chatIDParam:     chatID,
		offsetDateQuery: offsetDateStr,
		directionQuery:  direction,
		limitQuery:      limitStr,
		offsetQuery:     offsetStr,
	})
	ctx = logging.NewContextFromLogger(ctx, logger)

	offsetDateValidator := validator.NewTimeValidator(offsetDateQuery, offsetDateStr, time.RFC3339Nano, false)
	limitValidator := validator.NewIntValidator(limitQuery, limitStr, false)
	offsetValidator := validator.NewIntValidator(offsetQuery, offsetStr, false)

	vl := validator.ChainValidator(
		validator.UUIDValidator(chatIDParam, chatID),
		offsetDateValidator,
		limitValidator,
		offsetValidator,
	)

	if err := h.validate(vl); err != nil {
		respondError(ctx, w, err)
		return
	}

	dto := domain.NewMessageListDTO(offsetDateValidator.Value(), direction, limitValidator.Value(), offsetValidator.Value())
	if err := h.validate(validator.StructValidator(dto)); err != nil {
		respondError(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	memberKey := domain.ChatMemberIdentity{
		UserID: authUser.UserID,
		ChatID: chatID,
	}

	messageList, err := h.msgService.List(ctx, memberKey, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrChatNotFound):
			respondError(ctx, w, errChatNotFound.Wrap(err))
		default:
			respondError(ctx, w, err)
		}

		return
	}

	paginate := paginator.LimitOffsetPaginate{
		Total:  messageList.Total,
		Count:  len(messageList.Messages),
		Limit:  limitValidator.Value(),
		Offset: offsetValidator.Value(),
	}

	respondSuccess(ctx, http.StatusOK, w, MessageListResponse{
		Total:   messageList.Total,
		HasNext: paginate.HasNext(),
		HasPrev: paginate.HasPrev(),
		Result:  messageList.Messages,
	})
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
