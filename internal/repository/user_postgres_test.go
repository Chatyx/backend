package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

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
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, id string)

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
		mockBehavior mockBehavior
		expectedUser domain.User
		expectedErr  error
	}{
		{
			name: "Success get short user",
			id:   "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, id string) {
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
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, id string) {
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
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, id string) {
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
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, id string) {
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

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.id)
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

func TestUserPostgresRepository_Create(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, dto domain.CreateUserDTO, rowValues []interface{}, rowErr error)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	userTableCreateColumns := []string{
		"username", "password", "first_name",
		"last_name", "email", "birth_date", "department",
	}
	returnedColumns := []string{"id", "created_at"}

	placeholders := make([]string, 0, len(userTableCreateColumns))
	for i := range userTableCreateColumns {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}

	query := fmt.Sprintf(
		"INSERT INTO users (%s) VALUES (%s) RETURNING %s",
		strings.Join(userTableCreateColumns, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(returnedColumns, ", "),
	)

	var defaultMockBehaviour mockBehavior = func(mockPool pgxmock.PgxPoolIface, dto domain.CreateUserDTO, rowValues []interface{}, rowErr error) {
		var birthDate pgtype.Date
		if dto.BirthDate != "" {
			tm, _ := time.Parse("2006-01-02", dto.BirthDate)
			birthDate = pgtype.Date{Time: tm, Status: pgtype.Present}
		} else {
			birthDate = pgtype.Date{Status: pgtype.Null}
		}

		expected := mockPool.ExpectQuery(query).
			WithArgs(
				dto.Username, dto.Password, dto.FirstName,
				dto.LastName, dto.Email, birthDate, dto.Department,
			)

		if rowValues != nil {
			expected.WillReturnRows(pgxmock.NewRows(returnedColumns).AddRow(rowValues...))
		}

		if rowErr != nil {
			expected.WillReturnError(rowErr)
		}
	}

	testTable := []struct {
		name          string
		createUserDTO domain.CreateUserDTO
		rowValues     []interface{}
		rowErr        error
		mockBehavior  mockBehavior
		expectedUser  domain.User
		expectedErr   error
	}{
		{
			name: "Success with required fields",
			createUserDTO: domain.CreateUserDTO{
				Username: "john1967",
				Password: "8743b52063cd84097a65d1633f5c74f5",
				Email:    "john1967@gmail.com",
			},
			rowValues:    []interface{}{"6be043ca-3005-4b1c-b847-eb677897c618", &userCreatedAt},
			mockBehavior: defaultMockBehaviour,
			expectedUser: defaultShortUser,
			expectedErr:  nil,
		},
		{
			name: "Success with full fields",
			createUserDTO: domain.CreateUserDTO{
				Username:   "mick47",
				Password:   "8743b52063cd84097a65d1633f5c74f5",
				Email:      "mick47@gmail.com",
				FirstName:  "Mick",
				LastName:   "Tyson",
				BirthDate:  "1949-10-25",
				Department: "IoT",
			},
			rowValues:    []interface{}{"02185cd4-05b5-4688-836d-3154e9c8a340", &userCreatedAt},
			mockBehavior: defaultMockBehaviour,
			expectedUser: domain.User{
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
			},
		},
		{
			name: "User with such username or email already exists",
			createUserDTO: domain.CreateUserDTO{
				Username: "john1967",
				Password: "8743b52063cd84097a65d1633f5c74f5",
				Email:    "john1967@gmail.com",
			},
			rowErr:       &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			mockBehavior: defaultMockBehaviour,
			expectedErr:  domain.ErrUserUniqueViolation,
		},
		{
			name:         "Unexpected error while creating user",
			rowErr:       errUnexpected,
			mockBehavior: defaultMockBehaviour,
			expectedErr:  errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("An error occurred while opening a mock pool of database connections: %v", err)
			}
			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.createUserDTO, testCase.rowValues, testCase.rowErr)
			}

			userRepo := NewUserPostgresRepository(mockPool)

			user, err := userRepo.Create(context.Background(), testCase.createUserDTO)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedUser, user) {
				t.Errorf("Wrong created user. Expected %#v, got %#v", testCase.expectedUser, user)
			}

			if err = mockPool.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestInvalidDateBirthdayParse(t *testing.T) {
	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("An error occurred while opening a mock pool of database connections: %v", err)
	}
	defer mockPool.Close()

	userRepo := NewUserPostgresRepository(mockPool)

	t.Run("Create user", func(t *testing.T) {
		_, err = userRepo.Create(context.Background(), domain.CreateUserDTO{
			Username:  "john1967",
			Password:  "8743b52063cd84097a65d1633f5c74f5",
			Email:     "john1967@gmail.com",
			BirthDate: "invalid-date",
		})

		if err == nil {
			t.Errorf("Expected parse error, got nil")
		}

		if err != nil {
			if _, ok := err.(*time.ParseError); !ok {
				t.Errorf("Expected parse error, got: %v", err)
			}
		}
	})

	t.Run("Update user", func(t *testing.T) {
		_, err = userRepo.Update(context.Background(), domain.UpdateUserDTO{
			ID:        "02185cd4-05b5-4688-836d-3154e9c8a340",
			BirthDate: "invalid-date",
		})

		if err == nil {
			t.Errorf("Expected parse error, got nil")
		}

		if err != nil {
			if _, ok := err.(*time.ParseError); !ok {
				t.Errorf("Expected parse error, got: %v", err)
			}
		}
	})
}
