// Code generated by MockGen. DO NOT EDIT.
// Source: repository.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	reflect "reflect"
	time "time"

	domain "github.com/Mort4lis/scht-backend/internal/domain"
	repository "github.com/Mort4lis/scht-backend/internal/repository"
	gomock "github.com/golang/mock/gomock"
)

// MockUserRepository is a mock of UserRepository interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockUserRepository) Create(ctx context.Context, dto domain.CreateUserDTO) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, dto)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockUserRepositoryMockRecorder) Create(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUserRepository)(nil).Create), ctx, dto)
}

// Delete mocks base method.
func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUserRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserRepository)(nil).Delete), ctx, id)
}

// GetByID mocks base method.
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockUserRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUserRepository)(nil).GetByID), ctx, id)
}

// GetByUsername mocks base method.
func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByUsername", ctx, username)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByUsername indicates an expected call of GetByUsername.
func (mr *MockUserRepositoryMockRecorder) GetByUsername(ctx, username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUsername", reflect.TypeOf((*MockUserRepository)(nil).GetByUsername), ctx, username)
}

// List mocks base method.
func (m *MockUserRepository) List(ctx context.Context) ([]domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].([]domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockUserRepositoryMockRecorder) List(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUserRepository)(nil).List), ctx)
}

// Update mocks base method.
func (m *MockUserRepository) Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, dto)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockUserRepositoryMockRecorder) Update(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUserRepository)(nil).Update), ctx, dto)
}

// UpdatePassword mocks base method.
func (m *MockUserRepository) UpdatePassword(ctx context.Context, id, password string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePassword", ctx, id, password)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePassword indicates an expected call of UpdatePassword.
func (mr *MockUserRepositoryMockRecorder) UpdatePassword(ctx, id, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockUserRepository)(nil).UpdatePassword), ctx, id, password)
}

// MockSessionRepository is a mock of SessionRepository interface.
type MockSessionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSessionRepositoryMockRecorder
}

// MockSessionRepositoryMockRecorder is the mock recorder for MockSessionRepository.
type MockSessionRepositoryMockRecorder struct {
	mock *MockSessionRepository
}

// NewMockSessionRepository creates a new mock instance.
func NewMockSessionRepository(ctrl *gomock.Controller) *MockSessionRepository {
	mock := &MockSessionRepository{ctrl: ctrl}
	mock.recorder = &MockSessionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionRepository) EXPECT() *MockSessionRepositoryMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockSessionRepository) Delete(ctx context.Context, refreshToken, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, refreshToken, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockSessionRepositoryMockRecorder) Delete(ctx, refreshToken, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSessionRepository)(nil).Delete), ctx, refreshToken, userID)
}

// DeleteAllByUserID mocks base method.
func (m *MockSessionRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAllByUserID", ctx, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAllByUserID indicates an expected call of DeleteAllByUserID.
func (mr *MockSessionRepositoryMockRecorder) DeleteAllByUserID(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllByUserID", reflect.TypeOf((*MockSessionRepository)(nil).DeleteAllByUserID), ctx, userID)
}

// Get mocks base method.
func (m *MockSessionRepository) Get(ctx context.Context, refreshToken string) (domain.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, refreshToken)
	ret0, _ := ret[0].(domain.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSessionRepositoryMockRecorder) Get(ctx, refreshToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSessionRepository)(nil).Get), ctx, refreshToken)
}

// Set mocks base method.
func (m *MockSessionRepository) Set(ctx context.Context, session domain.Session, ttl time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, session, ttl)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockSessionRepositoryMockRecorder) Set(ctx, session, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockSessionRepository)(nil).Set), ctx, session, ttl)
}

// MockChatRepository is a mock of ChatRepository interface.
type MockChatRepository struct {
	ctrl     *gomock.Controller
	recorder *MockChatRepositoryMockRecorder
}

// MockChatRepositoryMockRecorder is the mock recorder for MockChatRepository.
type MockChatRepositoryMockRecorder struct {
	mock *MockChatRepository
}

// NewMockChatRepository creates a new mock instance.
func NewMockChatRepository(ctrl *gomock.Controller) *MockChatRepository {
	mock := &MockChatRepository{ctrl: ctrl}
	mock.recorder = &MockChatRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChatRepository) EXPECT() *MockChatRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockChatRepository) Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, dto)
	ret0, _ := ret[0].(domain.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockChatRepositoryMockRecorder) Create(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockChatRepository)(nil).Create), ctx, dto)
}

// Delete mocks base method.
func (m *MockChatRepository) Delete(ctx context.Context, chatID, creatorID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, chatID, creatorID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockChatRepositoryMockRecorder) Delete(ctx, chatID, creatorID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockChatRepository)(nil).Delete), ctx, chatID, creatorID)
}

// GetByID mocks base method.
func (m *MockChatRepository) GetByID(ctx context.Context, chatID, memberID string) (domain.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, chatID, memberID)
	ret0, _ := ret[0].(domain.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockChatRepositoryMockRecorder) GetByID(ctx, chatID, memberID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockChatRepository)(nil).GetByID), ctx, chatID, memberID)
}

// GetOwnByID mocks base method.
func (m *MockChatRepository) GetOwnByID(ctx context.Context, chatID, creatorID string) (domain.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOwnByID", ctx, chatID, creatorID)
	ret0, _ := ret[0].(domain.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOwnByID indicates an expected call of GetOwnByID.
func (mr *MockChatRepositoryMockRecorder) GetOwnByID(ctx, chatID, creatorID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOwnByID", reflect.TypeOf((*MockChatRepository)(nil).GetOwnByID), ctx, chatID, creatorID)
}

// List mocks base method.
func (m *MockChatRepository) List(ctx context.Context, memberID string) ([]domain.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, memberID)
	ret0, _ := ret[0].([]domain.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockChatRepositoryMockRecorder) List(ctx, memberID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockChatRepository)(nil).List), ctx, memberID)
}

// Update mocks base method.
func (m *MockChatRepository) Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, dto)
	ret0, _ := ret[0].(domain.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockChatRepositoryMockRecorder) Update(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockChatRepository)(nil).Update), ctx, dto)
}

// MockChatMemberRepository is a mock of ChatMemberRepository interface.
type MockChatMemberRepository struct {
	ctrl     *gomock.Controller
	recorder *MockChatMemberRepositoryMockRecorder
}

// MockChatMemberRepositoryMockRecorder is the mock recorder for MockChatMemberRepository.
type MockChatMemberRepositoryMockRecorder struct {
	mock *MockChatMemberRepository
}

// NewMockChatMemberRepository creates a new mock instance.
func NewMockChatMemberRepository(ctrl *gomock.Controller) *MockChatMemberRepository {
	mock := &MockChatMemberRepository{ctrl: ctrl}
	mock.recorder = &MockChatMemberRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChatMemberRepository) EXPECT() *MockChatMemberRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockChatMemberRepository) Create(ctx context.Context, userID, chatID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, userID, chatID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockChatMemberRepositoryMockRecorder) Create(ctx, userID, chatID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockChatMemberRepository)(nil).Create), ctx, userID, chatID)
}

// IsMemberInChat mocks base method.
func (m *MockChatMemberRepository) IsMemberInChat(ctx context.Context, userID, chatID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsMemberInChat", ctx, userID, chatID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsMemberInChat indicates an expected call of IsMemberInChat.
func (mr *MockChatMemberRepositoryMockRecorder) IsMemberInChat(ctx, userID, chatID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsMemberInChat", reflect.TypeOf((*MockChatMemberRepository)(nil).IsMemberInChat), ctx, userID, chatID)
}

// ListMembersInChat mocks base method.
func (m *MockChatMemberRepository) ListMembersInChat(ctx context.Context, chatID string) ([]domain.ChatMember, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMembersInChat", ctx, chatID)
	ret0, _ := ret[0].([]domain.ChatMember)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMembersInChat indicates an expected call of ListMembersInChat.
func (mr *MockChatMemberRepositoryMockRecorder) ListMembersInChat(ctx, chatID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMembersInChat", reflect.TypeOf((*MockChatMemberRepository)(nil).ListMembersInChat), ctx, chatID)
}

// MockMessageRepository is a mock of MessageRepository interface.
type MockMessageRepository struct {
	ctrl     *gomock.Controller
	recorder *MockMessageRepositoryMockRecorder
}

// MockMessageRepositoryMockRecorder is the mock recorder for MockMessageRepository.
type MockMessageRepositoryMockRecorder struct {
	mock *MockMessageRepository
}

// NewMockMessageRepository creates a new mock instance.
func NewMockMessageRepository(ctrl *gomock.Controller) *MockMessageRepository {
	mock := &MockMessageRepository{ctrl: ctrl}
	mock.recorder = &MockMessageRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageRepository) EXPECT() *MockMessageRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockMessageRepository) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, dto)
	ret0, _ := ret[0].(domain.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockMessageRepositoryMockRecorder) Create(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockMessageRepository)(nil).Create), ctx, dto)
}

// List mocks base method.
func (m *MockMessageRepository) List(ctx context.Context, chatID string, timestamp time.Time) ([]domain.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, chatID, timestamp)
	ret0, _ := ret[0].([]domain.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockMessageRepositoryMockRecorder) List(ctx, chatID, timestamp interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockMessageRepository)(nil).List), ctx, chatID, timestamp)
}

// MockMessagePubSub is a mock of MessagePubSub interface.
type MockMessagePubSub struct {
	ctrl     *gomock.Controller
	recorder *MockMessagePubSubMockRecorder
}

// MockMessagePubSubMockRecorder is the mock recorder for MockMessagePubSub.
type MockMessagePubSubMockRecorder struct {
	mock *MockMessagePubSub
}

// NewMockMessagePubSub creates a new mock instance.
func NewMockMessagePubSub(ctrl *gomock.Controller) *MockMessagePubSub {
	mock := &MockMessagePubSub{ctrl: ctrl}
	mock.recorder = &MockMessagePubSubMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessagePubSub) EXPECT() *MockMessagePubSubMockRecorder {
	return m.recorder
}

// Publish mocks base method.
func (m *MockMessagePubSub) Publish(ctx context.Context, message domain.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", ctx, message)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish.
func (mr *MockMessagePubSubMockRecorder) Publish(ctx, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockMessagePubSub)(nil).Publish), ctx, message)
}

// Subscribe mocks base method.
func (m *MockMessagePubSub) Subscribe(ctx context.Context, chatIDs ...string) repository.MessageSubscriber {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range chatIDs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(repository.MessageSubscriber)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockMessagePubSubMockRecorder) Subscribe(ctx interface{}, chatIDs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, chatIDs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockMessagePubSub)(nil).Subscribe), varargs...)
}

// MockMessageSubscriber is a mock of MessageSubscriber interface.
type MockMessageSubscriber struct {
	ctrl     *gomock.Controller
	recorder *MockMessageSubscriberMockRecorder
}

// MockMessageSubscriberMockRecorder is the mock recorder for MockMessageSubscriber.
type MockMessageSubscriberMockRecorder struct {
	mock *MockMessageSubscriber
}

// NewMockMessageSubscriber creates a new mock instance.
func NewMockMessageSubscriber(ctrl *gomock.Controller) *MockMessageSubscriber {
	mock := &MockMessageSubscriber{ctrl: ctrl}
	mock.recorder = &MockMessageSubscriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageSubscriber) EXPECT() *MockMessageSubscriberMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockMessageSubscriber) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockMessageSubscriberMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockMessageSubscriber)(nil).Close))
}

// MessageChannel mocks base method.
func (m *MockMessageSubscriber) MessageChannel() <-chan domain.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MessageChannel")
	ret0, _ := ret[0].(<-chan domain.Message)
	return ret0
}

// MessageChannel indicates an expected call of MessageChannel.
func (mr *MockMessageSubscriberMockRecorder) MessageChannel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MessageChannel", reflect.TypeOf((*MockMessageSubscriber)(nil).MessageChannel))
}

// ReceiveMessage mocks base method.
func (m *MockMessageSubscriber) ReceiveMessage(ctx context.Context) (domain.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReceiveMessage", ctx)
	ret0, _ := ret[0].(domain.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReceiveMessage indicates an expected call of ReceiveMessage.
func (mr *MockMessageSubscriberMockRecorder) ReceiveMessage(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReceiveMessage", reflect.TypeOf((*MockMessageSubscriber)(nil).ReceiveMessage), ctx)
}

// Subscribe mocks base method.
func (m *MockMessageSubscriber) Subscribe(ctx context.Context, chatIDs ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range chatIDs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockMessageSubscriberMockRecorder) Subscribe(ctx interface{}, chatIDs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, chatIDs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockMessageSubscriber)(nil).Subscribe), varargs...)
}

// Unsubscribe mocks base method.
func (m *MockMessageSubscriber) Unsubscribe(ctx context.Context, chatIDs ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range chatIDs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Unsubscribe", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe.
func (mr *MockMessageSubscriberMockRecorder) Unsubscribe(ctx interface{}, chatIDs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, chatIDs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockMessageSubscriber)(nil).Unsubscribe), varargs...)
}
