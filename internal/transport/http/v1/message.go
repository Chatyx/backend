package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/julienschmidt/httprouter"
)

const (
	messageListPath = "/api/v1/messages"
)

const (
	chatIDParam   = "chat_id"
	chatTypeParam = "chat_type"
	idAfterParam  = "id_after"
	limitParam    = "limit"
	sortParam     = "sort"
)

const defaultLimit = 20

type MessageList struct {
	Total int       `json:"total"`
	Data  []Message `json:"data"`
}

func NewMessageList(messages []entity.Message) MessageList {
	data := make([]Message, len(messages))
	for i, message := range messages {
		data[i] = NewMessage(message)
	}

	return MessageList{
		Total: len(messages),
		Data:  data,
	}
}

type Message struct {
	ID          int        `json:"id"`
	SenderID    int        `json:"sender_id"`
	Content     string     `json:"content"`
	ContentType string     `json:"content_type"`
	IsService   bool       `json:"is_service"`
	SentAt      time.Time  `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
}

func NewMessage(message entity.Message) Message {
	return Message{
		ID:          message.ID,
		SenderID:    message.SenderID,
		Content:     message.Content,
		ContentType: message.ContentType,
		IsService:   message.IsService,
		SentAt:      message.SentAt,
		DeliveredAt: message.DeliveredAt,
	}
}

type MessageCreate struct {
	ChatID      int    `json:"chat_id"      validate:"required"`
	ChatType    string `json:"chat_type"    validate:"required,oneof=dialog group"`
	Content     string `json:"content"      validate:"required,max=2000"`
	ContentType string `json:"content_type" validate:"required,oneof=text image"`
}

func (mc MessageCreate) DTO() dto.MessageCreate {
	return dto.MessageCreate{
		ChatID: entity.ChatID{
			ID:   mc.ChatID,
			Type: entity.ChatType(mc.ChatType),
		},
		Content:     mc.Content,
		ContentType: mc.ContentType,
	}
}

type MessageService interface {
	List(ctx context.Context, obj dto.MessageList) ([]entity.Message, error)
	Create(ctx context.Context, obj dto.MessageCreate) (entity.Message, error)
}

type MessageControllerConfig struct {
	Service   MessageService
	Authorize middleware.Middleware
	Validator validator.Validator
}

type MessageController struct {
	service   MessageService
	authorize middleware.Middleware
	validator validator.Validator
}

func NewMessageController(conf MessageControllerConfig) *MessageController {
	return &MessageController{
		service:   conf.Service,
		authorize: conf.Authorize,
		validator: conf.Validator,
	}
}

func (mc *MessageController) Register(mux *httprouter.Router) {
	mux.Handler(http.MethodGet, messageListPath, mc.authorize(http.HandlerFunc(mc.list)))
	mux.Handler(http.MethodPost, messageListPath, mc.authorize(http.HandlerFunc(mc.create)))
}

// list lists messages for a specified chat
//
//	@Summary	List messages for a specified chat
//	@Tags		messages
//	@Accept		json
//	@Produce	json
//	@Param		chat_id		query		int		true	"Chat id for dialog or group"
//	@Param		chat_type	query		string	true	"Chat type (dialog or group)"
//	@Param		id_after	query		int		false	"Message id that excludes already-retrieved messages"
//	@Param		limit		query		int		false	"Number of items to list per page (default: 20, max: 100)"
//	@Param		sort		query		string	true	"Sort order (asc or desc)"
//	@Success	200			{object}	MessageList
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/messages  [get]
func (mc *MessageController) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	query := req.URL.Query()

	chatType := query.Get(chatTypeParam)
	sort := query.Get(sortParam)

	chatID, err := strconv.Atoi(query.Get(chatIDParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodeQueryParamsFailed.Wrap(err))
		return
	}

	idAfter := 0
	if idAfterRaw := query.Get(idAfterParam); idAfterRaw != "" {
		idAfter, err = strconv.Atoi(idAfterRaw)
		if err != nil {
			httputil.RespondError(ctx, w, httputil.ErrDecodeQueryParamsFailed.Wrap(err))
			return
		}
	}

	limit := defaultLimit
	if limitRaw := query.Get(limitParam); limitRaw != "" {
		limit, err = strconv.Atoi(limitRaw)
		if err != nil {
			httputil.RespondError(ctx, w, httputil.ErrDecodeQueryParamsFailed.Wrap(err))
			return
		}
	}

	if err = validator.MergeResults(
		mc.validator.Var(chatType, chatTypeParam, "required,oneof=dialog group"),
		mc.validator.Var(sort, sortParam, "required,oneof=asc desc"),
		mc.validator.Var(limit, limitParam, "gt=0,max=100"),
	); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	obj := dto.MessageList{
		ChatID: entity.ChatID{
			ID:   chatID,
			Type: entity.ChatType(chatType),
		},
		IDAfter: idAfter,
		Limit:   limit,
		Sort:    dto.Sort(sort),
	}

	messages, err := mc.service.List(ctx, obj)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		case errors.Is(err, entity.ErrDialogNotFound):
			httputil.RespondError(ctx, w, errDialogNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewMessageList(messages))
}

// create sends message to the specified chat
//
//	@Summary	Send message to the specified chat
//	@Tags		messages
//	@Accept		json
//	@Produce	json
//	@Param		input	body		MessageCreate	true	"Body to create"
//	@Success	201		{object}	Message
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/messages  [post]
func (mc *MessageController) create(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var bodyObj MessageCreate

	if err := httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err := mc.validator.Struct(bodyObj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	message, err := mc.service.Create(ctx, bodyObj.DTO())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		case errors.Is(err, entity.ErrDialogNotFound):
			httputil.RespondError(ctx, w, errDialogNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusCreated, NewMessage(message))
}
