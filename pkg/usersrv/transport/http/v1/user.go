package v1

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	listUsersPath   = "/users"
	detailUserPath  = "/users/:user_id"
	currentUserPath = "/users/me"
)

type UserDetail struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	BirthDate time.Time `json:"birth_date,omitempty"`
	Bio       string    `json:"bio,omitempty"`
}

type UserList struct {
	Total int          `json:"total"`
	Data  []UserDetail `json:"data"`
}

type UserUpdate struct {
	Username  string    `json:"username"   validate:"required,max=50"`
	Email     string    `json:"email"      validate:"required,email,max=255"`
	FirstName string    `json:"first_name" validate:"max=50"`
	LastName  string    `json:"last_name"  validate:"max=50"`
	BirthDate time.Time `json:"birth_date" validate:"datetime=2006-01-02"`
	Bio       string    `json:"bio"        validate:"max=10000"`
}

type UserCreate struct {
	UserUpdate
	Password string `json:"password" validate:"required,min=8,max=27"`
}

type UserUpdatePassword struct {
	New     string `json:"new_password"     validate:"required,min=8,max=27"`
	Current string `json:"current_password" validate:"required,min=8,max=27"`
}

type UserController struct {
}

func (uc *UserController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodGet, listUsersPath, uc.list)
	mux.HandlerFunc(http.MethodGet, detailUserPath, uc.detail)
	mux.HandlerFunc(http.MethodPost, listUsersPath, uc.create)
	mux.HandlerFunc(http.MethodPut, currentUserPath, uc.update)
	mux.HandlerFunc(http.MethodPut, currentUserPath+"/password", uc.updatePassword)
	mux.HandlerFunc(http.MethodDelete, currentUserPath, uc.delete)
}

// list lists all existing users
//
//	@Summary	List all existing users
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	UserList
//	@Failure	500	{object}	httputil.Error
//	@Router		/users  [get]
func (uc *UserController) list(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// detail gets a specified user
//
//	@Summary	Get a specified user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		user_id	path		int	true	"User identity"
//	@Success	200		{object}	UserDetail
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/users/{user_id} [get]
func (uc *UserController) detail(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// create creates a new user
//
//	@Summary	Create a new user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		input	body		UserCreate	true	"Body to create"
//	@Success	200		{object}	UserDetail
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/users [post]
func (uc *UserController) create(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// update updates information about the current authenticated user
//
//	@Summary	Update information about the current authenticated user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		input	body		UserUpdate	true	"Body to update"
//	@Success	200		{object}	UserDetail
//	@Failure	400		{object}	httputil.Error
//	@Failure	404		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Router		/users/me [put]
func (uc *UserController) update(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
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
//	@Failure	500		{object}	httputil.Error
//	@Router		/users/me/password [patch]
func (uc *UserController) updatePassword(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
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
//	@Router		/users/me [delete]
func (uc *UserController) delete(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}
