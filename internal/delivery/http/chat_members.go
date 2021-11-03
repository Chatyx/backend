package http

import (
	"encoding/json"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/encoding"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
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
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/members [get]
func (h *chatMemberHandler) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	userID := domain.UserIDFromContext(ctx)

	members, err := h.chatMemberService.ListByChatID(ctx, chatID, userID)
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
	userID := req.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "validation error",
			Fields:     ErrorFields{"user_id": "must be required"},
		})

		return
	}

	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	creatorID := domain.UserIDFromContext(ctx)

	if err := h.chatMemberService.JoinToChat(ctx, chatID, creatorID, userID); err != nil {
		switch err {
		case domain.ErrChatNotFound, domain.ErrUserNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		case domain.ErrChatMemberUniqueViolation:
			respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: err.Error()})
		default:
			respondError(w, err)
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
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/members/{user_id} [get]
func (h *chatMemberHandler) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	userID := ps.ByName("user_id")
	authUserID := domain.UserIDFromContext(ctx)

	member, err := h.chatMemberService.GetAnother(ctx, authUserID, chatID, userID)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound, domain.ErrChatMemberNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, err)
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
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{chat_id}/member [get]
func (h *chatMemberHandler) detailCurrent(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	chatID := ps.ByName("chat_id")
	userID := domain.UserIDFromContext(ctx)

	member, err := h.chatMemberService.Get(ctx, chatID, userID)
	if err != nil {
		switch err {
		case domain.ErrChatMemberNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, err)
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
	dto := domain.UpdateChatMemberDTO{}
	if err := h.decodeBody(req.Body, encoding.NewJSONUpdateChaMemberDTOUnmarshaler(&dto)); err != nil {
		respondError(w, err)
		return
	}

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	dto.ChatID = ps.ByName("chat_id")
	dto.UserID = domain.UserIDFromContext(ctx)

	if err := h.chatMemberService.UpdateStatus(ctx, dto); err != nil {
		switch err {
		case domain.ErrChatMemberNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		case domain.ErrChatMemberWrongStatusTransit, domain.ErrChatCreatorInvalidUpdateStatus:
			respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: err.Error()})
		default:
			respondError(w, err)
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
	dto := domain.UpdateChatMemberDTO{}
	if err := h.decodeBody(req.Body, encoding.NewJSONUpdateChaMemberDTOUnmarshaler(&dto)); err != nil {
		respondError(w, err)
		return
	}

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	ctx := req.Context()
	ps := httprouter.ParamsFromContext(ctx)

	dto.ChatID = ps.ByName("chat_id")
	dto.UserID = ps.ByName("user_id")
	creatorID := domain.UserIDFromContext(ctx)

	if err := h.chatMemberService.UpdateStatusByCreator(ctx, creatorID, dto); err != nil {
		switch err {
		case domain.ErrChatMemberNotFound, domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		case domain.ErrChatMemberWrongStatusTransit, domain.ErrChatCreatorInvalidUpdateStatus:
			respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: err.Error()})
		default:
			respondError(w, err)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}
