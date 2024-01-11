package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	loginPath = "/auth/login"

	logoutPath = "/auth/logout"

	//nolint:gosec // G101: that's not credentials
	refreshTokensPath = "/auth/refresh-tokens"
)

type Credentials struct {
	Username string `json:"username" validate:"required,max=50"`
	Password string `json:"password" validate:"required,min=8,max=27"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	Token string `json:"refresh_token" validate:"required"`
}

type Controller struct {
}

func (ac *Controller) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodPost, loginPath, ac.login)
	mux.HandlerFunc(http.MethodPost, logoutPath, ac.logout)
	mux.HandlerFunc(http.MethodPost, refreshTokensPath, ac.refreshTokens)
}

// login performs user authentication
//
//	@Summary		User authentication
//	@Tags			auth
//	@Description	Direct authentication by username and password. Successful
//	response includes http-only cookie with refresh token.
//	@Accept			json
//	@Produce		json
//	@Param			fingerprint	header		string		true	"Fingerprint header"
//	@Param			input		body		Credentials	true	"Credentials body"
//	@Success		200			{object}	TokenPair
//	@Failure		400			{object}	httputil.Error
//	@Failure		401			{object}	httputil.Error
//	@Failure		500			{object}	httputil.Error
//	@Router			/auth/login [post]
func (ac *Controller) login(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// logout performs user logout
//
//	@Summary		User logout
//	@Tags			auth
//	@Description	Invalidate session by removing refresh token
//	@Produce		json
//	@Param			input	body	RefreshToken	true	"Refresh token body"
//	@Success		204		"No Content"
//	@Failure		500		{object}	httputil.Error
//	@Security		JWTAuth
//	@Router			/auth/logout [post]
func (ac *Controller) logout(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// refreshTokens refresh access and refresh tokens
//
//	@Summary		Refresh access and refresh token
//	@Tags			auth
//	@Description	Allows to get a pair of tokens (access and refresh)  by exchanging an existing token.
//	 Successful response includes http-only cookie with refresh token.
//	@Accept			json
//	@Produce		json
//	@Param			fingerprint	header		string			true	"Fingerprint header"
//	@Param			input		body		RefreshToken	true	"Refresh token body"
//	@Success		200			{object}	TokenPair
//	@Failure		400			{object}	httputil.Error
//	@Failure		500			{object}	httputil.Error
//	@Security		JWTAuth
//	@Router			/auth/refresh-tokens [post]
func (ac *Controller) refreshTokens(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}
