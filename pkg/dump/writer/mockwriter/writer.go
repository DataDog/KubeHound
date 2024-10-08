// Code generated by mockery v2.43.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// DumperWriter is an autogenerated mock type for the DumperWriter type
type DumperWriter struct {
	mock.Mock
}

type DumperWriter_Expecter struct {
	mock *mock.Mock
}

func (_m *DumperWriter) EXPECT() *DumperWriter_Expecter {
	return &DumperWriter_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields: _a0
func (_m *DumperWriter) Close(_a0 context.Context) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DumperWriter_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type DumperWriter_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *DumperWriter_Expecter) Close(_a0 interface{}) *DumperWriter_Close_Call {
	return &DumperWriter_Close_Call{Call: _e.mock.On("Close", _a0)}
}

func (_c *DumperWriter_Close_Call) Run(run func(_a0 context.Context)) *DumperWriter_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *DumperWriter_Close_Call) Return(_a0 error) *DumperWriter_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DumperWriter_Close_Call) RunAndReturn(run func(context.Context) error) *DumperWriter_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Flush provides a mock function with given fields: _a0
func (_m *DumperWriter) Flush(_a0 context.Context) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Flush")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DumperWriter_Flush_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Flush'
type DumperWriter_Flush_Call struct {
	*mock.Call
}

// Flush is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *DumperWriter_Expecter) Flush(_a0 interface{}) *DumperWriter_Flush_Call {
	return &DumperWriter_Flush_Call{Call: _e.mock.On("Flush", _a0)}
}

func (_c *DumperWriter_Flush_Call) Run(run func(_a0 context.Context)) *DumperWriter_Flush_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *DumperWriter_Flush_Call) Return(_a0 error) *DumperWriter_Flush_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DumperWriter_Flush_Call) RunAndReturn(run func(context.Context) error) *DumperWriter_Flush_Call {
	_c.Call.Return(run)
	return _c
}

// OutputPath provides a mock function with given fields:
func (_m *DumperWriter) OutputPath() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OutputPath")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// DumperWriter_OutputPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OutputPath'
type DumperWriter_OutputPath_Call struct {
	*mock.Call
}

// OutputPath is a helper method to define mock.On call
func (_e *DumperWriter_Expecter) OutputPath() *DumperWriter_OutputPath_Call {
	return &DumperWriter_OutputPath_Call{Call: _e.mock.On("OutputPath")}
}

func (_c *DumperWriter_OutputPath_Call) Run(run func()) *DumperWriter_OutputPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DumperWriter_OutputPath_Call) Return(_a0 string) *DumperWriter_OutputPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DumperWriter_OutputPath_Call) RunAndReturn(run func() string) *DumperWriter_OutputPath_Call {
	_c.Call.Return(run)
	return _c
}

// WorkerNumber provides a mock function with given fields:
func (_m *DumperWriter) WorkerNumber() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for WorkerNumber")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// DumperWriter_WorkerNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WorkerNumber'
type DumperWriter_WorkerNumber_Call struct {
	*mock.Call
}

// WorkerNumber is a helper method to define mock.On call
func (_e *DumperWriter_Expecter) WorkerNumber() *DumperWriter_WorkerNumber_Call {
	return &DumperWriter_WorkerNumber_Call{Call: _e.mock.On("WorkerNumber")}
}

func (_c *DumperWriter_WorkerNumber_Call) Run(run func()) *DumperWriter_WorkerNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DumperWriter_WorkerNumber_Call) Return(_a0 int) *DumperWriter_WorkerNumber_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DumperWriter_WorkerNumber_Call) RunAndReturn(run func() int) *DumperWriter_WorkerNumber_Call {
	_c.Call.Return(run)
	return _c
}

// Write provides a mock function with given fields: _a0, _a1, _a2
func (_m *DumperWriter) Write(_a0 context.Context, _a1 []byte, _a2 string) error {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for Write")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte, string) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DumperWriter_Write_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Write'
type DumperWriter_Write_Call struct {
	*mock.Call
}

// Write is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 []byte
//   - _a2 string
func (_e *DumperWriter_Expecter) Write(_a0 interface{}, _a1 interface{}, _a2 interface{}) *DumperWriter_Write_Call {
	return &DumperWriter_Write_Call{Call: _e.mock.On("Write", _a0, _a1, _a2)}
}

func (_c *DumperWriter_Write_Call) Run(run func(_a0 context.Context, _a1 []byte, _a2 string)) *DumperWriter_Write_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]byte), args[2].(string))
	})
	return _c
}

func (_c *DumperWriter_Write_Call) Return(_a0 error) *DumperWriter_Write_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DumperWriter_Write_Call) RunAndReturn(run func(context.Context, []byte, string) error) *DumperWriter_Write_Call {
	_c.Call.Return(run)
	return _c
}

// WriteMetadata provides a mock function with given fields: _a0
func (_m *DumperWriter) WriteMetadata(_a0 context.Context) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for WriteMetadata")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DumperWriter_WriteMetadata_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteMetadata'
type DumperWriter_WriteMetadata_Call struct {
	*mock.Call
}

// WriteMetadata is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *DumperWriter_Expecter) WriteMetadata(_a0 interface{}) *DumperWriter_WriteMetadata_Call {
	return &DumperWriter_WriteMetadata_Call{Call: _e.mock.On("WriteMetadata", _a0)}
}

func (_c *DumperWriter_WriteMetadata_Call) Run(run func(_a0 context.Context)) *DumperWriter_WriteMetadata_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *DumperWriter_WriteMetadata_Call) Return(_a0 error) *DumperWriter_WriteMetadata_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DumperWriter_WriteMetadata_Call) RunAndReturn(run func(context.Context) error) *DumperWriter_WriteMetadata_Call {
	_c.Call.Return(run)
	return _c
}

// NewDumperWriter creates a new instance of DumperWriter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDumperWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *DumperWriter {
	mock := &DumperWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
