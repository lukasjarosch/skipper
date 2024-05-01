// Code generated by mockery v2.42.2. DO NOT EDIT.

package skipper

import (
	data "github.com/lukasjarosch/skipper/v1/data"
	mock "github.com/stretchr/testify/mock"
)

// MockDataSetter is an autogenerated mock type for the DataSetter type
type MockDataSetter struct {
	mock.Mock
}

type MockDataSetter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDataSetter) EXPECT() *MockDataSetter_Expecter {
	return &MockDataSetter_Expecter{mock: &_m.Mock}
}

// SetPath provides a mock function with given fields: _a0, _a1
func (_m *MockDataSetter) SetPath(_a0 data.Path, _a1 interface{}) error {
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

// MockDataSetter_SetPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetPath'
type MockDataSetter_SetPath_Call struct {
	*mock.Call
}

// SetPath is a helper method to define mock.On call
//   - _a0 data.Path
//   - _a1 interface{}
func (_e *MockDataSetter_Expecter) SetPath(_a0 interface{}, _a1 interface{}) *MockDataSetter_SetPath_Call {
	return &MockDataSetter_SetPath_Call{Call: _e.mock.On("SetPath", _a0, _a1)}
}

func (_c *MockDataSetter_SetPath_Call) Run(run func(_a0 data.Path, _a1 interface{})) *MockDataSetter_SetPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path), args[1].(interface{}))
	})
	return _c
}

func (_c *MockDataSetter_SetPath_Call) Return(_a0 error) *MockDataSetter_SetPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDataSetter_SetPath_Call) RunAndReturn(run func(data.Path, interface{}) error) *MockDataSetter_SetPath_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDataSetter creates a new instance of MockDataSetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDataSetter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDataSetter {
	mock := &MockDataSetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
