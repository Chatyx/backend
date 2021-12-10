//go:build integration
// +build integration

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var userTableColumns = []string{
	"users.id", "users.username", "users.password",
	"users.first_name", "users.last_name", "users.email",
	"users.birth_date", "users.department", "users.is_deleted",
	"users.created_at", "users.updated_at",
}

func (s *AppTestSuite) TestUserList() {
	dbUsers, err := s.getAllUsersFromDB()
	s.NoError(err, "Failed to get all users from database")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)
	req, err := http.NewRequest(http.MethodGet, s.buildURL("/users"), nil)
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

func (s *AppTestSuite) TestUserCreate() {
	var (
		username  = "test_user"
		password  = "qwerty12345"
		email     = "test_user@gmail.com"
		birthDate = "1998-06-03"
	)

	strBody := fmt.Sprintf(
		`{"username":"%s","password":"%s","email":"%s","birth_date":"%s"}`,
		username, password, email, birthDate,
	)

	req, err := http.NewRequest(http.MethodPost, s.buildURL("/users"), strings.NewReader(strBody))
	s.NoError(err, "Failed to create request")

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var respUser domain.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	s.NoError(err, "Failed to decode response body")

	expectedUser := domain.User{
		ID:        respUser.ID,
		Username:  username,
		Email:     email,
		BirthDate: birthDate,
	}

	s.Require().NotEqual("", respUser.ID, "User id can't be empty")
	s.compareUsers(expectedUser, respUser)

	dbUser, err := s.getUserFromDB(respUser.ID)
	s.NoError(err, "Failed to get user from database")

	s.NoError(
		bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)),
		"Sent password and stored hash aren't equal",
	)
	s.compareUsers(expectedUser, dbUser)
}

func (s *AppTestSuite) TestGetUser() {
	userID := "ba566522-3305-48df-936a-73f47611934b"
	dbUser, err := s.getUserFromDB(userID)
	s.NoError(err, "Failed to get user by id")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)
	req, err := http.NewRequest(http.MethodGet, s.buildURL("/users/"+userID), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var respUser domain.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	s.NoError(err, "Failed to decode response body")

	s.compareUsers(dbUser, respUser)
}

func (s *AppTestSuite) TestUpdateUser() {
	tokenPair := s.authenticate("mick47", "helloworld12345", uuid.New().String())

	expectedUser := domain.User{
		ID:         "7e7b1825-ef9a-42ec-b4db-6f09dffe3850",
		Username:   "mick79",
		Email:      "mick79@gmail.com",
		FirstName:  "Micky",
		LastName:   "Tyson",
		BirthDate:  "1979-01-02",
		Department: "HR",
	}

	payload, err := json.Marshal(expectedUser)
	s.NoError(err, "Failed to marshal user request body")

	req, err := http.NewRequest(http.MethodPut, s.buildURL("/user"), bytes.NewReader(payload))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var respUser domain.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	s.NoError(err, "Failed to decode response body")

	s.compareUsers(expectedUser, respUser)

	dbUser, err := s.getUserFromDB(expectedUser.ID)
	s.NoError(err, "Failed to get user from database")

	s.compareUsers(expectedUser, dbUser)
}

func (s *AppTestSuite) TestUpdateUserPassword() {
	userID := "7e7b1825-ef9a-42ec-b4db-6f09dffe3850"
	currPassword, newPassword := "helloworld12345", "qwerty55555"
	strBody := fmt.Sprintf(`{"current_password":"%s","new_password":"%s"}`, currPassword, newPassword)
	tokenPair := s.authenticate("mick47", currPassword, uuid.New().String())

	req, err := http.NewRequest(http.MethodPut, s.buildURL("/user/password"), strings.NewReader(strBody))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusNoContent, resp.StatusCode)

	val, err := s.redisClient.Exists(context.Background(), tokenPair.RefreshToken).Result()
	s.NoError(err, "Failed to check exist old session")
	s.Require().Equal(int64(0), val, "Old session exists after change password")

	val, err = s.redisClient.Exists(context.Background(), userID).Result()
	s.NoError(err, "Failed to check exist userID key")
	s.Require().Equal(int64(0), val, "userID key must not exist")

	s.authenticate("mick47", newPassword, uuid.New().String())
}

func (s *AppTestSuite) TestUpdateUserPasswordWrong() {
	tokenPair := s.authenticate("mick47", "helloworld12345", uuid.New().String())

	strBody := fmt.Sprintf(`{"current_password":"%s","new_password":"%s"}`, "qQq123456", "qwerty12345")
	req, err := http.NewRequest(http.MethodPut, s.buildURL("/user/password"), strings.NewReader(strBody))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *AppTestSuite) TestDeleteUser() {
	id := "ba566522-3305-48df-936a-73f47611934b"

	// Check if the user exists before delete
	_, err := s.getUserFromDB(id)
	s.NoError(err, "Failed to get user from database")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	req, err := http.NewRequest(http.MethodDelete, s.buildURL("/user"), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusNoContent, resp.StatusCode)

	// Check if the user exists after delete
	user, err := s.getUserFromDB(id)
	s.NoError(err, "Failed to get user from database")
	s.Require().True(user.IsDeleted, "User hasn't been deleted")

	// Check that refresh sessions don't exist after user delete
	sessionKey := "session:" + tokenPair.RefreshToken
	userSessionsKey := fmt.Sprintf("user:%s:sessions", id)
	val, err := s.redisClient.Exists(context.Background(), sessionKey, userSessionsKey).Result()
	s.NoError(err, "Failed to check existence the session keys")
	s.Require().Equal(int64(0), val)
}

func (s *AppTestSuite) getUserFromDB(id string) (domain.User, error) {
	var (
		user      domain.User
		birthDate pgtype.Date
	)

	query := fmt.Sprintf("SELECT %s FROM users WHERE users.id = $1", strings.Join(userTableColumns, ", "))

	row := s.dbConn.QueryRow(query, id)
	if err := row.Scan(
		&user.ID, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.Email,
		&birthDate, &user.Department, &user.IsDeleted,
		&user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		return domain.User{}, err
	}

	if birthDate.Status == pgtype.Present {
		user.BirthDate = birthDate.Time.Format("2006-01-02")
	}

	return user, nil
}

func (s *AppTestSuite) getAllUsersFromDB() ([]domain.User, error) {
	users := make([]domain.User, 0)
	query := fmt.Sprintf("SELECT %s FROM users", strings.Join(userTableColumns, ", "))

	rows, err := s.dbConn.Query(query)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			user      domain.User
			birthDate pgtype.Date
		)

		if err = rows.Scan(
			&user.ID, &user.Username, &user.Password,
			&user.FirstName, &user.LastName, &user.Email,
			&birthDate, &user.Department, &user.IsDeleted,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if birthDate.Status == pgtype.Present {
			user.BirthDate = birthDate.Time.Format("2006-01-02")
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *AppTestSuite) compareUsers(expected, actual domain.User) {
	s.Require().Equal(expected.ID, actual.ID)
	s.Require().Equal(expected.Username, actual.Username)
	s.Require().Equal(expected.Email, actual.Email)
	s.Require().Equal(expected.FirstName, actual.FirstName)
	s.Require().Equal(expected.LastName, actual.LastName)
	s.Require().Equal(expected.BirthDate, actual.BirthDate)
	s.Require().Equal(expected.Department, actual.Department)

	if expected.CreatedAt != nil {
		s.Require().Equal(expected.CreatedAt, actual.CreatedAt)
	}

	if expected.UpdatedAt != nil {
		s.Require().Equal(expected.UpdatedAt, actual.UpdatedAt)
	}
}
