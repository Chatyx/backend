package test

import (
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	ws "github.com/gorilla/websocket"
)

func (s *AppTestSuite) TestSendAndReceiveMessageInTheSameChat() {
	johnID := "ba566522-3305-48df-936a-73f47611934b"
	johnConn := s.newWebsocketConnection("john1967", "qwerty12345", "111")
	mickConn := s.newWebsocketConnection("mick47", "helloworld12345", "222")

	msgCh := make(chan domain.Message, 1)
	msgDTO := domain.CreateMessageDTO{
		Text:   "Hello, Mick!",
		ChatID: "609fce45-458f-477a-b2bb-e886d75d22ab",
	}

	go func() {
		s.sendWebsocketMessage(johnConn, msgDTO)
	}()
	go func() {
		msg := s.receiveWebsocketMessage(mickConn)
		msgCh <- msg
	}()

	select {
	case msg := <-msgCh:
		s.Require().Equal(domain.MessageSendAction, msg.Action)
		s.Require().Equal(msgDTO.Text, msg.Text)
		s.Require().Equal(johnID, msg.SenderID)
		s.Require().Equal(msgDTO.ChatID, msg.ChatID)
	case <-time.After(100 * time.Second):
		s.T().Error("timeout exceeded")
	}
}

func (s *AppTestSuite) sendWebsocketMessage(conn *ws.Conn, dto domain.CreateMessageDTO) {
	payload, err := encoding.NewProtobufCreateDTOMessageMarshaler(dto).Marshal()
	s.NoError(err, "Failed to marshal message")

	err = conn.WriteMessage(ws.BinaryMessage, payload)
	s.NoError(err, "Failed to send websocket message")
}

func (s *AppTestSuite) receiveWebsocketMessage(conn *ws.Conn) domain.Message {
	mt, payload, err := conn.ReadMessage()
	s.NoError(err, "Failed to receive websocket message")

	s.Require().Equal(mt, ws.BinaryMessage, "Received websocket message must be binary")

	var message domain.Message
	err = encoding.NewProtobufMessageUnmarshaler(&message).Unmarshal(payload)
	s.NoError(err, "Failed to unmarshal message")

	return message
}
