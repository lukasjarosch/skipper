// Code generated by mockery v2.42.2. DO NOT EDIT.

package skipper

import (
	skipper "github.com/lukasjarosch/skipper"
	mock "github.com/stretchr/testify/mock"
)

// MockClassLoaderFunc is an autogenerated mock type for the ClassLoaderFunc type
type MockClassLoaderFunc struct {
	mock.Mock
}

type MockClassLoaderFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClassLoaderFunc) EXPECT() *MockClassLoaderFunc_Expecter {
	return &MockClassLoaderFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: filePaths
func (_m *MockClassLoaderFunc) Execute(filePaths []string) ([]*skipper.Class, error) {
	ret := _m.Called(filePaths)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 []*skipper.Class
	var r1 error
	if rf, ok := ret.Get(0).(func([]string) ([]*skipper.Class, error)); ok {
		return rf(filePaths)
	}
	if rf, ok := ret.Get(0).(func([]string) []*skipper.Class); ok {
		r0 = rf(filePaths)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*skipper.Class)
		}
	}

	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(filePaths)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClassLoaderFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockClassLoaderFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - filePaths []string
func (_e *MockClassLoaderFunc_Expecter) Execute(filePaths interface{}) *MockClassLoaderFunc_Execute_Call {
	return &MockClassLoaderFunc_Execute_Call{Call: _e.mock.On("Execute", filePaths)}
}

func (_c *MockClassLoaderFunc_Execute_Call) Run(run func(filePaths []string)) *MockClassLoaderFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]string))
	})
	return _c
}

func (_c *MockClassLoaderFunc_Execute_Call) Return(_a0 []*skipper.Class, _a1 error) *MockClassLoaderFunc_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClassLoaderFunc_Execute_Call) RunAndReturn(run func([]string) ([]*skipper.Class, error)) *MockClassLoaderFunc_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockClassLoaderFunc creates a new instance of MockClassLoaderFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClassLoaderFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockClassLoaderFunc {
	mock := &MockClassLoaderFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
