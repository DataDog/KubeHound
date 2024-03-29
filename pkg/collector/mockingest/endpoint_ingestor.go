// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	context "context"

	types "github.com/DataDog/KubeHound/pkg/globals/types"
	mock "github.com/stretchr/testify/mock"
)

// EndpointIngestor is an autogenerated mock type for the EndpointIngestor type
type EndpointIngestor struct {
	mock.Mock
}

type EndpointIngestor_Expecter struct {
	mock *mock.Mock
}

func (_m *EndpointIngestor) EXPECT() *EndpointIngestor_Expecter {
	return &EndpointIngestor_Expecter{mock: &_m.Mock}
}

// Complete provides a mock function with given fields: _a0
func (_m *EndpointIngestor) Complete(_a0 context.Context) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EndpointIngestor_Complete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Complete'
type EndpointIngestor_Complete_Call struct {
	*mock.Call
}

// Complete is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *EndpointIngestor_Expecter) Complete(_a0 interface{}) *EndpointIngestor_Complete_Call {
	return &EndpointIngestor_Complete_Call{Call: _e.mock.On("Complete", _a0)}
}

func (_c *EndpointIngestor_Complete_Call) Run(run func(_a0 context.Context)) *EndpointIngestor_Complete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *EndpointIngestor_Complete_Call) Return(_a0 error) *EndpointIngestor_Complete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EndpointIngestor_Complete_Call) RunAndReturn(run func(context.Context) error) *EndpointIngestor_Complete_Call {
	_c.Call.Return(run)
	return _c
}

// IngestEndpoint provides a mock function with given fields: _a0, _a1
func (_m *EndpointIngestor) IngestEndpoint(_a0 context.Context, _a1 types.EndpointType) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.EndpointType) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EndpointIngestor_IngestEndpoint_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IngestEndpoint'
type EndpointIngestor_IngestEndpoint_Call struct {
	*mock.Call
}

// IngestEndpoint is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 types.EndpointType
func (_e *EndpointIngestor_Expecter) IngestEndpoint(_a0 interface{}, _a1 interface{}) *EndpointIngestor_IngestEndpoint_Call {
	return &EndpointIngestor_IngestEndpoint_Call{Call: _e.mock.On("IngestEndpoint", _a0, _a1)}
}

func (_c *EndpointIngestor_IngestEndpoint_Call) Run(run func(_a0 context.Context, _a1 types.EndpointType)) *EndpointIngestor_IngestEndpoint_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(types.EndpointType))
	})
	return _c
}

func (_c *EndpointIngestor_IngestEndpoint_Call) Return(_a0 error) *EndpointIngestor_IngestEndpoint_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EndpointIngestor_IngestEndpoint_Call) RunAndReturn(run func(context.Context, types.EndpointType) error) *EndpointIngestor_IngestEndpoint_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewEndpointIngestor interface {
	mock.TestingT
	Cleanup(func())
}

// NewEndpointIngestor creates a new instance of EndpointIngestor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEndpointIngestor(t mockConstructorTestingTNewEndpointIngestor) *EndpointIngestor {
	mock := &EndpointIngestor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
