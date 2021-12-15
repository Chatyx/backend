package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/julienschmidt/httprouter"
)

const (
	currentUserURL  = "/api/user"
	userPasswordURI = "/api/user/password"
	listUserURL     = "/api/users"
	detailUserURI   = "/api/users/:user_id"
)

const (
	userIDParam = "user_id"
)

type UserListResponse struct {
	List []domain.User `json:"list"`
}

func (r UserListResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type userHandler struct {
	*baseHandler
	userService service.UserService
}

func newUserHandler(us service.UserService) *userHandler {
	return &userHandler{
		baseHandler: &baseHandler{logger: logging.GetLogger()},
		userService: us,
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
		respondErrorRefactored(req.Context(), w, err)
		return
	}

	respondSuccess(http.StatusOK, w, UserListResponse{List: users})
}

// @Summary Get user by id
// @Tags Users
// @Security JWTTokenAuth
// @Accept json
// @Produce json
// @Param user_id path string true "User id"
// @Success 200 {object} domain.User
// @Failure 400,404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /users/{user_id} [get]
func (h *userHandler) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID := httprouter.ParamsFromContext(ctx).ByName(userIDParam)
	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{"user_id": userID})

	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.UUIDValidator(userIDParam, userID)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			respondErrorRefactored(ctx, w, errUserNotFound.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusOK, w, encoding.NewJSONUserMarshaler(user))
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
	ctx := req.Context()
	dto := domain.CreateUserDTO{}

	if err := h.decodeBody(req.Body, encoding.NewJSONCreateUserDTOUnmarshaler(&dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	logFields := logging.Fields{}
	if dto.Username != "" {
		logFields["username"] = dto.Username
	}

	if dto.Email != "" {
		logFields["email"] = dto.Email
	}

	if dto.FirstName != "" {
		logFields["first_name"] = dto.FirstName
	}

	if dto.LastName != "" {
		logFields["last_name"] = dto.LastName
	}

	if dto.BirthDate != "" {
		logFields["birth_date"] = dto.BirthDate
	}

	if dto.Department != "" {
		logFields["department"] = dto.Department
	}

	logger := logging.GetLoggerFromContext(ctx).WithFields(logFields)
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.StructValidator(dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	user, err := h.userService.Create(ctx, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserUniqueViolation):
			respondErrorRefactored(ctx, w, errUserUniqueViolation.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusCreated, w, encoding.NewJSONUserMarshaler(user))
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
	ctx := req.Context()
	dto := domain.UpdateUserDTO{}

	if err := h.decodeBody(req.Body, encoding.NewJSONUpdateUserDTOUnmarshaler(&dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	logFields := logging.Fields{}
	if dto.Username != "" {
		logFields["username"] = dto.Username
	}

	if dto.Email != "" {
		logFields["email"] = dto.Email
	}

	if dto.FirstName != "" {
		logFields["first_name"] = dto.FirstName
	}

	if dto.LastName != "" {
		logFields["last_name"] = dto.LastName
	}

	if dto.BirthDate != "" {
		logFields["birth_date"] = dto.BirthDate
	}

	if dto.Department != "" {
		logFields["department"] = dto.Department
	}

	logger := logging.GetLoggerFromContext(ctx).WithFields(logFields)
	ctx = logging.NewContextFromLogger(ctx, logger)

	if err := h.validate(validator.StructValidator(dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	dto.ID = authUser.UserID

	user, err := h.userService.Update(ctx, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserUniqueViolation):
			respondErrorRefactored(ctx, w, errUserUniqueViolation.Wrap(err))
		case errors.Is(err, domain.ErrUserNotFound):
			respondErrorRefactored(ctx, w, errUserNotFound.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusOK, w, encoding.NewJSONUserMarshaler(user))
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
	ctx := req.Context()
	dto := domain.UpdateUserPasswordDTO{}

	if err := h.decodeBody(req.Body, encoding.NewJSONUpdateUserPasswordDTOUnmarshaler(&dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	if err := h.validate(validator.StructValidator(dto)); err != nil {
		respondErrorRefactored(ctx, w, err)
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	dto.UserID = authUser.UserID

	if err := h.userService.UpdatePassword(ctx, dto); err != nil {
		switch {
		case errors.Is(err, domain.ErrWrongCurrentPassword):
			respondErrorRefactored(ctx, w, errWrongCurrentPassword.Wrap(err))
		case errors.Is(err, domain.ErrUserNotFound):
			respondErrorRefactored(ctx, w, errUserNotFound.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
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
	ctx := req.Context()
	authUser := domain.AuthUserFromContext(ctx)

	err := h.userService.Delete(ctx, authUser.UserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			respondErrorRefactored(ctx, w, errUserNotFound.Wrap(err))
		default:
			respondErrorRefactored(ctx, w, err)
		}

		return
	}

	respondSuccess(http.StatusNoContent, w, nil)
}
