package test

import (
	"bytes"
	"database/sql"
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
	"id", "username", "password",
	"first_name", "last_name", "email",
	"birth_date", "department", "is_deleted",
	"created_at", "updated_at",
}

func (s *AppTestSuite) TestUserList() {
	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	req, err := http.NewRequest(http.MethodGet, s.buildURL("/users"), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.NoError(err, "Failed to send request")

	defer func() { _ = resp.Body.Close() }()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var responseList struct {
		Users []domain.User `json:"list"`
	}

	err = json.NewDecoder(resp.Body).Decode(&responseList)
	s.NoError(err, "Failed to decode response body")

	dbUsers, err := s.getAllUsersFromDB()
	s.NoError(err, "Failed to get all users from database")

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
	username, password, email, birthDate := "test_user", "qwerty12345", "test_user@gmail.com", "1998-06-03"
	strBody := fmt.Sprintf(
		`{"username":"%s","password":"%s","email":"%s","birth_date":"%s"}`,
		username, password, email, birthDate,
	)

	req, err := http.NewRequest(http.MethodPost, s.buildURL("/users"), strings.NewReader(strBody))
	s.NoError(err, "Failed to create request")

	resp, err := s.httpClient.Do(req)
	s.NoError(err, "Failed to send request")

	defer func() { _ = resp.Body.Close() }()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var respUser domain.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	s.NoError(err, "Failed to decode response body")

	s.Require().Equal(username, respUser.Username)
	s.Require().Equal(email, respUser.Email)
	s.Require().Equal(birthDate, respUser.BirthDate)
	s.Require().NotEqual("", respUser.ID, "User id can't be empty")

	dbUser, err := s.getUserFromDB(respUser.ID)
	s.NoError(err, "Failed to get user from database")

	s.Require().Equal(username, dbUser.Username)
	s.Require().Equal(email, dbUser.Email)
	s.Require().Equal(birthDate, dbUser.BirthDate)
	s.NoError(
		bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)),
		"Sent password and stored hash aren't equal",
	)
}

func (s *AppTestSuite) TestGetUser() {
	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	req, err := http.NewRequest(http.MethodGet, s.buildURL("/users/ba566522-3305-48df-936a-73f47611934b"), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.NoError(err, "Failed to send request")

	defer func() { _ = resp.Body.Close() }()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var respUser domain.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	s.NoError(err, "Failed to decode response body")

	dbUser, err := s.getUserFromDB(respUser.ID)
	s.NoError(err, "Failed to get user by id")

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

	req, err := http.NewRequest(http.MethodPatch, s.buildURL("/users/"+expectedUser.ID), bytes.NewReader(payload))
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.NoError(err, "Failed to send request")

	defer func() { _ = resp.Body.Close() }()
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
	oldPassword, newPassword := "helloworld12345", "qwerty55555"
	strBody := fmt.Sprintf(`{"password":"%s"}`, newPassword)
	tokenPair := s.authenticate("mick47", oldPassword, uuid.New().String())

	req, err := http.NewRequest(
		http.MethodPatch,
		s.buildURL("/users/7e7b1825-ef9a-42ec-b4db-6f09dffe3850"),
		strings.NewReader(strBody),
	)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	s.authenticate("mick47", newPassword, uuid.New().String())
}

func (s *AppTestSuite) TestDeleteUser() {
	id := "ba566522-3305-48df-936a-73f47611934b"

	// Check if the user exists before delete
	_, err := s.getUserFromDB(id)
	s.NoError(err, "Failed to get user from database")

	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	req, err := http.NewRequest(http.MethodDelete, s.buildURL("/users/"+id), nil)
	s.NoError(err, "Failed to create request")

	req.Header.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	resp, err := s.httpClient.Do(req)
	s.NoError(err, "Failed to send request")
	s.Require().Equal(http.StatusNoContent, resp.StatusCode)

	// Check if the user exists after delete
	_, err = s.getUserFromDB(id)
	s.Require().Equal(sql.ErrNoRows, err, "User hasn't been deleted")
}

func (s *AppTestSuite) getUserFromDB(id string) (domain.User, error) {
	var (
		user      domain.User
		birthDate pgtype.Date
	)

	query := fmt.Sprintf(
		`SELECT %s FROM users WHERE id = $1 AND is_deleted IS FALSE`,
		strings.Join(userTableColumns, ", "),
	)

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

	query := fmt.Sprintf(
		`SELECT %s FROM users WHERE is_deleted IS FALSE`,
		strings.Join(userTableColumns, ", "),
	)

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
