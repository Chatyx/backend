package http

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/services"

	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"

	"github.com/Mort4lis/scht-backend/internal/domain"

	"github.com/julienschmidt/httprouter"
)

type AuthHandler struct {
	*Handler
	service services.AuthService
	logger  logging.Logger
}

func NewAuthHandler(service services.AuthService, validate *validator.Validate) *AuthHandler {
	logger := logging.GetLogger()

	return &AuthHandler{
		Handler: &Handler{
			logger:   logger,
			validate: validate,
		},
		service: service,
		logger:  logger,
	}
}

func (h *AuthHandler) Register(router *httprouter.Router) {
	router.POST("/api/sign-in", h.SignIn)
	router.POST("/api/refresh", h.Refresh)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var dto domain.SignInDTO
	if err := h.DecodeJSONFromBody(req.Body, &dto); err != nil {
		RespondError(w, err)
		return
	}

	if err := h.Validate(dto); err != nil {
		RespondError(w, err)
		return
	}

	pair, err := h.service.SignIn(req.Context(), dto)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondSuccess(http.StatusOK, w, pair)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, req *http.Request, params httprouter.Params) {}
