// +build integration

package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
