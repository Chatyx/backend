package websocket

import (
	"context"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	ws "github.com/gorilla/websocket"
)

type chatSession struct {
	conn       *ws.Conn
	userID     string
	validate   *validator.Validate
	msgService service.MessageService
	logger     logging.Logger
}

func (s *chatSession) Serve() {
	defer s.conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := s.msgService.NewServeSession(ctx, s.userID)
	if err != nil {
		return
	}

	go s.readMessages(session.InChannel())

	for {
		msg, ok := <-session.OutChannel()
		if !ok {
			return
		}

		payload, err := encoding.NewProtobufMessageMarshaler(msg).Marshal()
		if err != nil {
			s.logger.WithError(err).Error("An error occurred while marshaling the message")
			return
		}

		if err = s.conn.WriteMessage(ws.BinaryMessage, payload); err != nil {
			s.logger.WithError(err).Error("An error occurred while writing the message to websocket")
			return
		}
	}
}

func (s *chatSession) readMessages(inCh chan<- domain.CreateMessageDTO) {
	defer close(inCh)

	for {
		_, payload, err := s.conn.ReadMessage()
		if err != nil {
			if closeErr, ok := err.(*ws.CloseError); ok {
				s.logger.Infof("User (id=%s) closed the websocket connection (%s)", s.userID, closeErr)
				return
			}

			s.logger.WithError(err).Error("An error occurred while reading the message from websocket")

			return
		}

		var dto domain.CreateMessageDTO
		if err = encoding.NewProtobufCreateDTOMessageUnmarshaler(&dto).Unmarshal(payload); err != nil {
			s.logger.WithError(err).Debug("failed to unmarshalling the message")
			return
		}

		if err = s.validate.Struct(dto); err != nil {
			s.logger.WithError(err).Debug("message validation error")
			return
		}

		dto.ActionID = domain.MessageSendAction
		dto.SenderID = s.userID

		inCh <- dto
	}
}

type chatSessionHandler struct {
	upgrader   *ws.Upgrader
	validate   *validator.Validate
	msgService service.MessageService
	logger     logging.Logger
}

func newChatSessionHandler(msgService service.MessageService, validate *validator.Validate) *chatSessionHandler {
	return &chatSessionHandler{
		upgrader: &ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		validate:   validate,
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
		validate:   h.validate,
		msgService: h.msgService,
		userID:     domain.UserIDFromContext(req.Context()),
	}
	go chs.Serve()

	w.WriteHeader(http.StatusNoContent)
}

func (h *chatSessionHandler) respondError(w http.ResponseWriter) {
	http.Error(w, "Failed to upgrade protocol", http.StatusInternalServerError)
}
