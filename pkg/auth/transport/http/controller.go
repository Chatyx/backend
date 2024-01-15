package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	core "github.com/Chatyx/backend/pkg/auth"
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

const (
	fingerprintHeaderKey = "X-Fingerprint"
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

type RTCookieSettings struct {
	Name       string
	Domain     string
	PrefixPath string
	TTL        time.Duration
}

type Config struct {
	RTCookieSettings RTCookieSettings
}

type Option func(c *Config)

func WithRTCookieName(s string) Option {
	return func(c *Config) {
		c.RTCookieSettings.Name = s
	}
}

func WithRTCookieDomain(s string) Option {
	return func(c *Config) {
		c.RTCookieSettings.Domain = s
	}
}

func WithRTCookiePrefixPath(s string) Option {
	return func(c *Config) {
		c.RTCookieSettings.PrefixPath = s
	}
}

func WithRTCookieTTL(d time.Duration) Option {
	return func(c *Config) {
		c.RTCookieSettings.TTL = d
	}
}

type Controller struct {
	service   Service
	validator validator.Validator
	rtc       RTCookieSettings
}

type Service interface {
	Login(ctx context.Context, cred core.Credentials, opts ...core.MetaOption) (core.TokenPair, error)
	RefreshSession(ctx context.Context, rs core.RefreshSession, opts ...core.MetaOption) (core.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

func NewController(srv Service, v validator.Validator, opts ...Option) *Controller {
	conf := Config{
		RTCookieSettings: RTCookieSettings{
			Name:   "refresh_token",
			Domain: "localhost:8080",
			TTL:    60 * 24 * time.Hour,
		},
	}
	for _, opt := range opts {
		opt(&conf)
	}

	return &Controller{
		service:   srv,
		validator: v,
		rtc:       conf.RTCookieSettings,
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

	var dto Credentials

	if err := httputil.DecodeBody(req.Body, &dto); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	fingerprint := req.Header.Get(fingerprintHeaderKey)

	if err := validator.MergeResults(
		c.validator.Struct(dto),
		c.validator.Var(fingerprint, "required"),
	); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	ctx = log.WithLogger(ctx, log.With("username", dto.Username))

	pair, err := c.service.Login(
		ctx,
		core.Credentials{
			Username:    dto.Username,
			Password:    dto.Password,
			Fingerprint: fingerprint,
		},
		core.WithIP(net.ParseIP(req.RemoteAddr)),
	)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrUserNotFound):
			httputil.RespondError(ctx, w, ErrFailedLogin.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     c.rtc.Name,
		Value:    pair.RefreshToken,
		Path:     c.rtc.PrefixPath + "/auth",
		Domain:   c.rtc.Domain,
		Expires:  time.Now().Add(c.rtc.TTL),
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

// logout performs user logout
//
//	@Summary		User logout
//	@Tags			auth
//	@Description	Invalidate session by removing refresh token
//	@Produce		json
//	@Param			input	body	RefreshToken	true	"Refresh token body"
//	@Success		204		"No Content"
//	@Failure		500		{object}	httputil.Error
//	@Router			/auth/logout [post]
func (c *Controller) logout(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var dto RefreshToken

	cookie, err := req.Cookie(c.rtc.Name)
	if err == nil {
		dto.Token = cookie.Value
	} else if err = httputil.DecodeBody(req.Body, &dto); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err = c.validator.Struct(dto); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	if err = c.service.Logout(ctx, dto.Token); err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidRefreshToken):
			httputil.RespondError(ctx, w, ErrInvalidRefreshToken.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     c.rtc.Name,
		Path:     c.rtc.PrefixPath + "/auth",
		Domain:   c.rtc.Domain,
		MaxAge:   -1,
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
//	@Router			/auth/refresh-tokens [post]
func (c *Controller) refreshTokens(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var dto RefreshToken

	cookie, err := req.Cookie(c.rtc.Name)
	if err == nil {
		dto.Token = cookie.Value
	} else if err = httputil.DecodeBody(req.Body, &dto); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	fingerprint := req.Header.Get(fingerprintHeaderKey)

	if err = validator.MergeResults(
		c.validator.Struct(dto),
		c.validator.Var(fingerprint, "required"),
	); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	pair, err := c.service.RefreshSession(
		ctx,
		core.RefreshSession{
			RefreshToken: dto.Token,
			Fingerprint:  fingerprint,
		},
		core.WithIP(net.ParseIP(req.RemoteAddr)),
	)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidRefreshToken):
			httputil.RespondError(ctx, w, ErrInvalidRefreshToken.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     c.rtc.Name,
		Value:    pair.RefreshToken,
		Path:     c.rtc.PrefixPath + "/auth",
		Domain:   c.rtc.Domain,
		Expires:  time.Now().Add(c.rtc.TTL),
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
