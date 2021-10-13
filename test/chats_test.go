// +build integration

package test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

var chatTableColumns = []string{
	"id", "name", "description",
	"creator_id", "created_at", "updated_at",
}

func (s *AppTestSuite) TestChatList() {
	userID := "ba566522-3305-48df-936a-73f47611934b"
	dbChats, err := s.getUserChatsFromDB(userID)
	s.NoError(err, "Failed to get user's chats from database")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)
	req, err := http.NewRequest(http.MethodGet, s.buildURL("/chats"), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var responseList struct {
		Chats []domain.Chat `json:"list"`
	}

	err = json.NewDecoder(resp.Body).Decode(&responseList)
	s.NoError(err, "Failed to decode response body")

	s.Require().Equal(
		len(dbChats), len(responseList.Chats),
		"The length of chats is not equal to each other",
	)

	respChatsMap := make(map[string]domain.Chat, len(responseList.Chats))
	for _, respChat := range responseList.Chats {
		respChatsMap[respChat.ID] = respChat
	}

	for _, dbChat := range dbChats {
		respChat, ok := respChatsMap[dbChat.ID]
		s.Require().Equalf(
			true, ok,
			"Chat with id = %q is not found in the response list", dbChat.ID,
		)

		s.compareChats(dbChat, respChat)
	}
}

func (s *AppTestSuite) TestChatCreate() {
	var (
		name        = "John's test created chat"
		description = "This is a chat description"
		userID      = "ba566522-3305-48df-936a-73f47611934b"
	)

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	strBody := fmt.Sprintf(`{"name":"%s","description":"%s"}`, name, description)

	req, err := http.NewRequest(http.MethodPost, s.buildURL("/chats"), strings.NewReader(strBody))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var respChat domain.Chat
	err = json.NewDecoder(resp.Body).Decode(&respChat)
	s.NoError(err, "Failed to decode response body")

	expectedChat := domain.Chat{
		ID:          respChat.ID,
		Name:        name,
		Description: description,
		CreatorID:   userID,
	}

	s.Require().NotEqual("", respChat.ID, "Chat id can't be empty")
	s.compareChats(expectedChat, respChat)

	dbChat, err := s.getUserChatFromDB(respChat.ID, userID)
	s.NoError(err, "Failed to get chat by id")

	s.compareChats(expectedChat, dbChat)
}

func (s *AppTestSuite) TestChatGet() {
	userID, chatID := "ba566522-3305-48df-936a-73f47611934b", "609fce45-458f-477a-b2bb-e886d75d22ab"

	dbChat, err := s.getUserChatFromDB(chatID, userID)
	s.NoError(err, "Failed to get chat by id")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)
	req, err := http.NewRequest(http.MethodGet, s.buildURL("/chats/"+chatID), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var respChat domain.Chat
	err = json.NewDecoder(resp.Body).Decode(&respChat)
	s.NoError(err, "Failed to decode response body")

	s.compareChats(dbChat, respChat)
}

func (s *AppTestSuite) TestChatUpdate() {
	var (
		name        = "Updated chat"
		description = "Updated description"
		chatID      = "92b37e8b-92e9-4c8b-a723-3a2925b62d91"
		userID      = "ba566522-3305-48df-936a-73f47611934b"
	)

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	reqBody := fmt.Sprintf(`{"name":"%s","description":"%s"}`, name, description)
	req, err := http.NewRequest(http.MethodPut, s.buildURL("/chats/"+chatID), strings.NewReader(reqBody))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var respChat domain.Chat
	err = json.NewDecoder(resp.Body).Decode(&respChat)
	s.NoError(err, "Failed to decode response body")

	expectedChat := domain.Chat{
		ID:          chatID,
		Name:        name,
		Description: description,
		CreatorID:   userID,
	}

	s.compareChats(expectedChat, respChat)

	dbChat, err := s.getUserChatFromDB(chatID, userID)
	s.NoError(err, "Failed to get chat from database")

	s.compareChats(expectedChat, dbChat)
}

func (s *AppTestSuite) TestChatDelete() {
	userID, chatID := "ba566522-3305-48df-936a-73f47611934b", "92b37e8b-92e9-4c8b-a723-3a2925b62d91"

	// Check if the chat exists before delete
	_, err := s.getUserChatFromDB(chatID, userID)
	s.NoError(err, "Failed to get chat from database")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	req, err := http.NewRequest(http.MethodDelete, s.buildURL("/chats/"+chatID), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusNoContent, resp.StatusCode)

	// Check if the chat doesn't exist after delete
	_, err = s.getUserChatFromDB(chatID, userID)
	s.Require().Equal(sql.ErrNoRows, err, "Chat hasn't been deleted")
}

func (s *AppTestSuite) TestChatNoPermittedAction() {
	chatID := "609fce45-458f-477a-b2bb-e886d75d22ab"
	testTable := []struct {
		name                 string
		uri                  string
		method               string
		reqBody              io.Reader
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			name:                 "Update chat, which doesn't belong to authenticated user",
			uri:                  "/chats/" + chatID,
			method:               http.MethodPut,
			reqBody:              strings.NewReader(`{"name":"Updated chat"}`),
			expectedStatus:       http.StatusNotFound,
			expectedResponseBody: `{"message":"chat is not found"}`,
		},
		{
			name:                 "Delete chat, which doesn't belong to authenticated user",
			uri:                  "/chats/" + chatID,
			method:               http.MethodDelete,
			expectedStatus:       http.StatusNotFound,
			expectedResponseBody: `{"message":"chat is not found"}`,
		},
	}

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	for _, testCase := range testTable {
		s.Run(testCase.name, func() {
			req, err := http.NewRequest(testCase.method, s.buildURL(testCase.uri), testCase.reqBody)
			s.NoError(err, "Failed to create request")

			req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

			resp, err := s.httpClient.Do(req)
			s.Require().NoError(err, "Failed to send request")

			defer resp.Body.Close()
			s.Require().Equal(testCase.expectedStatus, resp.StatusCode)

			respBody, err := ioutil.ReadAll(resp.Body)
			s.NoError(err, "Failed to read response body")

			s.Require().Equal(testCase.expectedResponseBody, string(respBody))
		})
	}
}

func (s *AppTestSuite) getUserChatFromDB(chatID, userID string) (domain.Chat, error) {
	var chat domain.Chat

	query := fmt.Sprintf(`SELECT %s 
	FROM chats 
	INNER JOIN users_chats 
		ON chats.id = users_chats.chat_id
	WHERE chats.id = $1 AND users_chats.user_id = $2`,
		strings.Join(chatTableColumns, ", "),
	)

	row := s.dbConn.QueryRow(query, chatID, userID)
	if err := row.Scan(
		&chat.ID, &chat.Name, &chat.Description,
		&chat.CreatorID, &chat.CreatedAt, &chat.UpdatedAt,
	); err != nil {
		return domain.Chat{}, err
	}

	return chat, nil
}

func (s *AppTestSuite) getUserChatsFromDB(userID string) ([]domain.Chat, error) {
	chats := make([]domain.Chat, 0)

	query := fmt.Sprintf(`SELECT %s 
	FROM chats 
	INNER JOIN users_chats 
		ON chats.id = users_chats.chat_id
	WHERE users_chats.user_id = $1`,
		strings.Join(chatTableColumns, ", "),
	)

	rows, err := s.dbConn.Query(query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var chat domain.Chat

		if err = rows.Scan(
			&chat.ID, &chat.Name, &chat.Description,
			&chat.CreatorID, &chat.CreatedAt, &chat.UpdatedAt,
		); err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (s *AppTestSuite) compareChats(expected, actual domain.Chat) {
	s.Require().Equal(expected.ID, actual.ID)
	s.Require().Equal(expected.Name, actual.Name)
	s.Require().Equal(expected.Description, actual.Description)
	s.Require().Equal(expected.CreatorID, actual.CreatorID)

	if expected.CreatedAt != nil {
		s.Require().Equal(expected.CreatedAt, actual.CreatedAt)
	}

	if expected.UpdatedAt != nil {
		s.Require().Equal(expected.UpdatedAt, actual.UpdatedAt)
	}
}
