package websocket

import (
	"context"
	"net/http"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	ws "github.com/gorilla/websocket"
)

type chatSession struct {
	conn       *ws.Conn
	userID     string
	msgService service.MessageService
	logger     logging.Logger
}

func (s *chatSession) Serve() {
	defer func() {
		if err := s.conn.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close websocket connection")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inCh, outCh, errCh := s.msgService.NewServeSession(ctx, s.userID)
	defer close(inCh)

	for {
		select {
		case <-errCh:
			return
		case <-s.readMessages(inCh):
			return
		case msg := <-outCh:
			payload, err := msg.Encode()
			if err != nil {
				s.logger.WithError(err).Error("An error occurred while marshalling the message")
				return
			}

			if err = s.conn.WriteMessage(ws.TextMessage, payload); err != nil {
				s.logger.WithError(err).Error("An error occurred while writing the message to websocket")
				return
			}
		}
	}
}

func (s *chatSession) readMessages(inCh chan<- domain.Message) <-chan error {
	errCh := make(chan error, 0)

	go func() {
		for {
			close(errCh)

			_, payload, err := s.conn.ReadMessage()
			if err != nil {
				if closeErr, ok := err.(*ws.CloseError); ok {
					s.logger.Infof("User (id=%s) closed the websocket connection (%v)", s.userID, closeErr)
				} else {
					s.logger.WithError(err).Error("An error occurred while reading the message from websocket")
				}

				return
			}

			var message domain.Message
			if err = message.Decode(payload); err != nil {
				s.logger.WithError(err).Error("An error occurred while unmarshalling the message")
				return
			}

			curTime := time.Now()
			message.SenderID = s.userID
			message.CreatedAt = &curTime

			inCh <- message
		}
	}()

	return errCh
}

type chatSessionHandler struct {
	upgrader   *ws.Upgrader
	msgService service.MessageService
	logger     logging.Logger
}

func newChatSessionHandler(msgService service.MessageService) *chatSessionHandler {
	return &chatSessionHandler{
		upgrader: &ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		msgService: msgService,
		logger:     logging.GetLogger(),
	}
}

func (h *chatSessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, err := h.upgrader.Upgrade(w, req, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade protocol")
		h.respondError(w)

		return
	}

	chs := &chatSession{
		conn:       conn,
		logger:     h.logger,
		msgService: h.msgService,
		userID:     domain.UserIDFromContext(req.Context()),
	}
	go chs.Serve()

	w.WriteHeader(http.StatusNoContent)
}

func (h *chatSessionHandler) respondError(w http.ResponseWriter) {
	http.Error(w, "Failed to upgrade protocol", http.StatusInternalServerError)
}
