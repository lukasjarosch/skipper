// Code generated by mockery v2.34.2. DO NOT EDIT.

package skipper

import (
	data "github.com/lukasjarosch/skipper/data"
	mock "github.com/stretchr/testify/mock"
)

// MockReferenceValueSource is an autogenerated mock type for the ReferenceValueSource type
type MockReferenceValueSource struct {
	mock.Mock
}

type MockReferenceValueSource_Expecter struct {
	mock *mock.Mock
}

func (_m *MockReferenceValueSource) EXPECT() *MockReferenceValueSource_Expecter {
	return &MockReferenceValueSource_Expecter{mock: &_m.Mock}
}

// AbsolutePath provides a mock function with given fields: path, context
func (_m *MockReferenceValueSource) AbsolutePath(path data.Path, context data.Path) (data.Path, error) {
	ret := _m.Called(path, context)

	var r0 data.Path
	var r1 error
	if rf, ok := ret.Get(0).(func(data.Path, data.Path) (data.Path, error)); ok {
		return rf(path, context)
	}
	if rf, ok := ret.Get(0).(func(data.Path, data.Path) data.Path); ok {
		r0 = rf(path, context)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(data.Path)
		}
	}

	if rf, ok := ret.Get(1).(func(data.Path, data.Path) error); ok {
		r1 = rf(path, context)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockReferenceValueSource_AbsolutePath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AbsolutePath'
type MockReferenceValueSource_AbsolutePath_Call struct {
	*mock.Call
}

// AbsolutePath is a helper method to define mock.On call
//   - path data.Path
//   - context data.Path
func (_e *MockReferenceValueSource_Expecter) AbsolutePath(path interface{}, context interface{}) *MockReferenceValueSource_AbsolutePath_Call {
	return &MockReferenceValueSource_AbsolutePath_Call{Call: _e.mock.On("AbsolutePath", path, context)}
}

func (_c *MockReferenceValueSource_AbsolutePath_Call) Run(run func(path data.Path, context data.Path)) *MockReferenceValueSource_AbsolutePath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path), args[1].(data.Path))
	})
	return _c
}

func (_c *MockReferenceValueSource_AbsolutePath_Call) Return(_a0 data.Path, _a1 error) *MockReferenceValueSource_AbsolutePath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockReferenceValueSource_AbsolutePath_Call) RunAndReturn(run func(data.Path, data.Path) (data.Path, error)) *MockReferenceValueSource_AbsolutePath_Call {
	_c.Call.Return(run)
	return _c
}

// GetPath provides a mock function with given fields: _a0
func (_m *MockReferenceValueSource) GetPath(_a0 data.Path) (data.Value, error) {
	ret := _m.Called(_a0)

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

// MockReferenceValueSource_GetPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPath'
type MockReferenceValueSource_GetPath_Call struct {
	*mock.Call
}

// GetPath is a helper method to define mock.On call
//   - _a0 data.Path
func (_e *MockReferenceValueSource_Expecter) GetPath(_a0 interface{}) *MockReferenceValueSource_GetPath_Call {
	return &MockReferenceValueSource_GetPath_Call{Call: _e.mock.On("GetPath", _a0)}
}

func (_c *MockReferenceValueSource_GetPath_Call) Run(run func(_a0 data.Path)) *MockReferenceValueSource_GetPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path))
	})
	return _c
}

func (_c *MockReferenceValueSource_GetPath_Call) Return(_a0 data.Value, _a1 error) *MockReferenceValueSource_GetPath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockReferenceValueSource_GetPath_Call) RunAndReturn(run func(data.Path) (data.Value, error)) *MockReferenceValueSource_GetPath_Call {
	_c.Call.Return(run)
	return _c
}

// Values provides a mock function with given fields:
func (_m *MockReferenceValueSource) Values() map[string]data.Value {
	ret := _m.Called()

	var r0 map[string]data.Value
	if rf, ok := ret.Get(0).(func() map[string]data.Value); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]data.Value)
		}
	}

	return r0
}

// MockReferenceValueSource_Values_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Values'
type MockReferenceValueSource_Values_Call struct {
	*mock.Call
}

// Values is a helper method to define mock.On call
func (_e *MockReferenceValueSource_Expecter) Values() *MockReferenceValueSource_Values_Call {
	return &MockReferenceValueSource_Values_Call{Call: _e.mock.On("Values")}
}

func (_c *MockReferenceValueSource_Values_Call) Run(run func()) *MockReferenceValueSource_Values_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockReferenceValueSource_Values_Call) Return(_a0 map[string]data.Value) *MockReferenceValueSource_Values_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockReferenceValueSource_Values_Call) RunAndReturn(run func() map[string]data.Value) *MockReferenceValueSource_Values_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockReferenceValueSource creates a new instance of MockReferenceValueSource. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockReferenceValueSource(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReferenceValueSource {
	mock := &MockReferenceValueSource{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}