package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/jackc/pgtype"

	"github.com/Mort4lis/scht-backend/pkg/logging"

	"github.com/Mort4lis/scht-backend/internal/domain"

	"github.com/pashagolub/pgxmock"
)

var userTableColumns = []string{
	"id", "username", "password",
	"first_name", "last_name", "email",
	"birth_date", "department", "is_deleted",
	"created_at", "updated_at",
}

var (
	userCreatedAt = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)
	userUpdatedAt = time.Date(2021, time.November, 14, 22, 0, 53, 512, time.Local)
)

var (
	defaultShortUser = domain.User{
		ID:        "6be043ca-3005-4b1c-b847-eb677897c618",
		Username:  "john1967",
		Password:  "8743b52063cd84097a65d1633f5c74f5",
		Email:     "john1967@gmail.com",
		CreatedAt: &userCreatedAt,
	}
	defaultFullUser = domain.User{
		ID:         "02185cd4-05b5-4688-836d-3154e9c8a340",
		Username:   "mick47",
		Password:   "8743b52063cd84097a65d1633f5c74f5",
		Email:      "mick47@gmail.com",
		FirstName:  "Mick",
		LastName:   "Tyson",
		BirthDate:  "1949-10-25",
		Department: "IoT",
		IsDeleted:  false,
		CreatedAt:  &userCreatedAt,
		UpdatedAt:  &userUpdatedAt,
	}
	defaultShortUserRowValues = []interface{}{
		"6be043ca-3005-4b1c-b847-eb677897c618", // ID
		"john1967",                             // Username
		"8743b52063cd84097a65d1633f5c74f5",     // Password
		"",                                     // FirstName
		"",                                     // LastName
		"john1967@gmail.com",                   // Email
		nil,                                    // BirthDate
		"",                                     // Department
		false,                                  // IsDeleted
		&userCreatedAt,                         // CreatedAt
		nil,                                    // UpdatedAt
	}
	defaultFullUserRowValues = []interface{}{
		"02185cd4-05b5-4688-836d-3154e9c8a340", // ID
		"mick47",                               // Username
		"8743b52063cd84097a65d1633f5c74f5",     // Password
		"Mick",                                 // FirstName
		"Tyson",                                // LastName
		"mick47@gmail.com",                     // Email
		pgtype.Date{
			Status: pgtype.Present,
			Time:   time.Date(1949, time.October, 25, 0, 0, 0, 0, time.Local),
		}, // BirthDate
		"IoT",          // Department
		false,          // IsDeleted
		&userCreatedAt, // CreatedAt
		&userUpdatedAt, // UpdatedAt
	}
)

var errUnexpected = errors.New("unexpected error")

func TestUserPostgresRepository_GetByID(t *testing.T) {
	type mockBehavour func(mockPool pgxmock.PgxPoolIface, id string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := fmt.Sprintf(
		"SELECT %s FROM users WHERE id = $1 AND is_deleted IS FALSE",
		strings.Join(userTableColumns, ", "),
	)
	testTable := []struct {
		name         string
		id           string
		mockBehavour mockBehavour
		expectedUser domain.User
		expectedErr  error
	}{
		{
			name: "Success get short user",
			id:   "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavour: func(mockPool pgxmock.PgxPoolIface, id string) {
				mockPool.ExpectQuery(query).
					WithArgs(id).
					WillReturnRows(pgxmock.NewRows(userTableColumns).AddRow(defaultShortUserRowValues...))
			},
			expectedUser: defaultShortUser,
			expectedErr:  nil,
		},
		{
			name: "Success get full user",
			id:   "02185cd4-05b5-4688-836d-3154e9c8a340",
			mockBehavour: func(mockPool pgxmock.PgxPoolIface, id string) {
				mockPool.ExpectQuery(query).
					WithArgs(id).
					WillReturnRows(pgxmock.NewRows(userTableColumns).AddRow(defaultFullUserRowValues...))
			},
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name: "User is not found",
			id:   "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavour: func(mockPool pgxmock.PgxPoolIface, id string) {
				mockPool.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:        "Get user with invalid id (not uuid4)",
			id:          "1",
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "Unexpected error while getting user",
			id:   "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavour: func(mockPool pgxmock.PgxPoolIface, id string) {
				mockPool.ExpectQuery(query).
					WithArgs(id).
					WillReturnError(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("An error occurred while opening a mock pool of database connections: %v", err)
			}
			defer mockPool.Close()

			if testCase.mockBehavour != nil {
				testCase.mockBehavour(mockPool, testCase.id)
			}

			userRepo := NewUserPostgresRepository(mockPool)
			user, err := userRepo.GetByID(context.Background(), testCase.id)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedUser, user) {
				t.Errorf("Wrong found user. Expected %#v, got %#v", testCase.expectedUser, user)
			}

			if err = mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %v", err)
			}
		})
	}
}
