// +build integration

package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

var chatMembersTableColumns = []string{
	"users.username", "chat_members.status_id",
	"chat_members.user_id", "chat_members.chat_id",
}

func (s *AppTestSuite) TestChatMembersList() {
	const chatID = "609fce45-458f-477a-b2bb-e886d75d22ab"

	dbMembers, err := s.getChatMembersFromDB(chatID)
	s.NoError(err, "Failed to get chat members from database")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	uri := fmt.Sprintf("/chats/%s/members", chatID)
	req, err := http.NewRequest(http.MethodGet, s.buildURL(uri), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var responseList struct {
		Members []domain.ChatMember `json:"list"`
	}

	err = json.NewDecoder(resp.Body).Decode(&responseList)
	s.NoError(err, "Failed to decode response body")

	s.Require().Equal(
		len(dbMembers), len(responseList.Members),
		"The length of users is not equal to each other",
	)

	respMembersMap := make(map[string]domain.ChatMember, len(responseList.Members))
	for _, respUser := range responseList.Members {
		respMembersMap[respUser.UserID] = respUser
	}

	for _, dbMember := range dbMembers {
		respUser, ok := respMembersMap[dbMember.UserID]
		s.Require().Equalf(
			true, ok,
			"User with id = %q is not found in the response list", dbMember.UserID,
		)

		s.compareChatMembers(dbMember, respUser)
	}
}

func (s *AppTestSuite) TestChatMemberJoin() {
	const (
		chatID       = "92b37e8b-92e9-4c8b-a723-3a2925b62d91"
		mickUserID   = "7e7b1825-ef9a-42ec-b4db-6f09dffe3850"
		johnUserID   = "ba566522-3305-48df-936a-73f47611934b"
		mickUsername = "mick47"
	)

	msgCh := make(chan domain.Message)

	johnConn, johnTokenPair := s.newWebsocketConnection("john1967", "qwerty12345", "111")
	mickConn, _ := s.newWebsocketConnection(mickUsername, "helloworld12345", "222")
	defer johnConn.Close()
	defer mickConn.Close()

	time.Sleep(50 * time.Millisecond)

	go func() {
		msgCh <- s.receiveWebsocketMessage(mickConn)
	}()

	uri := fmt.Sprintf("/chats/%s/members?user_id=%s", chatID, url.QueryEscape(mickUserID))
	req, err := http.NewRequest(http.MethodPost, s.buildURL(uri), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+johnTokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusNoContent, resp.StatusCode)

	expectedJoinMessage := domain.Message{
		ActionID: domain.MessageJoinAction,
		Text:     fmt.Sprintf("%s successfully joined to the chat", mickUsername),
		ChatID:   chatID,
		SenderID: mickUserID,
	}

	select {
	case msg := <-msgCh:
		expectedJoinMessage.ID = msg.ID
		s.compareMessages(expectedJoinMessage, msg)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
		return
	}

	go func() {
		msgCh <- s.receiveWebsocketMessage(mickConn)
	}()
	s.sendWebsocketMessage(johnConn, "Hi, Mick!", chatID)

	expectedReceivedMessage := domain.Message{
		ActionID: domain.MessageSendAction,
		Text:     "Hi, Mick!",
		ChatID:   chatID,
		SenderID: johnUserID,
	}

	select {
	case msg := <-msgCh:
		expectedReceivedMessage.ID = msg.ID
		s.compareMessages(expectedReceivedMessage, msg)
	case <-time.After(50 * time.Millisecond):
		s.T().Error("timeout exceeded")
	}
}

func (s *AppTestSuite) getChatMembersFromDB(chatID string) ([]domain.ChatMember, error) {
	members := make([]domain.ChatMember, 0)
	query := fmt.Sprintf(`SELECT %s FROM users 
	INNER JOIN chat_members 
		ON users.id = chat_members.user_id
	WHERE chat_members.chat_id = $1 AND chat_members.status_id = 1`,
		strings.Join(chatMembersTableColumns, ", "))

	rows, err := s.dbConn.Query(query, chatID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var member domain.ChatMember

		if err = rows.Scan(
			&member.Username, &member.StatusID,
			&member.UserID, &member.ChatID,
		); err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (s *AppTestSuite) compareChatMembers(expected, actual domain.ChatMember) {
	s.Require().Equal(expected.Username, actual.Username)
	s.Require().Equal(expected.UserID, actual.UserID)
	s.Require().Equal(expected.ChatID, actual.ChatID)
}
