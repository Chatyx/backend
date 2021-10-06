package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgtype"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

var userTableColumns = []string{
	"id", "username", "password",
	"first_name", "last_name", "email",
	"birth_date", "department", "is_deleted",
	"created_at", "updated_at",
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

func (s *AppTestSuite) getUserFromDB(id string) (domain.User, error) {
	var (
		user      domain.User
		birthDate pgtype.Date
	)

	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, strings.Join(userTableColumns, ", "))
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
