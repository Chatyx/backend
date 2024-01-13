package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/Chatyx/backend/pkg/auth"
	"github.com/Chatyx/backend/pkg/auth/model"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/log"
	"github.com/Chatyx/backend/pkg/validator"

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
	service                 Service
	bodyParser              httputil.BodyParser
	fingerprintHeaderParser httputil.HeaderParser
	refreshTokenTTL         time.Duration
}

type Service interface {
	Login(ctx context.Context, cred model.Credentials, opts ...auth.MetaOption) (model.TokenPair, error)
	RefreshSession(ctx context.Context, rs model.RefreshSession, opts ...auth.MetaOption) (model.TokenPair, error)
	Logout(ctx context.Context, userID, refreshToken string) error
}

func NewController(v validator.Validate, srv Service, refreshTokenTTL time.Duration) *Controller {
	return &Controller{
		service:                 srv,
		bodyParser:              httputil.NewBodyParser(v),
		fingerprintHeaderParser: httputil.NewHeaderParser(v, "X-Fingerprint", "required"),
		refreshTokenTTL:         refreshTokenTTL,
	}
}

func (c *Controller) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodPost, loginPath, c.login)
	mux.HandlerFunc(http.MethodPost, logoutPath, c.logout)
	mux.HandlerFunc(http.MethodPost, refreshTokensPath, c.refreshTokens)
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
func (c *Controller) login(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var (
		dto         Credentials
		fingerprint string
	)

	if err := c.bodyParser.Parse(ctx, req, &dto); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}
	if err := c.fingerprintHeaderParser.Parse(ctx, req, &fingerprint); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	ctx = log.WithLogger(ctx, log.With("username", dto.Username))

	pair, err := c.service.Login(
		ctx,
		model.Credentials{
			Username:    dto.Username,
			Password:    dto.Password,
			Fingerprint: fingerprint,
		},
		auth.WithIP(net.ParseIP(req.RemoteAddr)),
	)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserNotFound):
			httputil.RespondError(ctx, w, ErrFailedLogin.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(
		ctx, w,
		http.StatusOK,
		TokenPair{
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
		},
	)
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
func (c *Controller) logout(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var dto RefreshToken

	cookie, err := req.Cookie("refresh_token")
	if err == nil {
		dto.Token = cookie.Value
	} else if err = c.bodyParser.Parse(ctx, req, &dto); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	// TODO set userID
	if err = c.service.Logout(ctx, "", dto.Token); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidRefreshToken):
			httputil.RespondError(ctx, w, ErrInvalidRefreshToken.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Domain:   "localhost:8080", // TODO
		MaxAge:   0,
		HttpOnly: true,
	})

	httputil.RespondSuccess(ctx, w, http.StatusNoContent, nil)
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
func (c *Controller) refreshTokens(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var (
		dto         RefreshToken
		fingerprint string
	)

	cookie, err := req.Cookie("refresh_token")
	if err == nil {
		dto.Token = cookie.Value
	} else if err = c.bodyParser.Parse(ctx, req, &dto); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err = c.fingerprintHeaderParser.Parse(ctx, req, &fingerprint); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	pair, err := c.service.RefreshSession(
		ctx,
		model.RefreshSession{
			UserID:       "", // TODO
			RefreshToken: dto.Token,
			Fingerprint:  fingerprint,
		},
		auth.WithIP(net.ParseIP(req.RemoteAddr)),
	)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidRefreshToken):
			httputil.RespondError(ctx, w, ErrInvalidRefreshToken.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    pair.RefreshToken,
		Path:     "/api/v1/auth",
		Domain:   "localhost:8080", // TODO
		Expires:  time.Now().Add(c.refreshTokenTTL),
		HttpOnly: true,
	})

	httputil.RespondSuccess(
		ctx, w,
		http.StatusOK,
		TokenPair{
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
		},
	)
}
