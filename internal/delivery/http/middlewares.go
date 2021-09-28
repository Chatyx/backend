package http

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/julienschmidt/httprouter"
)

func authorizationMiddleware(handler httprouter.Handle, as service.AuthService) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		accessToken, err := extractTokenFromHeader(req.Header.Get("Authorization"))
		if err != nil {
			respondError(w, err)
			return
		}

		claims, err := as.Authorize(accessToken)
		if err != nil {
			switch err {
			case domain.ErrInvalidAccessToken:
				respondError(w, errInvalidAccessToken)
			default:
				respondError(w, errInternalServer)
			}

			return
		}

		ctx := domain.NewContextFromUserID(req.Context(), claims.Subject)

		handler(w, req.WithContext(ctx), params)
	}
}
