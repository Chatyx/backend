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
	currentUserURL  = "/api/user"
	userPasswordURI = "/api/user/password"
	listUserURL     = "/api/users"
	detailUserURI   = "/api/users/:id"
)

type UserListResponse struct {
	List []domain.User `json:"list"`
}

func (r UserListResponse) Encode() ([]byte, error) {
	return json.Marshal(r)
}

type userHandler struct {
	*baseHandler
	userService service.UserService
	logger      logging.Logger
}

func newUserHandler(us service.UserService, validate *validator.Validate) *userHandler {
	logger := logging.GetLogger()

	return &userHandler{
		baseHandler: &baseHandler{
			logger:   logger,
			validate: validate,
		},
		userService: us,
		logger:      logger,
	}
}

func (h *userHandler) register(router *httprouter.Router, authMid Middleware) {
	router.Handler(http.MethodGet, listUserURL, authMid(http.HandlerFunc(h.list)))
	router.HandlerFunc(http.MethodPost, listUserURL, h.create)
	router.Handler(http.MethodGet, detailUserURI, authMid(http.HandlerFunc(h.detail)))
	router.Handler(http.MethodPut, currentUserURL, authMid(http.HandlerFunc(h.update)))
	router.Handler(http.MethodPut, userPasswordURI, authMid(http.HandlerFunc(h.updatePassword)))
	router.Handler(http.MethodDelete, currentUserURL, authMid(http.HandlerFunc(h.delete)))
}

// @Summary Get list of users
// @Tags Users
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Success 200 {object} UserListResponse
// @Failure 500 {object} ResponseError
// @Router /users [get]
func (h *userHandler) list(w http.ResponseWriter, req *http.Request) {
	users, err := h.userService.List(req.Context())
	if err != nil {
		respondError(w, errInternalServer)
		return
	}

	respondSuccess(http.StatusOK, w, UserListResponse{List: users})
}

// @Summary Get user by id
// @Tags Users
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param id path string true "User id"
// @Success 200 {object} domain.User
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /users/{id} [get]
func (h *userHandler) detail(w http.ResponseWriter, req *http.Request) {
	ps := httprouter.ParamsFromContext(req.Context())

	user, err := h.userService.GetByID(req.Context(), ps.ByName("id"))
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

// @Summary Create user
// @Tags Users
// @Accept json
// @Produce json
// @Param input body domain.CreateUserDTO true "Create body"
// @Success 201 {object} domain.User
// @Failure 400 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /users [post]
func (h *userHandler) create(w http.ResponseWriter, req *http.Request) {
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

// @Summary Update current authenticated user
// @Security JWTTokenAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param input body domain.UpdateUserDTO true "Update body"
// @Success 200 {object} domain.User
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /user [put]
func (h *userHandler) update(w http.ResponseWriter, req *http.Request) {
	dto := domain.UpdateUserDTO{}
	if err := h.decodeJSONFromBody(req.Body, &dto); err != nil {
		respondError(w, err)
		return
	}

	dto.ID = domain.UserIDFromContext(req.Context())

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

// @Summary Update current authenticated user's password
// @Security JWTTokenAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param input body domain.UpdateUserPasswordDTO true "Update body"
// @Success 204 "No Content"
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /user/password [put]
func (h *userHandler) updatePassword(w http.ResponseWriter, req *http.Request) {
	dto := domain.UpdateUserPasswordDTO{}
	if err := h.decodeJSONFromBody(req.Body, &dto); err != nil {
		respondError(w, err)
		return
	}

	dto.UserID = domain.UserIDFromContext(req.Context())

	if err := h.validateStruct(dto); err != nil {
		respondError(w, err)
		return
	}

	if err := h.userService.UpdatePassword(req.Context(), dto); err != nil {
		switch err {
		case domain.ErrWrongCurrentPassword:
			respondError(w, ResponseError{StatusCode: http.StatusBadRequest, Message: err.Error()})
		case domain.ErrUserNotFound:
			respondError(w, ResponseError{StatusCode: http.StatusNotFound, Message: err.Error()})
		default:
			respondError(w, errInternalServer)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}

// @Summary Delete current authenticated user
// @Security JWTTokenAuth
// @Tags Users
// @Accept json
// @Produce json
// @Success 204 "No Content"
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /user [delete]
func (h *userHandler) delete(w http.ResponseWriter, req *http.Request) {
	err := h.userService.Delete(req.Context(), domain.UserIDFromContext(req.Context()))
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
