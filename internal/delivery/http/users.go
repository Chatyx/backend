package http

import (
	"encoding/json"
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

type UserListResponse struct {
	List []*domain.User `json:"list"`
}

func (r UserListResponse) Encode() ([]byte, error) {
	return json.Marshal(r)
}

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
	router.PATCH(detailUserURI, h.Update)
	router.DELETE(detailUserURI, h.Delete)
}

func (h *UserHandler) List(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	users, err := h.service.List(req.Context())
	if err != nil {
		RespondError(w, err)

		return
	}

	RespondSuccess(http.StatusOK, w, UserListResponse{List: users})
}

func (h *UserHandler) Detail(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := h.service.GetByID(req.Context(), params.ByName("id"))
	if err != nil {
		RespondError(w, err)

		return
	}

	RespondSuccess(http.StatusOK, w, user)
}

func (h *UserHandler) Create(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	dto := domain.CreateUserDTO{}
	if err := h.DecodeJSONFromBody(req.Body, &dto); err != nil {
		RespondError(w, err)

		return
	}

	if err := h.Validate(dto); err != nil {
		RespondError(w, err)

		return
	}

	user, err := h.service.Create(req.Context(), dto)
	if err != nil {
		RespondError(w, err)

		return
	}

	RespondSuccess(http.StatusCreated, w, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	dto := domain.UpdateUserDTO{}
	if err := h.DecodeJSONFromBody(req.Body, &dto); err != nil {
		RespondError(w, err)

		return
	}

	dto.ID = params.ByName("id")

	if err := h.Validate(dto); err != nil {
		RespondError(w, err)

		return
	}

	user, err := h.service.Update(req.Context(), dto)
	if err != nil {
		RespondError(w, err)

		return
	}

	RespondSuccess(http.StatusOK, w, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	err := h.service.Delete(req.Context(), params.ByName("id"))
	if err != nil {
		RespondError(w, err)

		return
	}

	RespondSuccess(http.StatusNoContent, w, nil)
}
