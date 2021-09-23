package http

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/services"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/julienschmidt/httprouter"
)

func AuthorizationMiddleware(handler httprouter.Handle, as services.AuthService) httprouter.Handle {
	logger := logging.GetLogger()

	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		accessToken, err := ExtractTokenFromHeader(req.Header.Get("Authorization"))
		if err != nil {
			logger.WithError(err).Debug("invalid access token")
			RespondError(w, domain.ErrInvalidToken)

			return
		}

		_, err = as.Authorize(req.Context(), accessToken)
		if err != nil {
			RespondError(w, err)

			return
		}

		handler(w, req, params)
	}
}
