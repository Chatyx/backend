// +build integration

package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

func (s *AppTestSuite) TestChatUserList() {
	const chatID = "609fce45-458f-477a-b2bb-e886d75d22ab"

	dbUsers, err := s.getChatUsersFromDB(chatID)
	s.NoError(err, "Failed to get chat's users from database")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	uri := fmt.Sprintf("/chats/%s/users", chatID)
	req, err := http.NewRequest(http.MethodGet, s.buildURL(uri), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var responseList struct {
		Users []domain.User `json:"list"`
	}

	err = json.NewDecoder(resp.Body).Decode(&responseList)
	s.NoError(err, "Failed to decode response body")

	s.Require().Equal(
		len(dbUsers), len(responseList.Users),
		"The length of users is not equal to each other",
	)

	respUsersMap := make(map[string]domain.User, len(responseList.Users))
	for _, respUser := range responseList.Users {
		respUsersMap[respUser.ID] = respUser
	}

	for _, dbUser := range dbUsers {
		respUser, ok := respUsersMap[dbUser.ID]
		s.Require().Equalf(
			true, ok,
			"User with id = %q is not found in the response list", dbUser.ID,
		)

		s.compareUsers(dbUser, respUser)
	}
}

func (s *AppTestSuite) getChatUsersFromDB(chatID string) ([]domain.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users 
	INNER JOIN users_chats 
		ON users.id = users_chats.user_id 
	WHERE users_chats.chat_id = $1`, strings.Join(userTableColumns, ", "))

	return s.queryUsers(query, chatID)
}
