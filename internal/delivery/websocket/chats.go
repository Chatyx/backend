package websocket

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type chatSessionHandler struct {
	chatService service.ChatService
	logger      logging.Logger
}

func (h *chatSessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
