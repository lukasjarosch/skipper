// Code generated by mockery v2.42.2. DO NOT EDIT.

package skipper

import (
	skipper "github.com/lukasjarosch/skipper"
	mock "github.com/stretchr/testify/mock"
)

// MockRegisterClassHookFunc is an autogenerated mock type for the RegisterClassHookFunc type
type MockRegisterClassHookFunc struct {
	mock.Mock
}

type MockRegisterClassHookFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRegisterClassHookFunc) EXPECT() *MockRegisterClassHookFunc_Expecter {
	return &MockRegisterClassHookFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: class
func (_m *MockRegisterClassHookFunc) Execute(class *skipper.Class) error {
	ret := _m.Called(class)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*skipper.Class) error); ok {
		r0 = rf(class)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRegisterClassHookFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockRegisterClassHookFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - class *skipper.Class
func (_e *MockRegisterClassHookFunc_Expecter) Execute(class interface{}) *MockRegisterClassHookFunc_Execute_Call {
	return &MockRegisterClassHookFunc_Execute_Call{Call: _e.mock.On("Execute", class)}
}

func (_c *MockRegisterClassHookFunc_Execute_Call) Run(run func(class *skipper.Class)) *MockRegisterClassHookFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*skipper.Class))
	})
	return _c
}

func (_c *MockRegisterClassHookFunc_Execute_Call) Return(_a0 error) *MockRegisterClassHookFunc_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRegisterClassHookFunc_Execute_Call) RunAndReturn(run func(*skipper.Class) error) *MockRegisterClassHookFunc_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRegisterClassHookFunc creates a new instance of MockRegisterClassHookFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRegisterClassHookFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRegisterClassHookFunc {
	mock := &MockRegisterClassHookFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}