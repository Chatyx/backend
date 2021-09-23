package http

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"

	"github.com/Mort4lis/scht-backend/internal/services"
	"github.com/julienschmidt/httprouter"
)

func AuthorizationMiddleware(handler httprouter.Handle, as services.AuthService) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		accessToken, err := ExtractTokenFromHeader(req.Header.Get("Authorization"))
		if err != nil {
			RespondError(w, err)
			return
		}

		user, err := as.Authorize(req.Context(), accessToken)
		if err != nil {
			RespondError(w, err)
			return
		}

		ctx := domain.NewContextFromUser(req.Context(), user)

		handler(w, req.WithContext(ctx), params)
	}
}
