// Code generated by mockery v2.40.1. DO NOT EDIT.

package service

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockTransactionManager is an autogenerated mock type for the TransactionManager type
type MockTransactionManager struct {
	mock.Mock
}

// Do provides a mock function with given fields: ctx, fn
func (_m *MockTransactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
	ret := _m.Called(ctx, fn)

	if len(ret) == 0 {
		panic("no return value specified for Do")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, func(context.Context) error) error); ok {
		r0 = rf(ctx, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockTransactionManager creates a new instance of MockTransactionManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTransactionManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTransactionManager {
	mock := &MockTransactionManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}