package v1

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	loginPath = "/auth/login"

	//nolint:gosec // G101: that's not credentials
	tokenPath = "/auth/token"
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

type AuthController struct {
}

func (ac *AuthController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodPost, loginPath, ac.login)
	mux.HandlerFunc(http.MethodPost, tokenPath, ac.refreshToken)
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
func (ac *AuthController) login(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}

// refreshToken refreshes authorization token
//
//	@Summary		Refresh authorization token
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
//	@Router			/auth/token [post]
func (ac *AuthController) refreshToken(w http.ResponseWriter, req *http.Request) {
	_, _ = w, req
}
