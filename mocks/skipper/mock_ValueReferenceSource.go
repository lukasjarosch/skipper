// Code generated by mockery v2.42.2. DO NOT EDIT.

package skipper

import (
	data "github.com/lukasjarosch/skipper/data"
	mock "github.com/stretchr/testify/mock"

	skipper "github.com/lukasjarosch/skipper"
)

// MockValueReferenceSource is an autogenerated mock type for the ValueReferenceSource type
type MockValueReferenceSource struct {
	mock.Mock
}

type MockValueReferenceSource_Expecter struct {
	mock *mock.Mock
}

func (_m *MockValueReferenceSource) EXPECT() *MockValueReferenceSource_Expecter {
	return &MockValueReferenceSource_Expecter{mock: &_m.Mock}
}

// AbsolutePath provides a mock function with given fields: _a0, _a1
func (_m *MockValueReferenceSource) AbsolutePath(_a0 data.Path, _a1 data.Path) (data.Path, error) {
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

// MockValueReferenceSource_AbsolutePath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AbsolutePath'
type MockValueReferenceSource_AbsolutePath_Call struct {
	*mock.Call
}

// AbsolutePath is a helper method to define mock.On call
//   - _a0 data.Path
//   - _a1 data.Path
func (_e *MockValueReferenceSource_Expecter) AbsolutePath(_a0 interface{}, _a1 interface{}) *MockValueReferenceSource_AbsolutePath_Call {
	return &MockValueReferenceSource_AbsolutePath_Call{Call: _e.mock.On("AbsolutePath", _a0, _a1)}
}

func (_c *MockValueReferenceSource_AbsolutePath_Call) Run(run func(_a0 data.Path, _a1 data.Path)) *MockValueReferenceSource_AbsolutePath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path), args[1].(data.Path))
	})
	return _c
}

func (_c *MockValueReferenceSource_AbsolutePath_Call) Return(_a0 data.Path, _a1 error) *MockValueReferenceSource_AbsolutePath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockValueReferenceSource_AbsolutePath_Call) RunAndReturn(run func(data.Path, data.Path) (data.Path, error)) *MockValueReferenceSource_AbsolutePath_Call {
	_c.Call.Return(run)
	return _c
}

// GetPath provides a mock function with given fields: _a0
func (_m *MockValueReferenceSource) GetPath(_a0 data.Path) (data.Value, error) {
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

// MockValueReferenceSource_GetPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPath'
type MockValueReferenceSource_GetPath_Call struct {
	*mock.Call
}

// GetPath is a helper method to define mock.On call
//   - _a0 data.Path
func (_e *MockValueReferenceSource_Expecter) GetPath(_a0 interface{}) *MockValueReferenceSource_GetPath_Call {
	return &MockValueReferenceSource_GetPath_Call{Call: _e.mock.On("GetPath", _a0)}
}

func (_c *MockValueReferenceSource_GetPath_Call) Run(run func(_a0 data.Path)) *MockValueReferenceSource_GetPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path))
	})
	return _c
}

func (_c *MockValueReferenceSource_GetPath_Call) Return(_a0 data.Value, _a1 error) *MockValueReferenceSource_GetPath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockValueReferenceSource_GetPath_Call) RunAndReturn(run func(data.Path) (data.Value, error)) *MockValueReferenceSource_GetPath_Call {
	_c.Call.Return(run)
	return _c
}

// RegisterPostSetHook provides a mock function with given fields: _a0
func (_m *MockValueReferenceSource) RegisterPostSetHook(_a0 skipper.SetHookFunc) {
	_m.Called(_a0)
}

// MockValueReferenceSource_RegisterPostSetHook_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RegisterPostSetHook'
type MockValueReferenceSource_RegisterPostSetHook_Call struct {
	*mock.Call
}

// RegisterPostSetHook is a helper method to define mock.On call
//   - _a0 skipper.SetHookFunc
func (_e *MockValueReferenceSource_Expecter) RegisterPostSetHook(_a0 interface{}) *MockValueReferenceSource_RegisterPostSetHook_Call {
	return &MockValueReferenceSource_RegisterPostSetHook_Call{Call: _e.mock.On("RegisterPostSetHook", _a0)}
}

func (_c *MockValueReferenceSource_RegisterPostSetHook_Call) Run(run func(_a0 skipper.SetHookFunc)) *MockValueReferenceSource_RegisterPostSetHook_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(skipper.SetHookFunc))
	})
	return _c
}

func (_c *MockValueReferenceSource_RegisterPostSetHook_Call) Return() *MockValueReferenceSource_RegisterPostSetHook_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockValueReferenceSource_RegisterPostSetHook_Call) RunAndReturn(run func(skipper.SetHookFunc)) *MockValueReferenceSource_RegisterPostSetHook_Call {
	_c.Call.Return(run)
	return _c
}

// SetPath provides a mock function with given fields: _a0, _a1
func (_m *MockValueReferenceSource) SetPath(_a0 data.Path, _a1 interface{}) error {
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

// MockValueReferenceSource_SetPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetPath'
type MockValueReferenceSource_SetPath_Call struct {
	*mock.Call
}

// SetPath is a helper method to define mock.On call
//   - _a0 data.Path
//   - _a1 interface{}
func (_e *MockValueReferenceSource_Expecter) SetPath(_a0 interface{}, _a1 interface{}) *MockValueReferenceSource_SetPath_Call {
	return &MockValueReferenceSource_SetPath_Call{Call: _e.mock.On("SetPath", _a0, _a1)}
}

func (_c *MockValueReferenceSource_SetPath_Call) Run(run func(_a0 data.Path, _a1 interface{})) *MockValueReferenceSource_SetPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(data.Path), args[1].(interface{}))
	})
	return _c
}

func (_c *MockValueReferenceSource_SetPath_Call) Return(_a0 error) *MockValueReferenceSource_SetPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockValueReferenceSource_SetPath_Call) RunAndReturn(run func(data.Path, interface{}) error) *MockValueReferenceSource_SetPath_Call {
	_c.Call.Return(run)
	return _c
}

// Values provides a mock function with given fields:
func (_m *MockValueReferenceSource) Values() map[string]data.Value {
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

// MockValueReferenceSource_Values_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Values'
type MockValueReferenceSource_Values_Call struct {
	*mock.Call
}

// Values is a helper method to define mock.On call
func (_e *MockValueReferenceSource_Expecter) Values() *MockValueReferenceSource_Values_Call {
	return &MockValueReferenceSource_Values_Call{Call: _e.mock.On("Values")}
}

func (_c *MockValueReferenceSource_Values_Call) Run(run func()) *MockValueReferenceSource_Values_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockValueReferenceSource_Values_Call) Return(_a0 map[string]data.Value) *MockValueReferenceSource_Values_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockValueReferenceSource_Values_Call) RunAndReturn(run func() map[string]data.Value) *MockValueReferenceSource_Values_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockValueReferenceSource creates a new instance of MockValueReferenceSource. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockValueReferenceSource(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockValueReferenceSource {
	mock := &MockValueReferenceSource{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
