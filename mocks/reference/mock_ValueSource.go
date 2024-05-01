// Code generated by mockery v2.42.2. DO NOT EDIT.

package reference

import (
	data "github.com/lukasjarosch/skipper/v1/data"
	mock "github.com/stretchr/testify/mock"
)

// MockValueSource is an autogenerated mock type for the ValueSource type
type MockValueSource struct {
	mock.Mock
}

type MockValueSource_Expecter struct {
	mock *mock.Mock
}

func (_m *MockValueSource) EXPECT() *MockValueSource_Expecter {
	return &MockValueSource_Expecter{mock: &_m.Mock}
}

// AbsolutePath provides a mock function with given fields: _a0, _a1
func (_m *MockValueSource) AbsolutePath(_a0 data.Path, _a1 data.Path) (data.Path, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for AbsolutePath")
	}

	var r0 data.Path
	var r1 error
	if rf, ok := ret.Get(0).(func(data.Path, data.Path) (data.Path, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(data.Path, data.Path) data.Path); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(data.Path)
		}
	}

	if rf, ok := ret.Get(1).(func(data.Path, data.Path) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockValueSource_AbsolutePath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AbsolutePath'
type MockValueSource_AbsolutePath_Call struct {
	*mock.Call
}

// AbsolutePath is a helper method to define mock.On call
//   - _a0 data.Path
//   - _a1 data.Path
func (_e *MockValueSource_Expecter) AbsolutePath(_a0 interface{}, _a1 interface{}) *MockValueSource_AbsolutePath_Call {
	return &MockValueSource_AbsolutePath_Call{Call: _e.mock.On("AbsolutePath", _a0, _a1)}
}

func (_c *MockValueSource_AbsolutePath_Call) Run(run func(_a0 data.Path, _a1 data.Path)) *MockValueSource_AbsolutePath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path), args[1].(data.Path))
	})
	return _c
}

func (_c *MockValueSource_AbsolutePath_Call) Return(_a0 data.Path, _a1 error) *MockValueSource_AbsolutePath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockValueSource_AbsolutePath_Call) RunAndReturn(run func(data.Path, data.Path) (data.Path, error)) *MockValueSource_AbsolutePath_Call {
	_c.Call.Return(run)
	return _c
}

// Values provides a mock function with given fields:
func (_m *MockValueSource) Values() map[string]data.Value {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Values")
	}

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

// MockValueSource_Values_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Values'
type MockValueSource_Values_Call struct {
	*mock.Call
}

// Values is a helper method to define mock.On call
func (_e *MockValueSource_Expecter) Values() *MockValueSource_Values_Call {
	return &MockValueSource_Values_Call{Call: _e.mock.On("Values")}
}

func (_c *MockValueSource_Values_Call) Run(run func()) *MockValueSource_Values_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockValueSource_Values_Call) Return(_a0 map[string]data.Value) *MockValueSource_Values_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockValueSource_Values_Call) RunAndReturn(run func() map[string]data.Value) *MockValueSource_Values_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockValueSource creates a new instance of MockValueSource. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockValueSource(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockValueSource {
	mock := &MockValueSource{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
