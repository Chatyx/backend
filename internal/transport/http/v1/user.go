package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/julienschmidt/httprouter"
)

const (
	userListPath   = "/api/v1/users"
	userDetailPath = "/api/v1/users/:user_id"
	userMePath     = "/api/v1/users/me"
)

const (
	userIDPathParam = "user_id"
)

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	BirthDate string `json:"birth_date,omitempty"`
	Bio       string `json:"bio,omitempty"`
}

func NewUser(user entity.User) User {
	res := User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Bio:       user.Bio,
	}
	if user.BirthDate != nil {
		res.BirthDate = user.BirthDate.Format("2006-01-02")
	}

	return res
}

type UserList struct {
	Total int    `json:"total"`
	Data  []User `json:"data"`
}

func NewUserList(users []entity.User) UserList {
	data := make([]User, len(users))
	for i, user := range users {
		data[i] = NewUser(user)
	}

	return UserList{
		Total: len(users),
		Data:  data,
	}
}

type UserUpdate struct {
	Username  string `json:"username"   validate:"required,max=50"`
	Email     string `json:"email"      validate:"required,email,max=255"`
	FirstName string `json:"first_name" validate:"max=50"`
	LastName  string `json:"last_name"  validate:"max=50"`
	BirthDate string `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Bio       string `json:"bio"        validate:"max=10000"`
}

func (u UserUpdate) DTO() dto.UserUpdate {
	obj := dto.UserUpdate{
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Bio:       u.Bio,
	}

	if u.BirthDate != "" {
		dt, _ := time.Parse("2006-01-02", u.BirthDate)
		obj.BirthDate = &dt
	}
	return obj
}

type UserCreate struct {
	UserUpdate
	Password string `json:"password" validate:"required,min=8,max=27"`
}

func (u UserCreate) DTO() dto.UserCreate {
	obj := dto.UserCreate{
		Username:  u.Username,
		Password:  u.Password,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Bio:       u.Bio,
	}

	if u.BirthDate != "" {
		dt, _ := time.Parse("2006-01-02", u.BirthDate)
		obj.BirthDate = &dt
	}
	return obj
}

type UserUpdatePassword struct {
	New     string `json:"new_password"     validate:"required,min=8,max=27"`
	Current string `json:"current_password" validate:"required,min=8,max=27"`
}

//go:generate mockery --inpackage --testonly --case underscore --name UserService
type UserService interface {
	List(ctx context.Context) ([]entity.User, error)
	Create(ctx context.Context, obj dto.UserCreate) (entity.User, error)
	GetByID(ctx context.Context, id int) (entity.User, error)
	Update(ctx context.Context, obj dto.UserUpdate) (entity.User, error)
	UpdatePassword(ctx context.Context, obj dto.UserUpdatePassword) error
	Delete(ctx context.Context, id int) error
}

type UserControllerConfig struct {
	Service   UserService
	Authorize middleware.Middleware
	Validator validator.Validator
}

type UserController struct {
	service   UserService
	authorize middleware.Middleware
	validator validator.Validator
}

func NewUserController(conf UserControllerConfig) *UserController {
	return &UserController{
		service:   conf.Service,
		authorize: conf.Authorize,
		validator: conf.Validator,
	}
}

func (uc *UserController) Register(mux *httprouter.Router) {
	mux.Handler(http.MethodGet, userListPath, uc.authorize(http.HandlerFunc(uc.list)))
	mux.HandlerFunc(http.MethodPost, userListPath, uc.create)
	mux.Handler(http.MethodGet, userDetailPath, uc.authorize(http.HandlerFunc(uc.detail)))
	mux.Handler(http.MethodPut, userMePath, uc.authorize(http.HandlerFunc(uc.update)))
	mux.Handler(http.MethodPut, userMePath+"/password", uc.authorize(http.HandlerFunc(uc.updatePassword)))
	mux.Handler(http.MethodDelete, userMePath, uc.authorize(http.HandlerFunc(uc.delete)))
}

// list lists all existing users
//
//	@Summary	List all existing users
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	UserList
//	@Failure	500	{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/users  [get]
func (uc *UserController) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	users, err := uc.service.List(ctx)
	if err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewUserList(users))
}

// create creates a new user
//
//	@Summary	Create a new user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		input	body		UserCreate	true	"Body to create"
//	@Success	201		{object}	User
//	@Failure	400		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/users [post]
func (uc *UserController) create(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var bodyObj UserCreate

	if err := httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err := uc.validator.Struct(bodyObj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	user, err := uc.service.Create(ctx, bodyObj.DTO())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrSuchUserAlreadyExists):
			httputil.RespondError(ctx, w, errSuchUserAlreadyExists.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusCreated, NewUser(user))
}

// detail gets a specified user
//
//	@Summary	Get a specified user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		user_id	path		int	true	"User identity"
//	@Success	200		{object}	User
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/users/{user_id} [get]
func (uc *UserController) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	userID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(userIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	user, err := uc.service.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUserNotFound):
			httputil.RespondError(ctx, w, errUserNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewUser(user))
}

// update updates information about the current authenticated user
//
//	@Summary	Update information about the current authenticated user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		input	body		UserUpdate	true	"Body to update"
//	@Success	200		{object}	User
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/users/me [put]
func (uc *UserController) update(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var bodyObj UserUpdate

	if err := httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err := uc.validator.Struct(bodyObj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	obj := bodyObj.DTO()
	obj.ID = ctxutil.UserIDFromContext(ctx).ToInt()

	user, err := uc.service.Update(ctx, obj)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrSuchUserAlreadyExists):
			httputil.RespondError(ctx, w, errSuchUserAlreadyExists.Wrap(err))
		case errors.Is(err, entity.ErrUserNotFound):
			httputil.RespondError(ctx, w, errUserNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewUser(user))
}

// updatePassword updates the current authenticated user's password
//
//	@Summary	Update the current authenticated user's password
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		input	body	UserUpdatePassword	true	"Body to update"
//	@Success	204		"No Content"
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/users/me/password [patch]
func (uc *UserController) updatePassword(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var obj UserUpdatePassword

	if err := httputil.DecodeBody(req.Body, &obj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err := uc.validator.Struct(obj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	err := uc.service.UpdatePassword(ctx, dto.UserUpdatePassword{
		UserID:      ctxutil.UserIDFromContext(ctx).ToInt(),
		CurPassword: obj.Current,
		NewPassword: obj.New,
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrWrongCurrentPassword):
			httputil.RespondError(ctx, w, errWrongCurrentPassword.Wrap(err))
		case errors.Is(err, entity.ErrUserNotFound):
			httputil.RespondError(ctx, w, errUserNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusNoContent, nil)
}

// delete deletes the current authenticated user
//
//	@Summary	Delete the current authenticated user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Success	204	"No Content"
//	@Failure	404	{object}	httputil.Error
//	@Failure	500	{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/users/me [delete]
func (uc *UserController) delete(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID := ctxutil.UserIDFromContext(ctx).ToInt()

	if err := uc.service.Delete(ctx, userID); err != nil {
		switch {
		case errors.Is(err, entity.ErrUserNotFound):
			httputil.RespondError(ctx, w, errUserNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusNoContent, nil)
}
