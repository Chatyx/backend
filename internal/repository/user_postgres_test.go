// +build unit

package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	defaultShortUserRowValues = Row{
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
	defaultFullUserRowValues = Row{
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

func TestUserPostgresRepository_List(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, resRows []RowResult, rowsErr error)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := fmt.Sprintf(
		"SELECT %s FROM users WHERE is_deleted IS FALSE",
		strings.Join(userTableColumns, ", "),
	)

	var defaultMockBehavior mockBehavior = func(mockPool pgxmock.PgxPoolIface, rowsRes []RowResult, rowsErr error) {
		expected := mockPool.ExpectQuery(query)

		if len(rowsRes) != 0 {
			rows := pgxmock.NewRows(userTableColumns)
			for i, rowRes := range rowsRes {
				if len(rowRes.Row) != 0 {
					rows.AddRow(rowRes.Row...)
				}

				if rowRes.Err != nil {
					rows.RowError(i, rowRes.Err)
				}
			}

			expected.WillReturnRows(rows)
		}

		if rowsErr != nil {
			expected.WillReturnError(rowsErr)
		} else {
			expected.RowsWillBeClosed()
		}
	}

	testTable := []struct {
		name          string
		rows          []RowResult
		rowsErr       error
		mockBehavior  mockBehavior
		expectedUsers []domain.User
		expectedErr   error
	}{
		{
			name: "Success",
			rows: []RowResult{
				{Row: defaultShortUserRowValues},
				{Row: defaultFullUserRowValues},
			},
			mockBehavior:  defaultMockBehavior,
			expectedUsers: []domain.User{defaultShortUser, defaultFullUser},
		},
		{
			name:         "Unexpected error while query list of users",
			rowsErr:      errUnexpected,
			mockBehavior: defaultMockBehavior,
			expectedErr:  errUnexpected,
		},
		{
			name: "Unexpected error while scanning the user",
			rows: []RowResult{
				{Row: defaultFullUserRowValues, Err: errUnexpected},
			},
			mockBehavior: defaultMockBehavior,
			expectedErr:  errUnexpected,
		},
		{
			name:         "Unexpected row error",
			rows:         []RowResult{{Err: errUnexpected}},
			mockBehavior: defaultMockBehavior,
			expectedErr:  errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.rows, testCase.rowsErr)
			}

			userRepo := NewUserPostgresRepository(mockPool)

			users, err := userRepo.List(context.Background())

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedUsers, users)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestUserPostgresRepository_Create(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, dto domain.CreateUserDTO, row Row, rowErr error)

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

	var defaultMockBehaviour mockBehavior = func(mockPool pgxmock.PgxPoolIface, dto domain.CreateUserDTO, row Row, rowErr error) {
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

		if len(row) != 0 {
			expected.WillReturnRows(pgxmock.NewRows(returnedColumns).AddRow(row...))
		}

		if rowErr != nil {
			expected.WillReturnError(rowErr)
		}
	}

	testTable := []struct {
		name          string
		createUserDTO domain.CreateUserDTO
		row           Row
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
			row:          Row{"6be043ca-3005-4b1c-b847-eb677897c618", &userCreatedAt},
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
			row:          Row{"02185cd4-05b5-4688-836d-3154e9c8a340", &userCreatedAt},
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
			require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.createUserDTO, testCase.row, testCase.rowErr)
			}

			userRepo := NewUserPostgresRepository(mockPool)

			user, err := userRepo.Create(context.Background(), testCase.createUserDTO)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedUser, user)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestUserPostgresRepository_GetByID(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, id string, row Row, rowErr error)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := fmt.Sprintf(
		"SELECT %s FROM users WHERE id = $1 AND is_deleted IS FALSE",
		strings.Join(userTableColumns, ", "),
	)

	var defaultMockBehaviour mockBehavior = func(mockPool pgxmock.PgxPoolIface, id string, row Row, rowErr error) {
		expected := mockPool.ExpectQuery(query).WithArgs(id)

		if len(row) != 0 {
			expected.WillReturnRows(
				pgxmock.NewRows(userTableColumns).AddRow(row...),
			)
		}

		if rowErr != nil {
			expected.WillReturnError(rowErr)
		}
	}

	testTable := []struct {
		name         string
		id           string
		row          Row
		rowErr       error
		mockBehavior mockBehavior
		expectedUser domain.User
		expectedErr  error
	}{
		{
			name:         "Success get short user",
			id:           "6be043ca-3005-4b1c-b847-eb677897c618",
			row:          defaultShortUserRowValues,
			mockBehavior: defaultMockBehaviour,
			expectedUser: defaultShortUser,
			expectedErr:  nil,
		},
		{
			name:         "Success get full user",
			id:           "02185cd4-05b5-4688-836d-3154e9c8a340",
			row:          defaultFullUserRowValues,
			mockBehavior: defaultMockBehaviour,
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name:         "User is not found",
			id:           "6be043ca-3005-4b1c-b847-eb677897c618",
			rowErr:       pgx.ErrNoRows,
			mockBehavior: defaultMockBehaviour,
			expectedErr:  domain.ErrUserNotFound,
		},
		{
			name:        "Get user with invalid id (not uuid4)",
			id:          "1",
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:         "Unexpected error while getting user",
			id:           "6be043ca-3005-4b1c-b847-eb677897c618",
			rowErr:       errUnexpected,
			mockBehavior: defaultMockBehaviour,
			expectedErr:  errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.id, testCase.row, testCase.rowErr)
			}

			userRepo := NewUserPostgresRepository(mockPool)
			user, err := userRepo.GetByID(context.Background(), testCase.id)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedUser, user)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestUserPostgresRepository_GetByUsername(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, username string, row Row, rowErr error)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := fmt.Sprintf(
		"SELECT %s FROM users WHERE username = $1 AND is_deleted IS FALSE",
		strings.Join(userTableColumns, ", "),
	)

	var defaultMockBehaviour mockBehavior = func(mockPool pgxmock.PgxPoolIface, username string, row Row, rowErr error) {
		expected := mockPool.ExpectQuery(query).WithArgs(username)

		if len(row) != 0 {
			expected.WillReturnRows(
				pgxmock.NewRows(userTableColumns).AddRow(row...),
			)
		}

		if rowErr != nil {
			expected.WillReturnError(rowErr)
		}
	}

	testTable := []struct {
		name         string
		username     string
		row          Row ``
		rowErr       error
		mockBehavior mockBehavior
		expectedUser domain.User
		expectedErr  error
	}{
		{
			name:         "Success get short user",
			username:     "john1967",
			row:          defaultShortUserRowValues,
			mockBehavior: defaultMockBehaviour,
			expectedUser: defaultShortUser,
			expectedErr:  nil,
		},
		{
			name:         "Success get full user",
			username:     "mick47",
			row:          defaultFullUserRowValues,
			mockBehavior: defaultMockBehaviour,
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name:         "User is not found",
			username:     "john1967",
			rowErr:       pgx.ErrNoRows,
			mockBehavior: defaultMockBehaviour,
			expectedErr:  domain.ErrUserNotFound,
		},
		{
			name:         "Unexpected error while getting user",
			username:     "john1967",
			rowErr:       errUnexpected,
			mockBehavior: defaultMockBehaviour,
			expectedErr:  errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.username, testCase.row, testCase.rowErr)
			}

			userRepo := NewUserPostgresRepository(mockPool)
			user, err := userRepo.GetByUsername(context.Background(), testCase.username)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedUser, user)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestUserPostgresRepository_Update(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateUserDTO)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := `UPDATE users SET
		username = $2, email = $3, first_name = $4, 
		last_name = $5, birth_date = $6, department = $7
	WHERE id = $1 AND is_deleted IS FALSE
	RETURNING password, is_deleted, created_at, updated_at`

	testTable := []struct {
		name          string
		updateUserDTO domain.UpdateUserDTO
		mockBehavior  mockBehavior
		expectedUser  domain.User
		expectedErr   error
	}{
		{
			name: "Success",
			updateUserDTO: domain.UpdateUserDTO{
				ID:         "02185cd4-05b5-4688-836d-3154e9c8a340",
				Username:   "mick47",
				Email:      "mick47@gmail.com",
				FirstName:  "Mick",
				LastName:   "Tyson",
				BirthDate:  "1949-10-25",
				Department: "IoT",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateUserDTO) {
				birthDate := pgtype.Date{
					Status: pgtype.Present,
					Time:   time.Date(1949, time.October, 25, 0, 0, 0, 0, time.UTC),
				}
				returnedColumns := []string{
					"password", "is_deleted",
					"created_at", "updated_at",
				}
				returnedRow := Row{
					"8743b52063cd84097a65d1633f5c74f5", false,
					&userCreatedAt, &userUpdatedAt,
				}

				mockPool.ExpectQuery(query).
					WithArgs(
						dto.ID, dto.Username,
						dto.Email, dto.FirstName, dto.LastName,
						birthDate, dto.Department,
					).WillReturnRows(pgxmock.NewRows(returnedColumns).AddRow(returnedRow...))
			},
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name: "User is not found",
			updateUserDTO: domain.UpdateUserDTO{
				ID:       "02185cd4-05b5-4688-836d-3154e9c8a340",
				Username: "mick47",
				Email:    "mick47@gmail.com",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateUserDTO) {
				mockPool.ExpectQuery(query).
					WithArgs(
						dto.ID, dto.Username, dto.Email,
						dto.FirstName, dto.LastName,
						pgtype.Date{Status: pgtype.Null}, dto.Department,
					).WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "Update user with invalid id",
			updateUserDTO: domain.UpdateUserDTO{
				ID:       "1",
				Username: "mick47",
				Email:    "mick47@gmail.com",
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "Update user with such username or email",
			updateUserDTO: domain.UpdateUserDTO{
				ID:       "02185cd4-05b5-4688-836d-3154e9c8a340",
				Username: "mick47",
				Email:    "mick47@gmail.com",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateUserDTO) {
				mockPool.ExpectQuery(query).
					WithArgs(
						dto.ID, dto.Username, dto.Email,
						dto.FirstName, dto.LastName,
						pgtype.Date{Status: pgtype.Null}, dto.Department,
					).WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			expectedErr: domain.ErrUserUniqueViolation,
		},
		{
			name: "Unexpected error while updating user",
			updateUserDTO: domain.UpdateUserDTO{
				ID:       "02185cd4-05b5-4688-836d-3154e9c8a340",
				Username: "mick47",
				Email:    "mick47@gmail.com",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateUserDTO) {
				mockPool.ExpectQuery(query).
					WithArgs(
						dto.ID, dto.Username, dto.Email,
						dto.FirstName, dto.LastName,
						pgtype.Date{Status: pgtype.Null}, dto.Department,
					).WillReturnError(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.updateUserDTO)
			}

			userRepo := NewUserPostgresRepository(mockPool)
			user, err := userRepo.Update(context.Background(), testCase.updateUserDTO)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedUser, user)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestUserPostgresRepository_UpdatePassword(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, userID, password string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := `UPDATE users SET password = $2 WHERE id = $1 AND is_deleted IS FALSE`

	testTable := []struct {
		name         string
		userID       string
		password     string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:     "Success",
			userID:   "02185cd4-05b5-4688-836d-3154e9c8a340",
			password: "qwerty12345",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, userID, password string) {
				mockPool.ExpectExec(query).
					WithArgs(userID, password).
					WillReturnResult(pgxmock.NewResult("updated", 1))
			},
			expectedErr: nil,
		},
		{
			name:        "With invalid user id (not uuid4)",
			userID:      "1",
			password:    "qwerty12345",
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:     "User is not found",
			userID:   "02185cd4-05b5-4688-836d-3154e9c8a340",
			password: "qwerty12345",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, userID, password string) {
				mockPool.ExpectExec(query).
					WithArgs(userID, password).
					WillReturnResult(pgxmock.NewResult("updated", 0))
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:     "Unexpected error",
			userID:   "02185cd4-05b5-4688-836d-3154e9c8a340",
			password: "qwerty12345",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, userID, password string) {
				mockPool.ExpectExec(query).
					WithArgs(userID, password).
					WillReturnError(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.userID, testCase.password)
			}

			userRepo := NewUserPostgresRepository(mockPool)
			err = userRepo.UpdatePassword(context.Background(), testCase.userID, testCase.password)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestUserPostgresRepository_Delete(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, id string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := "UPDATE users SET is_deleted = TRUE WHERE id = $1 AND is_deleted IS FALSE"

	testTable := []struct {
		name         string
		id           string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "Success",
			id:   "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, id string) {
				mockPool.ExpectExec(query).WithArgs(id).
					WillReturnResult(
						pgxmock.NewResult("updated", 1),
					)
			},
			expectedErr: nil,
		},
		{
			name:        "Deleting user with invalid id",
			id:          "1",
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "User is not found",
			id:   "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, id string) {
				mockPool.ExpectExec(query).WithArgs(id).
					WillReturnResult(
						pgxmock.NewResult("updated", 0),
					)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "Unexpected error while deleting the user",
			id:   "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, id string) {
				mockPool.ExpectExec(query).WithArgs(id).WillReturnError(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

			defer mockPool.Close()

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockPool, testCase.id)
			}

			userRepo := NewUserPostgresRepository(mockPool)
			err = userRepo.Delete(context.Background(), testCase.id)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
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
	require.NoError(t, err, "An error occurred while opening a mock pool of database connections")

	defer mockPool.Close()

	userRepo := NewUserPostgresRepository(mockPool)

	t.Run("Create user", func(t *testing.T) {
		_, err = userRepo.Create(context.Background(), domain.CreateUserDTO{
			Username:  "john1967",
			Password:  "8743b52063cd84097a65d1633f5c74f5",
			Email:     "john1967@gmail.com",
			BirthDate: "invalid-date",
		})

		assert.Error(t, err)
		if err != nil {
			_, ok := err.(*time.ParseError)
			assert.Truef(t, ok, "Expected parse error, got: %v", err)
		}
	})

	t.Run("Update user", func(t *testing.T) {
		_, err = userRepo.Update(context.Background(), domain.UpdateUserDTO{
			ID:        "02185cd4-05b5-4688-836d-3154e9c8a340",
			BirthDate: "invalid-date",
		})

		assert.Error(t, err)
		if err != nil {
			_, ok := err.(*time.ParseError)
			assert.Truef(t, ok, "Expected parse error, got: %v", err)
		}
	})
}
