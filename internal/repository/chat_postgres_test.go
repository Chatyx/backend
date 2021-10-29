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
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	chatCreatedAt = time.Date(2021, time.October, 25, 18, 05, 00, 0, time.Local)
	chatUpdatedAt = time.Date(2021, time.November, 17, 23, 0, 42, 142, time.Local)
)

var chatTableColumns = []string{
	"chats.id", "chats.name", "chats.description",
	"chats.creator_id", "chats.created_at", "chats.updated_at",
}

var (
	chatWithRequiredFields = domain.Chat{
		ID:        "d1596312-4943-434a-86aa-edadc7e9aaf2",
		Name:      "Test chat_1 name",
		CreatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
		CreatedAt: &chatCreatedAt,
	}
	chatWithFullFields = domain.Chat{
		ID:          "afccfc65-b8c3-4e37-8717-3136a246bf09",
		Name:        "Test chat_2 name",
		Description: "Test chat_2 description",
		CreatorID:   "6be043ca-3005-4b1c-b847-eb677897c618",
		CreatedAt:   &chatCreatedAt,
		UpdatedAt:   &chatUpdatedAt,
	}
)

var (
	chatWithRequiredFieldsRowValues = Row{
		"d1596312-4943-434a-86aa-edadc7e9aaf2", // ID
		"Test chat_1 name",                     // Name
		"",                                     // Description
		"6be043ca-3005-4b1c-b847-eb677897c618", // CreatorID
		&chatCreatedAt,                         // CreatedAt
		nil,                                    // UpdatedAt
	}
	chatWithFullFieldsRowValues = Row{
		"afccfc65-b8c3-4e37-8717-3136a246bf09", // ID
		"Test chat_2 name",                     // Name
		"Test chat_2 description",              // Description
		"6be043ca-3005-4b1c-b847-eb677897c618", // CreatorID
		&chatCreatedAt,                         // CreatedAt
		&chatUpdatedAt,                         // UpdatedAt
	}
)

func TestChatPostgresRepository_List(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, memberID string, rowsRes []RowResult, rowsErr error)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := fmt.Sprintf(`SELECT %s FROM chats 
	INNER JOIN chat_members 
		ON chats.id = chat_members.chat_id
	INNER JOIN users
		ON chat_members.user_id = users.id
	WHERE chat_members.user_id = $1 AND users.is_deleted IS FALSE`,
		strings.Join(chatTableColumns, ", "))

	var defaultMockBehaviour mockBehavior = func(mockPool pgxmock.PgxPoolIface, memberID string, rowsRes []RowResult, rowsErr error) {
		expected := mockPool.ExpectQuery(query).WithArgs(memberID)

		if len(rowsRes) != 0 {
			rows := pgxmock.NewRows(chatTableColumns)

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
		memberID      string
		rowsResult    []RowResult
		rowsErr       error
		mockBehavior  mockBehavior
		expectedChats []domain.Chat
		expectedErr   error
	}{
		{
			name:     "Success",
			memberID: "bf7f421b-b60e-494b-8afe-2bb4edd8b5fc",
			rowsResult: []RowResult{
				{Row: chatWithFullFieldsRowValues},
				{Row: chatWithRequiredFieldsRowValues},
			},
			mockBehavior: defaultMockBehaviour,
			expectedChats: []domain.Chat{
				chatWithFullFields,
				chatWithRequiredFields,
			},
			expectedErr: nil,
		},
		{
			name:         "Unexpected error while query list of chats",
			memberID:     "bf7f421b-b60e-494b-8afe-2bb4edd8b5fc",
			rowsErr:      errUnexpected,
			mockBehavior: defaultMockBehaviour,
			expectedErr:  errUnexpected,
		},
		{
			name:     "Unexpected error while scanning chat",
			memberID: "bf7f421b-b60e-494b-8afe-2bb4edd8b5fc",
			rowsResult: []RowResult{
				{Row: chatWithRequiredFieldsRowValues, Err: errUnexpected},
			},
			mockBehavior: defaultMockBehaviour,
			expectedErr:  errUnexpected,
		},
		{
			name:     "Unexpected row error",
			memberID: "bf7f421b-b60e-494b-8afe-2bb4edd8b5fc",
			rowsResult: []RowResult{
				{Err: errUnexpected},
			},
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
				testCase.mockBehavior(mockPool, testCase.memberID, testCase.rowsResult, testCase.rowsErr)
			}

			chatRepo := NewChatPostgresRepository(mockPool)

			chats, err := chatRepo.List(context.Background(), testCase.memberID)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedChats, chats)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestChatPostgresRepository_Create(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, dto domain.CreateChatDTO)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	chatInsertQuery := `INSERT INTO chats (name, description, creator_id) VALUES ($1, $2, $3) RETURNING id, created_at`
	usersChatsInsertQuery := `INSERT INTO chat_members (user_id, chat_id) VALUES ($1, $2)`

	testTable := []struct {
		name          string
		createChatDTO domain.CreateChatDTO
		mockBehavior  mockBehavior
		expectedChat  domain.Chat
		expectedErr   error
	}{
		{
			name: "Success",
			createChatDTO: domain.CreateChatDTO{
				Name:        "Test chat_2 name",
				Description: "Test chat_2 description",
				CreatorID:   "6be043ca-3005-4b1c-b847-eb677897c618",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.CreateChatDTO) {
				createdChatID := "afccfc65-b8c3-4e37-8717-3136a246bf09"

				mockPool.ExpectBeginTx(pgx.TxOptions{})

				mockPool.ExpectQuery(chatInsertQuery).
					WithArgs(dto.Name, dto.Description, dto.CreatorID).
					WillReturnRows(
						pgxmock.NewRows([]string{"id", "created_at"}).
							AddRow(createdChatID, &chatCreatedAt),
					)
				mockPool.ExpectExec(usersChatsInsertQuery).
					WithArgs(dto.CreatorID, createdChatID).
					WillReturnResult(pgxmock.NewResult("created", 1))

				mockPool.ExpectCommit()
			},
			expectedChat: domain.Chat{
				ID:          "afccfc65-b8c3-4e37-8717-3136a246bf09",
				Name:        "Test chat_2 name",
				Description: "Test chat_2 description",
				CreatorID:   "6be043ca-3005-4b1c-b847-eb677897c618",
				CreatedAt:   &chatCreatedAt,
			},
			expectedErr: nil,
		},
		{
			name: "Unexpected error after inserting into chats table",
			createChatDTO: domain.CreateChatDTO{
				Name:      "Test chat_1 name",
				CreatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.CreateChatDTO) {
				mockPool.ExpectBeginTx(pgx.TxOptions{})
				mockPool.ExpectQuery(chatInsertQuery).
					WithArgs(dto.Name, dto.Description, dto.CreatorID).
					WillReturnError(errUnexpected)

				mockPool.ExpectRollback()
				mockPool.ExpectRollback()
			},
			expectedErr: errUnexpected,
		},
		{
			name: "Unexpected error after inserting chat_members table",
			createChatDTO: domain.CreateChatDTO{
				Name:      "Test chat_1 name",
				CreatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.CreateChatDTO) {
				createdChatID := "d1596312-4943-434a-86aa-edadc7e9aaf2"

				mockPool.ExpectBeginTx(pgx.TxOptions{})
				mockPool.ExpectQuery(chatInsertQuery).
					WithArgs(dto.Name, dto.Description, dto.CreatorID).
					WillReturnRows(
						pgxmock.NewRows([]string{"id", "created_at"}).
							AddRow(createdChatID, &chatCreatedAt),
					)
				mockPool.ExpectExec(usersChatsInsertQuery).
					WithArgs(dto.CreatorID, createdChatID).
					WillReturnError(errUnexpected)

				mockPool.ExpectRollback()
				mockPool.ExpectRollback()
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
				testCase.mockBehavior(mockPool, testCase.createChatDTO)
			}

			chatRepo := NewChatPostgresRepository(mockPool)

			chat, err := chatRepo.Create(context.Background(), testCase.createChatDTO)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedChat, chat)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestChatPostgresRepository_GetByID(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, chatID, memberID string, returnRow Row, rowErr error)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := fmt.Sprintf(`SELECT %s FROM chats 
	INNER JOIN chat_members 
		ON chats.id = chat_members.chat_id
	WHERE chats.id = $1 AND chat_members.user_id = $2`,
		strings.Join(chatTableColumns, ", "))

	var defaultMockBehavior mockBehavior = func(mockPool pgxmock.PgxPoolIface, chatID, memberID string, returnRow Row, rowErr error) {
		expected := mockPool.ExpectQuery(query).WithArgs(chatID, memberID)

		if len(returnRow) != 0 {
			expected.WillReturnRows(
				pgxmock.NewRows(chatTableColumns).AddRow(returnRow...),
			)
		}

		if rowErr != nil {
			expected.WillReturnError(rowErr)
		}
	}

	testTable := []struct {
		name         string
		chatID       string
		memberID     string
		returnRow    Row
		rowErr       error
		mockBehavior mockBehavior
		expectedChat domain.Chat
		expectedErr  error
	}{
		{
			name:         "Success",
			chatID:       "afccfc65-b8c3-4e37-8717-3136a246bf09",
			memberID:     "6be043ca-3005-4b1c-b847-eb677897c618",
			returnRow:    chatWithFullFieldsRowValues,
			mockBehavior: defaultMockBehavior,
			expectedChat: chatWithFullFields,
			expectedErr:  nil,
		},
		{
			name:         "Chat is not found",
			chatID:       "d1596312-4943-434a-86aa-edadc7e9aaf2",
			memberID:     "6be043ca-3005-4b1c-b847-eb677897c618",
			rowErr:       pgx.ErrNoRows,
			mockBehavior: defaultMockBehavior,
			expectedErr:  domain.ErrChatNotFound,
		},
		{
			name:        "With invalid chat_id (not uuid4)",
			chatID:      "1",
			memberID:    "6be043ca-3005-4b1c-b847-eb677897c618",
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name:        "With invalid member_id (not uuid4)",
			chatID:      "d1596312-4943-434a-86aa-edadc7e9aaf2",
			memberID:    "1",
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name:         "Unexpected error while getting chat",
			chatID:       "d1596312-4943-434a-86aa-edadc7e9aaf2",
			memberID:     "6be043ca-3005-4b1c-b847-eb677897c618",
			rowErr:       errUnexpected,
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
				testCase.mockBehavior(mockPool, testCase.chatID, testCase.memberID, testCase.returnRow, testCase.rowErr)
			}

			chatRepo := NewChatPostgresRepository(mockPool)

			chat, err := chatRepo.GetByID(context.Background(), testCase.chatID, testCase.memberID)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedChat, chat)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestChatPostgresRepository_Update(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateChatDTO)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := `UPDATE chats SET name = $1, description = $2 WHERE id = $3 AND creator_id = $4 RETURNING created_at, updated_at`

	testTable := []struct {
		name          string
		updateChatDTO domain.UpdateChatDTO
		mockBehavior  mockBehavior
		expectedChat  domain.Chat
		expectedErr   error
	}{
		{
			name: "Success",
			updateChatDTO: domain.UpdateChatDTO{
				ID:          "afccfc65-b8c3-4e37-8717-3136a246bf09",
				Name:        "Test chat_2 name",
				Description: "Test chat_2 description",
				CreatorID:   "6be043ca-3005-4b1c-b847-eb677897c618",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateChatDTO) {
				mockPool.ExpectQuery(query).
					WithArgs(dto.Name, dto.Description, dto.ID, dto.CreatorID).
					WillReturnRows(
						pgxmock.NewRows([]string{"created_at", "updated_at"}).
							AddRow(&chatCreatedAt, &chatUpdatedAt),
					)
			},
			expectedChat: chatWithFullFields,
			expectedErr:  nil,
		},
		{
			name: "With invalid chat_id (not uuid4)",
			updateChatDTO: domain.UpdateChatDTO{
				ID:        "1",
				Name:      "Test chat_1 name",
				CreatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			},
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name: "With invalid creator_id (not uuid4)",
			updateChatDTO: domain.UpdateChatDTO{
				ID:        "d1596312-4943-434a-86aa-edadc7e9aaf2",
				Name:      "Test chat_1 name",
				CreatorID: "1",
			},
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name: "Chat is not found",
			updateChatDTO: domain.UpdateChatDTO{
				ID:        "d1596312-4943-434a-86aa-edadc7e9aaf2",
				Name:      "Test chat_1 name",
				CreatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateChatDTO) {
				mockPool.ExpectQuery(query).
					WithArgs(dto.Name, dto.Description, dto.ID, dto.CreatorID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name: "Unexpected error",
			updateChatDTO: domain.UpdateChatDTO{
				ID:        "d1596312-4943-434a-86aa-edadc7e9aaf2",
				Name:      "Test chat_1 name",
				CreatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			},
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, dto domain.UpdateChatDTO) {
				mockPool.ExpectQuery(query).
					WithArgs(dto.Name, dto.Description, dto.ID, dto.CreatorID).
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
				testCase.mockBehavior(mockPool, testCase.updateChatDTO)
			}

			chatRepo := NewChatPostgresRepository(mockPool)

			chat, err := chatRepo.Update(context.Background(), testCase.updateChatDTO)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			assert.Equal(t, testCase.expectedChat, chat)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestChatPostgresRepository_Delete(t *testing.T) {
	type mockBehavior func(mockPool pgxmock.PgxPoolIface, chatID, creatorID string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	query := `DELETE FROM chats WHERE id = $1 AND creator_id = $2`

	testTable := []struct {
		name         string
		chatID       string
		creatorID    string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:      "Success",
			chatID:    "afccfc65-b8c3-4e37-8717-3136a246bf09",
			creatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, chatID, creatorID string) {
				mockPool.ExpectExec(query).
					WithArgs(chatID, creatorID).
					WillReturnResult(pgxmock.NewResult("deleted", 1))
			},
			expectedErr: nil,
		},
		{
			name:      "Chat is not found",
			chatID:    "afccfc65-b8c3-4e37-8717-3136a246bf09",
			creatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, chatID, creatorID string) {
				mockPool.ExpectExec(query).
					WithArgs(chatID, creatorID).
					WillReturnResult(pgxmock.NewResult("deleted", 0))
			},
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name:        "With invalid chat_id (not uuid4)",
			chatID:      "1",
			creatorID:   "6be043ca-3005-4b1c-b847-eb677897c618",
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name:        "With invalid creator_id (not uuid4)",
			chatID:      "afccfc65-b8c3-4e37-8717-3136a246bf09",
			creatorID:   "1",
			expectedErr: domain.ErrChatNotFound,
		},
		{
			name:      "Unexpected error",
			chatID:    "afccfc65-b8c3-4e37-8717-3136a246bf09",
			creatorID: "6be043ca-3005-4b1c-b847-eb677897c618",
			mockBehavior: func(mockPool pgxmock.PgxPoolIface, chatID, creatorID string) {
				mockPool.ExpectExec(query).
					WithArgs(chatID, creatorID).
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
				testCase.mockBehavior(mockPool, testCase.chatID, testCase.creatorID)
			}

			chatRepo := NewChatPostgresRepository(mockPool)

			err = chatRepo.Delete(context.Background(), testCase.chatID, testCase.creatorID)

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
