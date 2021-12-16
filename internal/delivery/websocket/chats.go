package websocket

import (
	"context"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	ws "github.com/gorilla/websocket"
)

type chatSession struct {
	conn       *ws.Conn
	userID     string
	msgService service.MessageService
	logger     logging.Logger
}

func (s *chatSession) Serve() {
	defer s.conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inCh, outCh, errCh, err := s.msgService.NewServeSession(ctx, s.userID)
	if err != nil {
		return
	}

	go s.readMessages(inCh)

	for {
		select {
		case msg, ok := <-outCh:
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
		case err = <-errCh:
			s.logger.WithError(err).Error("an error occurred while serving message session")
			return
		}
	}
}

func (s *chatSession) readMessages(inCh chan<- domain.CreateMessageDTO) {
	defer close(inCh)

	for {
		_, payload, err := s.conn.ReadMessage()
		if err != nil {
			if _, ok := err.(*ws.CloseError); ok {
				s.logger.WithError(err).Info("User closed the websocket connection")
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

		if err = validator.StructValidator(dto).Validate(); err != nil {
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
	msgService service.MessageService
}

func newChatSessionHandler(msgService service.MessageService) *chatSessionHandler {
	return &chatSessionHandler{
		upgrader: &ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		msgService: msgService,
	}
}

func (h *chatSessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := logging.GetLoggerFromContext(ctx)

	conn, err := h.upgrader.Upgrade(w, req, nil)
	if err != nil {
		logger.WithError(err).Debug("failed to upgrade protocol")
		return
	}

	authUser := domain.AuthUserFromContext(ctx)
	chs := &chatSession{
		conn:       conn,
		logger:     logger,
		msgService: h.msgService,
		userID:     authUser.UserID,
	}

	go chs.Serve()

	w.WriteHeader(http.StatusNoContent)
}
