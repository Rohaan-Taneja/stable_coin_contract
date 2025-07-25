// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	evm "github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	mock "github.com/stretchr/testify/mock"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	sdk "github.com/smartcontractkit/cre-sdk-go/sdk"
)

// EVMClient is an autogenerated mock type for the EVMClient type
type EVMClient struct {
	mock.Mock
}

type EVMClient_Expecter struct {
	mock *mock.Mock
}

func (_m *EVMClient) EXPECT() *EVMClient_Expecter {
	return &EVMClient_Expecter{mock: &_m.Mock}
}

// CallContract provides a mock function with given fields: _a0, _a1
func (_m *EVMClient) CallContract(_a0 sdk.Runtime, _a1 *evm.CallContractRequest) sdk.Promise[*evm.CallContractReply] {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CallContract")
	}

	var r0 sdk.Promise[*evm.CallContractReply]
	if rf, ok := ret.Get(0).(func(sdk.Runtime, *evm.CallContractRequest) sdk.Promise[*evm.CallContractReply]); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sdk.Promise[*evm.CallContractReply])
		}
	}

	return r0
}

// EVMClient_CallContract_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CallContract'
type EVMClient_CallContract_Call struct {
	*mock.Call
}

// CallContract is a helper method to define mock.On call
//   - _a0 sdk.Runtime
//   - _a1 *evm.CallContractRequest
func (_e *EVMClient_Expecter) CallContract(_a0 interface{}, _a1 interface{}) *EVMClient_CallContract_Call {
	return &EVMClient_CallContract_Call{Call: _e.mock.On("CallContract", _a0, _a1)}
}

func (_c *EVMClient_CallContract_Call) Run(run func(_a0 sdk.Runtime, _a1 *evm.CallContractRequest)) *EVMClient_CallContract_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(sdk.Runtime), args[1].(*evm.CallContractRequest))
	})
	return _c
}

func (_c *EVMClient_CallContract_Call) Return(_a0 sdk.Promise[*evm.CallContractReply]) *EVMClient_CallContract_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EVMClient_CallContract_Call) RunAndReturn(run func(sdk.Runtime, *evm.CallContractRequest) sdk.Promise[*evm.CallContractReply]) *EVMClient_CallContract_Call {
	_c.Call.Return(run)
	return _c
}

// FilterLogs provides a mock function with given fields: _a0, _a1
func (_m *EVMClient) FilterLogs(_a0 sdk.Runtime, _a1 *evm.FilterLogsRequest) sdk.Promise[*evm.FilterLogsReply] {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for FilterLogs")
	}

	var r0 sdk.Promise[*evm.FilterLogsReply]
	if rf, ok := ret.Get(0).(func(sdk.Runtime, *evm.FilterLogsRequest) sdk.Promise[*evm.FilterLogsReply]); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sdk.Promise[*evm.FilterLogsReply])
		}
	}

	return r0
}

// EVMClient_FilterLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FilterLogs'
type EVMClient_FilterLogs_Call struct {
	*mock.Call
}

// FilterLogs is a helper method to define mock.On call
//   - _a0 sdk.Runtime
//   - _a1 *evm.FilterLogsRequest
func (_e *EVMClient_Expecter) FilterLogs(_a0 interface{}, _a1 interface{}) *EVMClient_FilterLogs_Call {
	return &EVMClient_FilterLogs_Call{Call: _e.mock.On("FilterLogs", _a0, _a1)}
}

func (_c *EVMClient_FilterLogs_Call) Run(run func(_a0 sdk.Runtime, _a1 *evm.FilterLogsRequest)) *EVMClient_FilterLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(sdk.Runtime), args[1].(*evm.FilterLogsRequest))
	})
	return _c
}

func (_c *EVMClient_FilterLogs_Call) Return(_a0 sdk.Promise[*evm.FilterLogsReply]) *EVMClient_FilterLogs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EVMClient_FilterLogs_Call) RunAndReturn(run func(sdk.Runtime, *evm.FilterLogsRequest) sdk.Promise[*evm.FilterLogsReply]) *EVMClient_FilterLogs_Call {
	_c.Call.Return(run)
	return _c
}

// LatestAndFinalizedHead provides a mock function with given fields: runtime, input
func (_m *EVMClient) LatestAndFinalizedHead(runtime sdk.Runtime, input *emptypb.Empty) sdk.Promise[*evm.LatestAndFinalizedHeadReply] {
	ret := _m.Called(runtime, input)

	if len(ret) == 0 {
		panic("no return value specified for LatestAndFinalizedHead")
	}

	var r0 sdk.Promise[*evm.LatestAndFinalizedHeadReply]
	if rf, ok := ret.Get(0).(func(sdk.Runtime, *emptypb.Empty) sdk.Promise[*evm.LatestAndFinalizedHeadReply]); ok {
		r0 = rf(runtime, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sdk.Promise[*evm.LatestAndFinalizedHeadReply])
		}
	}

	return r0
}

// EVMClient_LatestAndFinalizedHead_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LatestAndFinalizedHead'
type EVMClient_LatestAndFinalizedHead_Call struct {
	*mock.Call
}

// LatestAndFinalizedHead is a helper method to define mock.On call
//   - runtime sdk.Runtime
//   - input *emptypb.Empty
func (_e *EVMClient_Expecter) LatestAndFinalizedHead(runtime interface{}, input interface{}) *EVMClient_LatestAndFinalizedHead_Call {
	return &EVMClient_LatestAndFinalizedHead_Call{Call: _e.mock.On("LatestAndFinalizedHead", runtime, input)}
}

func (_c *EVMClient_LatestAndFinalizedHead_Call) Run(run func(runtime sdk.Runtime, input *emptypb.Empty)) *EVMClient_LatestAndFinalizedHead_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(sdk.Runtime), args[1].(*emptypb.Empty))
	})
	return _c
}

func (_c *EVMClient_LatestAndFinalizedHead_Call) Return(_a0 sdk.Promise[*evm.LatestAndFinalizedHeadReply]) *EVMClient_LatestAndFinalizedHead_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EVMClient_LatestAndFinalizedHead_Call) RunAndReturn(run func(sdk.Runtime, *emptypb.Empty) sdk.Promise[*evm.LatestAndFinalizedHeadReply]) *EVMClient_LatestAndFinalizedHead_Call {
	_c.Call.Return(run)
	return _c
}

// RegisterLogTracking provides a mock function with given fields: _a0, _a1
func (_m *EVMClient) RegisterLogTracking(_a0 sdk.Runtime, _a1 *evm.RegisterLogTrackingRequest) {
	_m.Called(_a0, _a1)
}

// EVMClient_RegisterLogTracking_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RegisterLogTracking'
type EVMClient_RegisterLogTracking_Call struct {
	*mock.Call
}

// RegisterLogTracking is a helper method to define mock.On call
//   - _a0 sdk.Runtime
//   - _a1 *evm.RegisterLogTrackingRequest
func (_e *EVMClient_Expecter) RegisterLogTracking(_a0 interface{}, _a1 interface{}) *EVMClient_RegisterLogTracking_Call {
	return &EVMClient_RegisterLogTracking_Call{Call: _e.mock.On("RegisterLogTracking", _a0, _a1)}
}

func (_c *EVMClient_RegisterLogTracking_Call) Run(run func(_a0 sdk.Runtime, _a1 *evm.RegisterLogTrackingRequest)) *EVMClient_RegisterLogTracking_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(sdk.Runtime), args[1].(*evm.RegisterLogTrackingRequest))
	})
	return _c
}

func (_c *EVMClient_RegisterLogTracking_Call) Return() *EVMClient_RegisterLogTracking_Call {
	_c.Call.Return()
	return _c
}

func (_c *EVMClient_RegisterLogTracking_Call) RunAndReturn(run func(sdk.Runtime, *evm.RegisterLogTrackingRequest)) *EVMClient_RegisterLogTracking_Call {
	_c.Run(run)
	return _c
}

// UnregisterLogTracking provides a mock function with given fields: _a0, _a1
func (_m *EVMClient) UnregisterLogTracking(_a0 sdk.Runtime, _a1 *evm.UnregisterLogTrackingRequest) {
	_m.Called(_a0, _a1)
}

// EVMClient_UnregisterLogTracking_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UnregisterLogTracking'
type EVMClient_UnregisterLogTracking_Call struct {
	*mock.Call
}

// UnregisterLogTracking is a helper method to define mock.On call
//   - _a0 sdk.Runtime
//   - _a1 *evm.UnregisterLogTrackingRequest
func (_e *EVMClient_Expecter) UnregisterLogTracking(_a0 interface{}, _a1 interface{}) *EVMClient_UnregisterLogTracking_Call {
	return &EVMClient_UnregisterLogTracking_Call{Call: _e.mock.On("UnregisterLogTracking", _a0, _a1)}
}

func (_c *EVMClient_UnregisterLogTracking_Call) Run(run func(_a0 sdk.Runtime, _a1 *evm.UnregisterLogTrackingRequest)) *EVMClient_UnregisterLogTracking_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(sdk.Runtime), args[1].(*evm.UnregisterLogTrackingRequest))
	})
	return _c
}

func (_c *EVMClient_UnregisterLogTracking_Call) Return() *EVMClient_UnregisterLogTracking_Call {
	_c.Call.Return()
	return _c
}

func (_c *EVMClient_UnregisterLogTracking_Call) RunAndReturn(run func(sdk.Runtime, *evm.UnregisterLogTrackingRequest)) *EVMClient_UnregisterLogTracking_Call {
	_c.Run(run)
	return _c
}

// NewEVMClient creates a new instance of EVMClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEVMClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *EVMClient {
	mock := &EVMClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
