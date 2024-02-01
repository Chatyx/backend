package websocket

import (
	"context"
	"fmt"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/internal/transport/websocket/model"
	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/log"
	ws "github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

const (
	writeTimeout = 15 * time.Second
	pongTimeout  = 60 * time.Second
	pingInterval = 30 * time.Second

	maxMessageSize = 8 * 1024
)

type MessageServeManager interface {
	BeginServe(ctx context.Context, inCh <-chan dto.MessageCreate) (<-chan entity.Message, <-chan error, error)
}

//go:generate protoc --go_out=./model ./model/message.proto
type ClientSession struct {
	userID  ctxutil.UserID
	logger  *log.Logger
	conn    *ws.Conn
	manager MessageServeManager
}

func (s *ClientSession) Serve() {
	defer func() {
		s.conn.Close()
		s.logger.Info("Client session was finished")
	}()

	ctx := log.WithLogger(ctxutil.WithUserID(context.Background(), s.userID), s.logger)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	inCh := make(chan dto.MessageCreate)
	outCh, errCh, err := s.manager.BeginServe(ctx, inCh)
	if err != nil {
		s.logger.WithError(err).Error("Failed to begin serving messages")
		return
	}

	s.logger.Info("Started client session for handling messages")

	go s.readMessages(inCh)
	s.writeMessages(outCh, errCh)
}

func (s *ClientSession) readMessages(inCh chan<- dto.MessageCreate) {
	defer close(inCh)

	s.conn.SetReadLimit(maxMessageSize)
	if err := s.conn.SetReadDeadline(time.Now().Add(pongTimeout)); err != nil {
		s.logger.WithError(err).Error("Failed to set read deadline")
		return
	}
	s.conn.SetPongHandler(func(string) error {
		if err := s.conn.SetReadDeadline(time.Now().Add(pongTimeout)); err != nil {
			s.logger.WithError(err).Error("Failed to set read deadline")
			return fmt.Errorf("set read deadline: %w", err)
		}
		return nil
	})

	for {
		_, payload, err := s.conn.ReadMessage()
		if err != nil {
			if ws.IsCloseError(err, ws.CloseNormalClosure, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				s.logger.Info("User closed websocket connection")
				return
			}

			s.logger.WithError(err).Error("Failed to read message from websocket")
			return
		}

		messageCreateModel := &model.MessageCreate{}
		if err = proto.Unmarshal(payload, messageCreateModel); err != nil {
			s.logger.WithError(err).Debug("Failed to unmarshal message create body")
			return
		}

		inCh <- messageCreateModel.DTO()
	}
}

func (s *ClientSession) writeMessages(outCh <-chan entity.Message, errCh <-chan error) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-outCh:
			if !ok {
				return
			}

			messageModel := model.NewMessageFromEntity(message)
			payload, err := proto.Marshal(messageModel)
			if err != nil {
				s.logger.WithError(err).Error("Failed to marshal message")
				return
			}

			if err = s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				s.logger.WithError(err).Error("Failed to set write deadline")
				return
			}
			if err = s.conn.WriteMessage(ws.BinaryMessage, payload); err != nil {
				s.logger.WithError(err).Error("Failed to write message")
				return
			}
		case err, ok := <-errCh:
			if !ok {
				return
			}

			s.logger.WithError(err).Error("Error while serving messages")
			return
		case <-ticker.C:
			if err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				s.logger.WithError(err).Error("Failed to set write deadline")
				return
			}
			if err := s.conn.WriteMessage(ws.PingMessage, nil); err != nil {
				s.logger.WithError(err).Error("Failed to write ping message")
				return
			}
		}
	}
}
