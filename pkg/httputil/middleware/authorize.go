package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/log"

	"github.com/golang-jwt/jwt/v5"
)

func Authorize(signedKey any) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			tokenStr, err := extractTokenFromRequest(req)
			if err != nil {
				httputil.RespondError(ctx, w, httputil.ErrInvalidAuthorization.Wrap(err))
				return
			}

			token, err := jwt.Parse(tokenStr, func(*jwt.Token) (interface{}, error) {
				return signedKey, nil
			})
			if err != nil {
				httputil.RespondError(ctx, w, httputil.ErrInvalidAuthorization.Wrap(err))
				return
			}

			subject, err := token.Claims.GetSubject()
			if err != nil {
				httputil.RespondError(ctx, w, httputil.ErrInvalidAuthorization.Wrap(err))
				return
			}

			logger := log.FromContext(ctx).With("user_id", subject)

			ctx = log.WithLogger(ctxutil.WithUserID(ctx, ctxutil.UserID(subject)), logger)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func extractTokenFromRequest(req *http.Request) (string, error) {
	header := req.Header.Get("Authorization")
	if header == "" {
		// browser's websocket API doesn't support natively pass http Headers to establish connection
		token := req.URL.Query().Get("token")
		if token == "" {
			return "", errors.New("authorization header and token query are both empty")
		}

		return token, nil
	}

	headerParts := strings.SplitN(header, " ", 2)
	if headerParts[0] != "Bearer" {
		return "", errors.New("authorization header doesn't begin with Bearer")
	}
	if headerParts[1] == "" {
		return "", errors.New("authorization header value is empty")
	}

	return headerParts[1], nil
}
