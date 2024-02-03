package service

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	errUnexpected = errors.New("unexpected error")
)

func TestGroupParticipant_List(t *testing.T) {
	defaultParticipants := []entity.GroupParticipant{
		{
			GroupID: 1,
			UserID:  1,
			IsAdmin: true,
			Status:  entity.JoinedStatus,
		},
		{
			GroupID: 1,
			UserID:  2,
			IsAdmin: false,
			Status:  entity.JoinedStatus,
		},
		{
			GroupID: 1,
			UserID:  3,
			IsAdmin: false,
			Status:  entity.KickedStatus,
		},
	}

	testCases := []struct {
		name                 string
		currentUserID        int
		mockBehavior         func(repo *MockGroupParticipantRepository)
		expectedParticipants []entity.GroupParticipant
		expectedError        error
	}{
		{
			name:          "Successful",
			currentUserID: 1,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("List", mock.Anything, 1).Return(defaultParticipants, nil)
			},
			expectedParticipants: defaultParticipants,
			expectedError:        nil,
		},
		{
			name:          "Current user isn't in the group",
			currentUserID: 100,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("List", mock.Anything, 1).Return(defaultParticipants, nil)
			},
			expectedParticipants: nil,
			expectedError:        entity.ErrGroupNotFound,
		},
		{
			name:          "Current user is left from group",
			currentUserID: 3,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("List", mock.Anything, 1).Return(defaultParticipants, nil)
			},
			expectedParticipants: nil,
			expectedError:        entity.ErrGroupNotFound,
		},
		{
			name:          "Unexpected error",
			currentUserID: 1,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("List", mock.Anything, 1).Return(nil, errUnexpected)
			},
			expectedParticipants: nil,
			expectedError:        errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			txm := NewMockTransactionManager(t)
			repo := NewMockGroupParticipantRepository(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(repo)
			}

			service := NewGroupParticipant(GroupParticipantConfig{
				TxManager:  txm,
				Repository: repo,
			})
			ctx := ctxutil.WithUserID(context.Background(), ctxutil.UserID(strconv.Itoa(testCase.currentUserID)))

			participants, err := service.List(ctx, 1)
			if testCase.expectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedParticipants, participants)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestGroupParticipant_Get(t *testing.T) {
	defaultParticipant := entity.GroupParticipant{
		GroupID: 1,
		UserID:  2,
		IsAdmin: false,
		Status:  entity.JoinedStatus,
	}

	testCases := []struct {
		name                string
		currentUserID       int
		mockBehavior        func(repo *MockGroupParticipantRepository)
		expectedParticipant entity.GroupParticipant
		expectedError       error
	}{
		{
			name:          "Successful",
			currentUserID: 1,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				repo.On("Get", mock.Anything, 1, 2, false).Return(defaultParticipant, nil)
			},
			expectedParticipant: defaultParticipant,
			expectedError:       nil,
		},
		{
			name:          "Group participant doesn't exists",
			currentUserID: 1,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				repo.On("Get", mock.Anything, 1, 2, false).Return(entity.GroupParticipant{}, entity.ErrGroupParticipantNotFound)
			},
			expectedError: entity.ErrGroupParticipantNotFound,
		},
		{
			name:          "Current user isn't in the group",
			currentUserID: 1,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{}, entity.ErrGroupParticipantNotFound)
			},
			expectedError: entity.ErrGroupNotFound,
		},
		{
			name:          "Current user is kicked from group",
			currentUserID: 1,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: false,
					Status:  entity.KickedStatus,
				}, nil)
			},
			expectedError: entity.ErrGroupNotFound,
		},
		{
			name:          "Unexpected error",
			currentUserID: 1,
			mockBehavior: func(repo *MockGroupParticipantRepository) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{}, errUnexpected)
			},
			expectedError: errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			txm := NewMockTransactionManager(t)
			repo := NewMockGroupParticipantRepository(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(repo)
			}

			service := NewGroupParticipant(GroupParticipantConfig{
				TxManager:  txm,
				Repository: repo,
			})
			ctx := ctxutil.WithUserID(context.Background(), ctxutil.UserID(strconv.Itoa(testCase.currentUserID)))

			participants, err := service.Get(ctx, 1, 2)
			if testCase.expectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedParticipant, participants)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestGroupParticipant_Invite(t *testing.T) {
	defaultInvitedParticipant := entity.GroupParticipant{
		GroupID: 1,
		UserID:  2,
		IsAdmin: false,
		Status:  entity.JoinedStatus,
	}

	testCases := []struct {
		name                string
		currentUserID       int
		mockBehavior        func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer)
		expectedParticipant entity.GroupParticipant
		expectedError       error
	}{
		{
			name: "Successful",
			mockBehavior: func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				repo.On("Create", mock.Anything, &defaultInvitedParticipant).Return(nil)

				prod.On("Produce", mock.Anything, entity.ParticipantEvent{
					Type: entity.AddedParticipant,
					ChatID: entity.ChatID{
						ID:   1,
						Type: entity.GroupChatType,
					},
					UserID: 2,
				}).Return(nil)
			},
			expectedParticipant: defaultInvitedParticipant,
		},
		{
			name: "Current user isn't in the group",
			mockBehavior: func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{}, entity.ErrGroupParticipantNotFound)
			},
			expectedError: entity.ErrGroupNotFound,
		},
		{
			name: "Current user is kicked from group",
			mockBehavior: func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: false,
					Status:  entity.KickedStatus,
				}, nil)
			},
			expectedError: entity.ErrGroupNotFound,
		},
		{
			name: "Current user isn't admin",
			mockBehavior: func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: false,
					Status:  entity.JoinedStatus,
				}, nil)
			},
			expectedError: entity.ErrForbiddenPerformAction,
		},
		{
			name: "Unexpected error while getting current participant",
			mockBehavior: func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{}, errUnexpected)
			},
			expectedError: errUnexpected,
		},
		{
			name: "Unexpected error while creating participant",
			mockBehavior: func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				repo.On("Create", mock.Anything, &defaultInvitedParticipant).Return(errUnexpected)
			},
			expectedError: errUnexpected,
		},
		{
			name: "Unexpected error while producing participant event",
			mockBehavior: func(repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				repo.On("Create", mock.Anything, &defaultInvitedParticipant).Return(nil)

				prod.On("Produce", mock.Anything, entity.ParticipantEvent{
					Type: entity.AddedParticipant,
					ChatID: entity.ChatID{
						ID:   1,
						Type: entity.GroupChatType,
					},
					UserID: 2,
				}).Return(errUnexpected)
			},
			expectedError: errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			txm := NewMockTransactionManager(t)
			repo := NewMockGroupParticipantRepository(t)
			prod := NewMockGroupParticipantEventProducer(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(repo, prod)
			}

			service := NewGroupParticipant(GroupParticipantConfig{
				TxManager:     txm,
				Repository:    repo,
				EventProducer: prod,
			})
			ctx := ctxutil.WithUserID(context.Background(), "1")

			participants, err := service.Invite(ctx, 1, 2)
			if testCase.expectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedParticipant, participants)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestGroupParticipant_UpdateStatus(t *testing.T) {
	testCases := []struct {
		name            string
		userIDForUpdate int
		statusForUpdate entity.GroupParticipantStatus
		mockBehavior    func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer)
		expectedError   error
	}{
		{
			name:            "Successful kick another user",
			userIDForUpdate: 2,
			statusForUpdate: entity.KickedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				txm.On("Do", mock.Anything, mock.Anything).Return(nil)

				prod.On("Produce", mock.Anything, entity.ParticipantEvent{
					Type: entity.RemovedParticipant,
					ChatID: entity.ChatID{
						ID:   1,
						Type: entity.GroupChatType,
					},
					UserID: 2,
				}).Return(nil)
			},
		},
		{
			name:            "Successful return of another user",
			userIDForUpdate: 2,
			statusForUpdate: entity.JoinedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				txm.On("Do", mock.Anything, mock.Anything).Return(nil)

				prod.On("Produce", mock.Anything, entity.ParticipantEvent{
					Type: entity.AddedParticipant,
					ChatID: entity.ChatID{
						ID:   1,
						Type: entity.GroupChatType,
					},
					UserID: 2,
				}).Return(nil)
			},
		},
		{
			name:            "Successful leave from the group",
			userIDForUpdate: 1,
			statusForUpdate: entity.LeftStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				txm.On("Do", mock.Anything, mock.Anything).Return(nil)

				prod.On("Produce", mock.Anything, entity.ParticipantEvent{
					Type: entity.RemovedParticipant,
					ChatID: entity.ChatID{
						ID:   1,
						Type: entity.GroupChatType,
					},
					UserID: 1,
				}).Return(nil)
			},
		},
		{
			name:            "Current user isn't in the group",
			userIDForUpdate: 2,
			statusForUpdate: entity.KickedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{}, entity.ErrGroupParticipantNotFound)
			},
			expectedError: entity.ErrGroupNotFound,
		},
		{
			name:            "Current user is kicked from group",
			userIDForUpdate: 2,
			statusForUpdate: entity.KickedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: false,
					Status:  entity.KickedStatus,
				}, nil)
			},
			expectedError: entity.ErrGroupNotFound,
		},
		{
			name:            "Current user isn't admin",
			userIDForUpdate: 2,
			statusForUpdate: entity.KickedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: false,
					Status:  entity.JoinedStatus,
				}, nil)
			},
			expectedError: entity.ErrForbiddenPerformAction,
		},
		{
			name:            "Unexpected error while getting current participant",
			userIDForUpdate: 2,
			statusForUpdate: entity.KickedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{}, errUnexpected)
			},
			expectedError: errUnexpected,
		},
		{
			name:            "Unexpected error while updating participant status",
			userIDForUpdate: 2,
			statusForUpdate: entity.KickedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				txm.On("Do", mock.Anything, mock.Anything).Return(errUnexpected)
			},
			expectedError: errUnexpected,
		},
		{
			name:            "Unexpected error while producing participant event",
			userIDForUpdate: 2,
			statusForUpdate: entity.KickedStatus,
			mockBehavior: func(txm *MockTransactionManager, repo *MockGroupParticipantRepository, prod *MockGroupParticipantEventProducer) {
				repo.On("Get", mock.Anything, 1, 1, false).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  1,
					IsAdmin: true,
					Status:  entity.JoinedStatus,
				}, nil)

				txm.On("Do", mock.Anything, mock.Anything).Return(nil)

				prod.On("Produce", mock.Anything, entity.ParticipantEvent{
					Type: entity.RemovedParticipant,
					ChatID: entity.ChatID{
						ID:   1,
						Type: entity.GroupChatType,
					},
					UserID: 2,
				}).Return(errUnexpected)
			},
			expectedError: errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			txm := NewMockTransactionManager(t)
			repo := NewMockGroupParticipantRepository(t)
			prod := NewMockGroupParticipantEventProducer(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(txm, repo, prod)
			}

			service := NewGroupParticipant(GroupParticipantConfig{
				TxManager:     txm,
				Repository:    repo,
				EventProducer: prod,
			})
			ctx := ctxutil.WithUserID(context.Background(), "1")

			err := service.UpdateStatus(ctx, 1, testCase.userIDForUpdate, testCase.statusForUpdate)
			if testCase.expectedError == nil {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}
