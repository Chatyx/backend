package http

import (
	"encoding/json"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/services"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/julienschmidt/httprouter"
)

const (
	listUserURL   = "/api/users"
	detailUserURI = "/api/users/:id"
)

type UserListResponse struct {
	List []domain.User `json:"list"`
}

func (r UserListResponse) Encode() ([]byte, error) {
	return json.Marshal(r)
}

type userHandler struct {
	*baseHandler
	userService services.UserService
	authService services.AuthService
	logger      logging.Logger
}

func (h *userHandler) register(router *httprouter.Router) {
	router.GET(listUserURL, authorizationMiddleware(h.list, h.authService))
	router.POST(listUserURL, h.create)
	router.GET(detailUserURI, authorizationMiddleware(h.detail, h.authService))
	router.PATCH(detailUserURI, authorizationMiddleware(h.update, h.authService))
	router.DELETE(detailUserURI, authorizationMiddleware(h.delete, h.authService))
}

func (h *userHandler) list(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	users, err := h.userService.List(req.Context())
	if err != nil {
		respondError(w, errInternalServer)
		return
	}

	respondSuccess(http.StatusOK, w, UserListResponse{List: users})
}

func (h *userHandler) detail(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := h.userService.GetByID(req.Context(), params.ByName("id"))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusOK, w, &user)
}

func (h *userHandler) create(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	dto := domain.CreateUserDTO{}
	if err := h.decodeJSONFromBody(req.Body, &dto); err != nil {
		respondError(w, err)
		return
	}

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	user, err := h.userService.Create(req.Context(), dto)
	if err != nil {
		switch err {
		case domain.ErrUserUniqueViolation:
			respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusCreated, w, &user)
}

func (h *userHandler) update(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	dto := domain.UpdateUserDTO{}
	if err := h.decodeJSONFromBody(req.Body, &dto); err != nil {
		respondError(w, err)
		return
	}

	dto.ID = params.ByName("id")

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	user, err := h.userService.Update(req.Context(), dto)
	if err != nil {
		switch err {
		case domain.ErrUserUniqueViolation:
			respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: err.Error()})
		case domain.ErrUserNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusOK, w, &user)
}

func (h *userHandler) delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	err := h.userService.Delete(req.Context(), params.ByName("id"))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}
