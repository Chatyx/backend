package http

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/services"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

const (
	listUserURL   = "/api/users"
	detailUserURI = "/api/users/:id"
)

type UserHandler struct {
	*Handler
	service services.UserService
	logger  logging.Logger
}

func NewUserHandler(service services.UserService, validate *validator.Validate) *UserHandler {
	logger := logging.GetLogger()

	return &UserHandler{
		Handler: &Handler{
			logger:   logger,
			validate: validate,
		},
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) Register(router *httprouter.Router) {
	router.GET(listUserURL, h.List)
	router.POST(listUserURL, h.Create)
	router.GET(detailUserURI, h.Detail)
	router.PUT(detailUserURI, h.Update)
	router.DELETE(detailUserURI, h.Delete)
}

func (h *UserHandler) List(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	panic("unimplemented")
}

func (h *UserHandler) Detail(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	panic("unimplemented")
}

func (h *UserHandler) Create(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	dto := domain.CreateUserDTO{}
	if err := h.DecodeJSONFromBody(req.Body, &dto); err != nil {
		h.RespondError(w, err)

		return
	}

	if err := h.Validate(dto); err != nil {
		h.RespondError(w, err)

		return
	}

	user, err := h.service.Create(req.Context(), dto)
	if err != nil {
		h.RespondError(w, err)

		return
	}

	h.RespondSuccess(http.StatusCreated, w, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	panic("unimplemented")
}

func (h *UserHandler) Delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	panic("unimplemented")
}
