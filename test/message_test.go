package test

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/internal/transport/websocket/model"

	ws "github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func (s *AppTestSuite) TestMessagesInGroupViaWebsocket() {
	johnPair, err := s.authenticate(johnUsername)
	s.Require().NoError(err, "Failed to authenticate")

	johnConn, err := s.newWebsocketConn(johnPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer johnConn.Close()

	mickPair, err := s.authenticate(mickUsername)
	s.Require().NoError(err, "Failed to authenticate")

	mickConn, err := s.newWebsocketConn(mickPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer mickConn.Close()

	time.Sleep(delay)

	chatID := entity.ChatID{ID: 2, Type: entity.GroupChatType}
	s.testCommunicationViaWebsocket(
		johnConn, mickConn, johnUserID, chatID,
		"Hello, Mick!",
		"How are you?",
		"Are you here?",
	)
}

func (s *AppTestSuite) TestMessagesInDialogViaWebsocket() {
	johnPair, err := s.authenticate(johnUsername)
	s.Require().NoError(err, "Failed to authenticate")

	johnConn, err := s.newWebsocketConn(johnPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer johnConn.Close()

	mickPair, err := s.authenticate(mickUsername)
	s.Require().NoError(err, "Failed to authenticate")

	mickConn, err := s.newWebsocketConn(mickPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer mickConn.Close()

	time.Sleep(delay)

	chatID := entity.ChatID{ID: 1, Type: entity.DialogChatType}
	s.testCommunicationViaWebsocket(
		johnConn, mickConn, johnUserID, chatID,
		"Hello, Mick!",
		"How are you?",
		"Are you here?",
	)
}

func (s *AppTestSuite) TestMessagesNotInGroupViaWebsocket() {
	johnPair, err := s.authenticate(johnUsername)
	s.Require().NoError(err, "Failed to authenticate")

	johnConn, err := s.newWebsocketConn(johnPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer johnConn.Close()

	jacobPair, err := s.authenticate(jacobUsername)
	s.Require().NoError(err, "Failed to authenticate")

	jacobConn, err := s.newWebsocketConn(jacobPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer jacobConn.Close()

	time.Sleep(delay)

	chatID := entity.ChatID{ID: 2, Type: entity.GroupChatType}
	s.testCommunicationIfSenderNotInChat(jacobConn, johnConn, chatID)
}

func (s *AppTestSuite) TestMessagesNotInDialogViaWebsocket() {
	johnPair, err := s.authenticate(johnUsername)
	s.Require().NoError(err, "Failed to authenticate")

	johnConn, err := s.newWebsocketConn(johnPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer johnConn.Close()

	jacobPair, err := s.authenticate(jacobUsername)
	s.Require().NoError(err, "Failed to authenticate")

	jacobConn, err := s.newWebsocketConn(jacobPair.AccessToken)
	s.Require().NoError(err, "Failed to open websocket connection")
	defer jacobConn.Close()

	time.Sleep(delay)

	chatID := entity.ChatID{ID: 1, Type: entity.DialogChatType}
	s.testCommunicationIfSenderNotInChat(jacobConn, johnConn, chatID)
}

func (s *AppTestSuite) TestGoroutinesLeak() {
	const numConnections = 100

	beforeGoroNum := runtime.NumGoroutine()
	connList := make([]*ws.Conn, 0, numConnections)

	johnPair, authErr := s.authenticate(johnUsername)
	s.Require().NoError(authErr, "Failed to authenticate")

	chatID := entity.ChatID{ID: 1, Type: entity.DialogChatType}

	for i := 0; i < numConnections; i++ {
		johnConn, err := s.newWebsocketConn(johnPair.AccessToken)
		s.Require().NoError(err, "Failed to open websocket connection")

		err = s.sendMessageViaWebsocket(johnConn, chatID, "Hello, world!")
		s.Require().NoError(err, "Failed to send message via websocket")

		connList = append(connList, johnConn)
		time.Sleep(5 * time.Millisecond)
	}

	for _, conn := range connList {
		s.NoError(conn.Close(), "Failed to close websocket connection")
	}

	time.Sleep(20 * time.Second)

	s.LessOrEqualf(runtime.NumGoroutine(), beforeGoroNum, "Goroutines leak detected")
}

func (s *AppTestSuite) testCommunicationViaWebsocket(sendConn, recConn *ws.Conn, senderID int, chatID entity.ChatID, texts ...string) {
	beginAt := time.Now()
	sendMessages := make([]entity.Message, len(texts))
	for i, text := range texts {
		message := entity.Message{
			ChatID:      chatID,
			SenderID:    senderID,
			Content:     text,
			ContentType: entity.TextContentType,
			SentAt:      time.Now(),
		}

		sendMessages[i] = message
	}

	errCh := make(chan error)
	msgCh := make(chan entity.Message)

	go func() {
		for _, message := range sendMessages {
			if err := s.sendMessageViaWebsocket(sendConn, message.ChatID, message.Content); err != nil {
				errCh <- fmt.Errorf("send message: %w", err)
			}
		}
	}()
	go func() {
		for range sendMessages {
			msg, err := s.receiveMessageViaWebsocket(recConn)
			if err != nil {
				errCh <- fmt.Errorf("receive message: %w", err)
				continue
			}

			msgCh <- msg
		}
	}()

	timer := time.NewTimer(receiveMessageTimeout)
	defer timer.Stop()

	for _, expected := range sendMessages {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(receiveMessageTimeout)

		select {
		case got := <-msgCh:
			s.testCompareMessage(expected, got)
		case err := <-errCh:
			s.Require().NoError(err)
		case <-timer.C:
			s.T().Error("timeout exceeded")
		}
	}

	s.testSavingMessage(chatID, beginAt, sendMessages)
}

func (s *AppTestSuite) testCommunicationIfSenderNotInChat(sendConn, recConn *ws.Conn, chatID entity.ChatID) {
	errCh := make(chan error)
	msgCh := make(chan entity.Message)

	go func() {
		if err := s.sendMessageViaWebsocket(sendConn, chatID, "Hello! I'm not in this chat"); err != nil {
			errCh <- fmt.Errorf("send message: %w", err)
			return
		}
		if _, err := s.receiveMessageViaWebsocket(sendConn); err != nil {
			errCh <- fmt.Errorf("receve message: %w", err)
		}
	}()
	go func() {
		msg, err := s.receiveMessageViaWebsocket(recConn)
		if err != nil {
			errCh <- fmt.Errorf("receive message: %w", err)
			return
		}

		msgCh <- msg
	}()

	select {
	case <-msgCh:
		s.T().Error("Got a message, but not expected")
	case <-time.After(receiveMessageTimeout):
	}

	select {
	case err := <-errCh:
		closeErr := &ws.CloseError{}
		s.ErrorAs(err, &closeErr, "Unexpected error")
	case <-time.After(receiveMessageTimeout):
		s.T().Errorf("timeout exceeded, while waiting close error")
	}
}

func (s *AppTestSuite) sendMessageViaWebsocket(conn *ws.Conn, chatID entity.ChatID, text string) error {
	var chatType model.ChatType
	switch chatID.Type {
	case entity.DialogChatType:
		chatType = model.ChatType_DIALOG
	case entity.GroupChatType:
		chatType = model.ChatType_GROUP
	}

	msg := &model.MessageCreate{
		ChatId:   int64(chatID.ID),
		ChatType: chatType,
		Content:  text,
	}
	payload, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = conn.WriteMessage(ws.BinaryMessage, payload)
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (s *AppTestSuite) receiveMessageViaWebsocket(conn *ws.Conn) (entity.Message, error) {
	_, payload, err := conn.ReadMessage()
	if err != nil {
		return entity.Message{}, fmt.Errorf("read message: %w", err)
	}

	msg := &model.Message{}
	if err = proto.Unmarshal(payload, msg); err != nil {
		return entity.Message{}, fmt.Errorf("unmarshal message: %w", err)
	}

	var chatType entity.ChatType
	switch msg.ChatType {
	case model.ChatType_DIALOG:
		chatType = entity.DialogChatType
	case model.ChatType_GROUP:
		chatType = entity.GroupChatType
	}

	var contentType entity.ContentType
	switch msg.ContentType {
	case model.ContentType_TEXT:
		contentType = entity.TextContentType
	case model.ContentType_IMAGE:
		contentType = entity.ImageContentType
	}

	var deliveredAt *time.Time
	if msg.Delivered != nil {
		t := msg.Delivered.AsTime()
		deliveredAt = &t
	}

	return entity.Message{
		ID: int(msg.Id),
		ChatID: entity.ChatID{
			ID:   int(msg.ChatId),
			Type: chatType,
		},
		SenderID:    int(msg.SenderId),
		Content:     msg.Content,
		ContentType: contentType,
		IsService:   msg.IsService,
		SentAt:      msg.SentAt.AsTime(),
		DeliveredAt: deliveredAt,
	}, nil
}

func (s *AppTestSuite) testSavingMessage(chatID entity.ChatID, beginAt time.Time, expectedMessages []entity.Message) {
	if len(expectedMessages) == 0 {
		return
	}

	dbMessages, err := s.listMessagesFromDB(chatID, beginAt)
	s.Require().NoError(err, "Failed to list messages from db")
	s.Require().Equal(len(expectedMessages), len(dbMessages), "Len of expected and got messages aren't equal")

	for i, expected := range expectedMessages {
		got := dbMessages[i]
		s.testCompareMessage(expected, got)
	}
}

func (s *AppTestSuite) listMessagesFromDB(chatID entity.ChatID, beginAt time.Time) ([]entity.Message, error) {
	query := `SELECT id, sender_id, chat_id, chat_type, 
	content, content_type, is_service, 
	sent_at, delivered_at
	FROM messages 
	WHERE chat_id = $1 AND chat_type = $2 AND sent_at >= $3
	ORDER BY sent_at`

	rows, err := s.db.Query(query, chatID.ID, chatID.Type, beginAt)
	if err != nil {
		return nil, fmt.Errorf("exec query to select messages: %w", err)
	}
	defer rows.Close()

	var messages []entity.Message

	for rows.Next() {
		var message entity.Message

		if err = rows.Scan(
			&message.ID, &message.SenderID, &message.ChatID.ID, &message.ChatID.Type,
			&message.Content, &message.ContentType, &message.IsService,
			&message.SentAt, &message.DeliveredAt,
		); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("reading message rows: %w", err)
	}

	return messages, nil
}

func (s *AppTestSuite) testCompareMessage(expected, got entity.Message) {
	s.Equal(expected.ChatID, got.ChatID)
	s.Equal(expected.SenderID, got.SenderID)
	s.Equal(expected.Content, got.Content)
	s.Equal(expected.ContentType, got.ContentType)
	s.Equal(expected.IsService, got.IsService)
	s.Greater(got.SentAt, expected.SentAt)
	s.Less(got.SentAt, time.Now())
}

func (s *AppTestSuite) newWebsocketConn(accessToken string) (*ws.Conn, error) {
	dialer := &ws.Dialer{}

	reqHeaders := http.Header{}
	reqHeaders.Add("Authorization", "Bearer "+accessToken)

	conn, _, err := dialer.Dial(s.chatURL().String(), reqHeaders)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
