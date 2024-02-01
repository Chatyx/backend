package websocket

import (
	"net/http"

	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/log"

	ws "github.com/gorilla/websocket"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

type ClientSessionInitHandler struct {
	upgrader *ws.Upgrader
	manager  MessageServeManager
}

func NewClientSessionInitHandler(manager MessageServeManager) *ClientSessionInitHandler {
	return &ClientSessionInitHandler{
		upgrader: &ws.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: fix it
		},
		manager: manager,
	}
}

func (h *ClientSessionInitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(req.Context())
	userID := ctxutil.UserIDFromContext(req.Context())

	conn, err := h.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.WithError(err).Debug("Failed to upgrade protocol")
		return
	}

	sess := ClientSession{conn: conn, userID: userID, logger: logger, manager: h.manager}
	go sess.Serve()

	logger.Info("User successfully opened websocket connection")
}
