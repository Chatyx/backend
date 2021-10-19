// +build integration

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	ws "github.com/gorilla/websocket"
)

var messageTableColumns = []string{
	"action", "text",
	"sender_id", "chat_id", "created_at",
}

func (s *AppTestSuite) TestSendAndReceiveMessageInTheSameChat() {
	const (
		johnID      = "ba566522-3305-48df-936a-73f47611934b"
		chatID      = "609fce45-458f-477a-b2bb-e886d75d22ab"
		messageText = "Hello, Mick!"
	)

	msgCh := make(chan domain.Message, 1)
	johnConn := s.newWebsocketConnection("john1967", "qwerty12345", "111")
	mickConn := s.newWebsocketConnection("mick47", "helloworld12345", "222")

	go func() {
		s.sendWebsocketMessage(johnConn, messageText, chatID)
	}()
	go func() {
		msg := s.receiveWebsocketMessage(mickConn)
		msgCh <- msg
	}()

	select {
	case msg := <-msgCh:
		s.Require().Equal(domain.MessageSendAction, msg.Action)
		s.Require().Equal(messageText, msg.Text)
		s.Require().Equal(johnID, msg.SenderID)
		s.Require().Equal(chatID, msg.ChatID)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
	}
}

func (s *AppTestSuite) TestMessageList() {
	const (
		sendMessageLen = 100
		chatID         = "609fce45-458f-477a-b2bb-e886d75d22ab"
	)

	expectedStoredMessages, err := s.getChatMessagesFromDB(chatID, time.Time{})
	s.NoError(err, "Failed to get chat's messages from database")

	johnTokenPair := s.authenticate("john1967", "qwerty12345", "111")

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+johnTokenPair.AccessToken)

	storedMessages := s.getRequestMessages(chatID, time.Time{}, headers)
	s.Require().Equal(expectedStoredMessages, storedMessages)

	mickConn := s.newWebsocketConnection("mick47", "helloworld12345", "222")

	beginSendMessages := time.Now()
	for i := 0; i < sendMessageLen; i++ {
		s.sendWebsocketMessage(mickConn, "Hi, "+strconv.Itoa(i), chatID)
		time.Sleep(1 * time.Millisecond)
	}

	cachedMessages := s.getRequestMessages(chatID, beginSendMessages, headers)

	s.Require().Equal(sendMessageLen, len(cachedMessages))
	s.Require().Equal(sendMessageLen, s.messageCountInCache(chatID))
}

func (s *AppTestSuite) sendWebsocketMessage(conn *ws.Conn, text, chatID string) {
	dto := domain.CreateMessageDTO{
		Text:   text,
		ChatID: chatID,
	}

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

func (s *AppTestSuite) getRequestMessages(chatID string, timestamp time.Time, headers http.Header) []domain.Message {
	uri := fmt.Sprintf(
		"/chats/%s/messages?timestamp=%s",
		chatID, url.QueryEscape(timestamp.Format(time.RFC3339Nano)),
	)

	req, err := http.NewRequest(http.MethodGet, s.buildURL(uri), nil)
	s.NoError(err, "Failed to create get chat's messages request")

	req.Header = headers

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var responseList struct {
		Messages []domain.Message `json:"list"`
	}

	err = json.NewDecoder(resp.Body).Decode(&responseList)
	s.NoError(err, "Failed to decode response body")

	return responseList.Messages
}

func (s *AppTestSuite) getChatMessagesFromDB(chatID string, timestamp time.Time) ([]domain.Message, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM messages WHERE chat_id = $1 AND created_at >= $2 ORDER BY created_at",
		strings.Join(messageTableColumns, ", "),
	)

	rows, err := s.dbConn.Query(query, chatID, timestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)

	for rows.Next() {
		var message domain.Message

		if err = rows.Scan(
			&message.Action, &message.Text,
			&message.SenderID, &message.ChatID, &message.CreatedAt,
		); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *AppTestSuite) messageCountInCache(chatID string) int {
	key := fmt.Sprintf("chat:%s:messages", chatID)

	val, err := s.redisClient.ZCount(context.Background(), key, "-inf", "+inf").Result()
	s.NoError(err, "Failed to get count messages in the redis")

	return int(val)
}
