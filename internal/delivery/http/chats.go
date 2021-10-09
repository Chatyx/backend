package http

import (
	"encoding/json"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

const (
	listChatURI   = "/api/chats"
	detailChatURI = "/api/chats/:id"
)

type ChatListResponse struct {
	List []domain.Chat `json:"list"`
}

func (r ChatListResponse) Encode() ([]byte, error) {
	return json.Marshal(r)
}

type chatHandler struct {
	*baseHandler
	chatService service.ChatService
	authService service.AuthService
	logger      logging.Logger
}

func newChatHandler(chatService service.ChatService, authService service.AuthService, validate *validator.Validate) *chatHandler {
	logger := logging.GetLogger()

	return &chatHandler{
		baseHandler: &baseHandler{
			logger:   logger,
			validate: validate,
		},
		chatService: chatService,
		authService: authService,
		logger:      logger,
	}
}

func (h *chatHandler) register(router *httprouter.Router) {
	router.GET(listChatURI, authorizationMiddleware(h.list, h.authService))
	router.POST(listChatURI, authorizationMiddleware(h.create, h.authService))
	router.GET(detailChatURI, authorizationMiddleware(h.detail, h.authService))
	router.PUT(detailChatURI, authorizationMiddleware(h.update, h.authService))
	router.DELETE(detailChatURI, authorizationMiddleware(h.delete, h.authService))
}

// @Summary Get list of chats where user consists
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Success 200 {object} ChatListResponse
// @Failure 500 {object} ResponseError
// @Router /chats [get]
func (h *chatHandler) list(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	memberID := domain.UserIDFromContext(req.Context())

	chats, err := h.chatService.List(req.Context(), memberID)
	if err != nil {
		respondError(w, err)
	}

	respondSuccess(http.StatusOK, w, ChatListResponse{List: chats})
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
func (h *chatHandler) create(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	dto := domain.CreateChatDTO{}
	if err := h.decodeJSONFromBody(req.Body, &dto); err != nil {
		respondError(w, err)
		return
	}

	dto.CreatorID = domain.UserIDFromContext(req.Context())

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	chat, err := h.chatService.Create(req.Context(), dto)
	if err != nil {
		respondError(w, errInternalServer)
		return
	}

	respondSuccess(http.StatusCreated, w, &chat)
}

// @Summary Get chat by id where user consists
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param id path string true "Chat id"
// @Success 200 {object} domain.Chat
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{id} [get]
func (h *chatHandler) detail(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	chatID, memberID := params.ByName("id"), domain.UserIDFromContext(req.Context())

	chat, err := h.chatService.GetByID(req.Context(), chatID, memberID)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusOK, w, &chat)
}

// @Summary Update chat where user is creator
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param id path string true "Chat id"
// @Param input body domain.UpdateChatDTO true "Update body"
// @Success 200 {object} domain.Chat
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{id} [put]
func (h *chatHandler) update(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	dto := domain.UpdateChatDTO{}
	if err := h.decodeJSONFromBody(req.Body, &dto); err != nil {
		respondError(w, err)
		return
	}

	dto.ID = params.ByName("id")
	dto.CreatorID = domain.UserIDFromContext(req.Context())

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	chat, err := h.chatService.Update(req.Context(), dto)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusOK, w, &chat)
}

// @Summary Delete chat where user is creator
// @Tags Chats
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param id path string true "Chat id"
// @Success 204 "No Content"
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /chats/{id} [delete]
func (h *chatHandler) delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	chatID, creatorID := params.ByName("id"), domain.UserIDFromContext(req.Context())

	err := h.chatService.Delete(req.Context(), chatID, creatorID)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}
