// Code generated by mockery v2.40.1. DO NOT EDIT.

package v1

import (
	context "context"

	dto "github.com/Chatyx/backend/internal/dto"
	entity "github.com/Chatyx/backend/internal/entity"

	mock "github.com/stretchr/testify/mock"
)

// MockMessageService is an autogenerated mock type for the MessageService type
type MockMessageService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, obj
func (_m *MockMessageService) Create(ctx context.Context, obj dto.MessageCreate) (entity.Message, error) {
	ret := _m.Called(ctx, obj)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 entity.Message
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, dto.MessageCreate) (entity.Message, error)); ok {
		return rf(ctx, obj)
	}
	if rf, ok := ret.Get(0).(func(context.Context, dto.MessageCreate) entity.Message); ok {
		r0 = rf(ctx, obj)
	} else {
		r0 = ret.Get(0).(entity.Message)
	}

	if rf, ok := ret.Get(1).(func(context.Context, dto.MessageCreate) error); ok {
		r1 = rf(ctx, obj)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, obj
func (_m *MockMessageService) List(ctx context.Context, obj dto.MessageList) ([]entity.Message, error) {
	ret := _m.Called(ctx, obj)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []entity.Message
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, dto.MessageList) ([]entity.Message, error)); ok {
		return rf(ctx, obj)
	}
	if rf, ok := ret.Get(0).(func(context.Context, dto.MessageList) []entity.Message); ok {
		r0 = rf(ctx, obj)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.Message)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, dto.MessageList) error); ok {
		r1 = rf(ctx, obj)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockMessageService creates a new instance of MockMessageService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMessageService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMessageService {
	mock := &MockMessageService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
