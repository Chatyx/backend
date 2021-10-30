// +build integration

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/go-redis/redis/v8"
	ws "github.com/gorilla/websocket"
)

var messageTableColumns = []string{
	"id", "action_id", "text",
	"sender_id", "chat_id", "created_at",
}

func (s *AppTestSuite) TestSendMessageViaWebsocket() {
	const (
		johnID = "ba566522-3305-48df-936a-73f47611934b"
		mickID = "7e7b1825-ef9a-42ec-b4db-6f09dffe3850"
		chatID = "609fce45-458f-477a-b2bb-e886d75d22ab"
	)

	msgCh := make(chan domain.Message, 1)
	johnConn, _ := s.newWebsocketConnection("john1967", "qwerty12345", "111")
	mickConn, _ := s.newWebsocketConnection("mick47", "helloworld12345", "222")
	defer johnConn.Close()
	defer mickConn.Close()

	go s.sendWebsocketMessage(johnConn, "Hello, Mick!", chatID)
	go func() {
		msg := s.receiveWebsocketMessage(mickConn)
		msgCh <- msg
	}()

	select {
	case msg := <-msgCh:
		s.Require().Equal(domain.MessageSendAction, msg.ActionID)
		s.Require().Equal("Hello, Mick!", msg.Text)
		s.Require().Equal(johnID, msg.SenderID)
		s.Require().Equal(chatID, msg.ChatID)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
	}

	go s.sendWebsocketMessage(mickConn, "Hi, John!", chatID)
	go func() {
		msg := s.receiveWebsocketMessage(johnConn)
		msgCh <- msg
	}()

	select {
	case msg := <-msgCh:
		s.Require().Equal(domain.MessageSendAction, msg.ActionID)
		s.Require().Equal("Hi, John!", msg.Text)
		s.Require().Equal(mickID, msg.SenderID)
		s.Require().Equal(chatID, msg.ChatID)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
	}
}

func (s *AppTestSuite) TestSendMessageViaAPI() {
	const chatID = "609fce45-458f-477a-b2bb-e886d75d22ab"

	johnTokenPair := s.authenticate("john1967", "qwerty12345", "111")
	mickConn, _ := s.newWebsocketConnection("mick47", "helloworld12345", "222")
	defer mickConn.Close()

	strBody := fmt.Sprintf(`{"text":"Hi, Mick!","chat_id":"%s"}`, chatID)
	req, err := http.NewRequest(http.MethodPost, s.buildURL("/messages"), strings.NewReader(strBody))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+johnTokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var respMessage domain.Message
	err = json.NewDecoder(resp.Body).Decode(&respMessage)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", respMessage.ID, "Message id can't be empty")

	expectedMessage := domain.Message{
		ID:       respMessage.ID,
		ActionID: domain.MessageSendAction,
		Text:     "Hi, Mick!",
		ChatID:   chatID,
		SenderID: "ba566522-3305-48df-936a-73f47611934b",
	}
	s.compareMessages(expectedMessage, respMessage)

	msgCh := make(chan domain.Message)
	go func() {
		msg := s.receiveWebsocketMessage(mickConn)
		msgCh <- msg
	}()

	select {
	case msg := <-msgCh:
		s.compareMessages(expectedMessage, msg)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
	}

	cacheMessages, err := s.getChatMessagesFromCache(chatID, time.Time{})
	s.NoError(err, "Failed to get messages for cache")

	s.Require().Equal(1, len(cacheMessages))
	s.Require().Equal(1, s.messageCountInCache(chatID))
	s.compareMessages(expectedMessage, cacheMessages[0])
}

func (s *AppTestSuite) TestSendMessageNotInChat() {
	const (
		chatID      = "92b37e8b-92e9-4c8b-a723-3a2925b62d91"
		messageText = "Hi, John. I wrote this message, but I'm not in this chat!"
	)

	conn, tokenPair := s.newWebsocketConnection("mick47", "helloworld12345", "222")
	defer conn.Close()

	s.sendWebsocketMessage(conn, messageText, chatID)

	errCh := make(chan error)
	go func() {
		_, _, err := conn.ReadMessage()
		errCh <- err
	}()

	select {
	case err := <-errCh:
		s.Require().IsType(&ws.CloseError{}, err)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
	}

	strBody := fmt.Sprintf(`{"text":"%s","chat_id":"%s"}`, messageText, chatID)
	req, err := http.NewRequest(http.MethodPost, s.buildURL("/messages"), strings.NewReader(strBody))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusNotFound, resp.StatusCode)

	s.Require().Equal(0, s.messageCountInCache(chatID))
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

	mickConn, _ := s.newWebsocketConnection("mick47", "helloworld12345", "222")
	defer mickConn.Close()

	beginSendMessages := time.Now()
	for i := 0; i < sendMessageLen; i++ {
		s.sendWebsocketMessage(mickConn, "Hi, "+strconv.Itoa(i), chatID)
		time.Sleep(1 * time.Millisecond)
	}
	time.Sleep(20 * time.Millisecond)

	expectedCachedMessages, err := s.getChatMessagesFromCache(chatID, beginSendMessages)
	s.NoError(err, "Failed to get chat's messages from cache")

	cachedMessages := s.getRequestMessages(chatID, beginSendMessages, headers)

	s.Require().Equal(sendMessageLen, len(cachedMessages))
	s.Require().Equal(sendMessageLen, s.messageCountInCache(chatID))
	s.Require().Equal(expectedCachedMessages, cachedMessages)
}

func (s *AppTestSuite) TestMessagesAfterChatDelete() {
	const chatID = "609fce45-458f-477a-b2bb-e886d75d22ab"

	johnConn, _ := s.newWebsocketConnection("john1967", "qwerty12345", "111")
	defer johnConn.Close()

	mickTokenPair := s.authenticate("mick47", "helloworld12345", "222")

	s.sendWebsocketMessage(johnConn, "Hi, Mick!", chatID)
	time.Sleep(50 * time.Millisecond)

	req, err := http.NewRequest(http.MethodDelete, s.buildURL("/chats/"+chatID), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+mickTokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusNoContent, resp.StatusCode)

	s.sendWebsocketMessage(johnConn, "Hi, Mick!", chatID)

	errCh := make(chan error)
	go func() {
		_, _, err = johnConn.ReadMessage()
		errCh <- err
	}()

	select {
	case err = <-errCh:
		s.Require().IsType(&ws.CloseError{}, err)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
	}

	s.Require().Equal(1, s.messageCountInCache(chatID))
}

func (s *AppTestSuite) TestGoroutinesLeak() {
	const numConnections = 100

	beforeGoroNum := runtime.NumGoroutine()
	connList := make([]*ws.Conn, 0, numConnections)

	for i := 0; i < numConnections; i++ {
		conn, _ := s.newWebsocketConnection("john1967", "qwerty12345", "111")
		connList = append(connList, conn)
		time.Sleep(5 * time.Millisecond)
	}

	for _, conn := range connList[:numConnections/2] {
		s.NoError(conn.Close(), "Failed to close websocket connection")
	}

	for _, conn := range connList[numConnections/2:] {
		err := conn.WriteMessage(ws.BinaryMessage, []byte("Hello, world"))
		s.NoError(err, "Failed to send message")
	}

	time.Sleep(20 * time.Second)

	s.Require().LessOrEqualf(runtime.NumGoroutine(), beforeGoroNum, "Goroutines leak detected")
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
			&message.ID, &message.ActionID, &message.Text,
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

func (s *AppTestSuite) getChatMessagesFromCache(chatID string, timestamp time.Time) ([]domain.Message, error) {
	key := fmt.Sprintf("chat:%s:messages", chatID)

	payloads, err := s.redisClient.ZRangeByScore(context.Background(), key, &redis.ZRangeBy{
		Min: fmt.Sprintf("(%d", timestamp.UnixNano()),
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, 0, len(payloads))

	for _, payload := range payloads {
		var message domain.Message

		if err = encoding.NewProtobufMessageUnmarshaler(&message).Unmarshal([]byte(payload)); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (s *AppTestSuite) messageCountInCache(chatID string) int {
	key := fmt.Sprintf("chat:%s:messages", chatID)

	val, err := s.redisClient.ZCount(context.Background(), key, "-inf", "+inf").Result()
	s.NoError(err, "Failed to get count messages in the redis")

	return int(val)
}

func (s *AppTestSuite) compareMessages(expected, actual domain.Message) {
	s.Require().Equal(expected.ID, actual.ID)
	s.Require().Equal(expected.ActionID, actual.ActionID)
	s.Require().Equal(expected.Text, actual.Text)
	s.Require().Equal(expected.ChatID, actual.ChatID)
	s.Require().Equal(expected.SenderID, actual.SenderID)

	if expected.CreatedAt != nil {
		s.Require().Equal(expected.CreatedAt, actual.CreatedAt)
	}
}
