// Code generated by mockery v2.42.2. DO NOT EDIT.

package reference

import (
	data "github.com/lukasjarosch/skipper/data"
	mock "github.com/stretchr/testify/mock"
)

// MockValueTarget is an autogenerated mock type for the ValueTarget type
type MockValueTarget struct {
	mock.Mock
}

type MockValueTarget_Expecter struct {
	mock *mock.Mock
}

func (_m *MockValueTarget) EXPECT() *MockValueTarget_Expecter {
	return &MockValueTarget_Expecter{mock: &_m.Mock}
}

// GetPath provides a mock function with given fields: _a0
func (_m *MockValueTarget) GetPath(_a0 data.Path) (data.Value, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetPath")
	}

	var r0 data.Value
	var r1 error
	if rf, ok := ret.Get(0).(func(data.Path) (data.Value, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(data.Path) data.Value); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(data.Value)
	}

	if rf, ok := ret.Get(1).(func(data.Path) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockValueTarget_GetPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPath'
type MockValueTarget_GetPath_Call struct {
	*mock.Call
}

// GetPath is a helper method to define mock.On call
//   - _a0 data.Path
func (_e *MockValueTarget_Expecter) GetPath(_a0 interface{}) *MockValueTarget_GetPath_Call {
	return &MockValueTarget_GetPath_Call{Call: _e.mock.On("GetPath", _a0)}
}

func (_c *MockValueTarget_GetPath_Call) Run(run func(_a0 data.Path)) *MockValueTarget_GetPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path))
	})
	return _c
}

func (_c *MockValueTarget_GetPath_Call) Return(_a0 data.Value, _a1 error) *MockValueTarget_GetPath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockValueTarget_GetPath_Call) RunAndReturn(run func(data.Path) (data.Value, error)) *MockValueTarget_GetPath_Call {
	_c.Call.Return(run)
	return _c
}

// SetPath provides a mock function with given fields: _a0, _a1
func (_m *MockValueTarget) SetPath(_a0 data.Path, _a1 interface{}) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SetPath")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(data.Path, interface{}) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockValueTarget_SetPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetPath'
type MockValueTarget_SetPath_Call struct {
	*mock.Call
}

// SetPath is a helper method to define mock.On call
//   - _a0 data.Path
//   - _a1 interface{}
func (_e *MockValueTarget_Expecter) SetPath(_a0 interface{}, _a1 interface{}) *MockValueTarget_SetPath_Call {
	return &MockValueTarget_SetPath_Call{Call: _e.mock.On("SetPath", _a0, _a1)}
}

func (_c *MockValueTarget_SetPath_Call) Run(run func(_a0 data.Path, _a1 interface{})) *MockValueTarget_SetPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path), args[1].(interface{}))
	})
	return _c
}

func (_c *MockValueTarget_SetPath_Call) Return(_a0 error) *MockValueTarget_SetPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockValueTarget_SetPath_Call) RunAndReturn(run func(data.Path, interface{}) error) *MockValueTarget_SetPath_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockValueTarget creates a new instance of MockValueTarget. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockValueTarget(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockValueTarget {
	mock := &MockValueTarget{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
