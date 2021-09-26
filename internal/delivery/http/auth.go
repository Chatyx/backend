package http

import (
	"net/http"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/services"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

const (
	signInURI  = "/api/auth/sign-in"
	refreshURI = "/api/auth/refresh"
)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	*Handler
	service services.AuthService
	logger  logging.Logger

	domain          string
	refreshTokenTTL time.Duration
}

func NewAuthHandler(service services.AuthService, validate *validator.Validate, domain string,
	refreshTokenTTL time.Duration) *AuthHandler {
	logger := logging.GetLogger()

	return &AuthHandler{
		Handler: &Handler{
			logger:   logger,
			validate: validate,
		},
		service:         service,
		domain:          domain,
		refreshTokenTTL: refreshTokenTTL,
		logger:          logger,
	}
}

func (h *AuthHandler) Register(router *httprouter.Router) {
	router.POST(signInURI, h.SignIn)
	router.POST(refreshURI, h.Refresh)
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
		switch err {
		case domain.ErrInvalidCredentials:
			RespondError(w, ResponseError{StatusCode: http.StatusUnauthorized, Message: err.Error()})
		default:
			RespondError(w, err)
		}

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    pair.RefreshToken,
		Path:     refreshURI,
		Domain:   h.domain,
		Expires:  time.Now().Add(h.refreshTokenTTL),
		HttpOnly: true,
	})

	RespondSuccess(http.StatusOK, w, pair)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var dto domain.RT
	if cookie, err := req.Cookie(refreshCookieName); err == nil {
		dto.RefreshToken = cookie.Value
	} else if err = h.DecodeJSONFromBody(req.Body, &dto); err != nil {
		RespondError(w, err)
		return
	}

	if err := h.Validate(dto); err != nil {
		RespondError(w, err)
		return
	}

	pair, err := h.service.Refresh(req.Context(), dto.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrSessionNotFound, domain.ErrUserNotFound:
			RespondError(w, ErrInvalidRefreshToken)
		default:
			RespondError(w, err)
		}

		return
	}

	RespondSuccess(http.StatusOK, w, pair)
}
