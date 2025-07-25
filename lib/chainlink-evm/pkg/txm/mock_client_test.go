// Code generated by mockery v2.53.3. DO NOT EDIT.

package txm

import (
	context "context"
	big "math/big"

	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	types "github.com/smartcontractkit/chainlink-evm/pkg/txm/types"
)

// mockClient is an autogenerated mock type for the Client type
type mockClient struct {
	mock.Mock
}

type mockClient_Expecter struct {
	mock *mock.Mock
}

func (_m *mockClient) EXPECT() *mockClient_Expecter {
	return &mockClient_Expecter{mock: &_m.Mock}
}

// NonceAt provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockClient) NonceAt(_a0 context.Context, _a1 common.Address, _a2 *big.Int) (uint64, error) {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for NonceAt")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *big.Int) (uint64, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *big.Int) uint64); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address, *big.Int) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockClient_NonceAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NonceAt'
type mockClient_NonceAt_Call struct {
	*mock.Call
}

// NonceAt is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 common.Address
//   - _a2 *big.Int
func (_e *mockClient_Expecter) NonceAt(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockClient_NonceAt_Call {
	return &mockClient_NonceAt_Call{Call: _e.mock.On("NonceAt", _a0, _a1, _a2)}
}

func (_c *mockClient_NonceAt_Call) Run(run func(_a0 context.Context, _a1 common.Address, _a2 *big.Int)) *mockClient_NonceAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address), args[2].(*big.Int))
	})
	return _c
}

func (_c *mockClient_NonceAt_Call) Return(_a0 uint64, _a1 error) *mockClient_NonceAt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockClient_NonceAt_Call) RunAndReturn(run func(context.Context, common.Address, *big.Int) (uint64, error)) *mockClient_NonceAt_Call {
	_c.Call.Return(run)
	return _c
}

// PendingNonceAt provides a mock function with given fields: _a0, _a1
func (_m *mockClient) PendingNonceAt(_a0 context.Context, _a1 common.Address) (uint64, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for PendingNonceAt")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) (uint64, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) uint64); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockClient_PendingNonceAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PendingNonceAt'
type mockClient_PendingNonceAt_Call struct {
	*mock.Call
}

// PendingNonceAt is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 common.Address
func (_e *mockClient_Expecter) PendingNonceAt(_a0 interface{}, _a1 interface{}) *mockClient_PendingNonceAt_Call {
	return &mockClient_PendingNonceAt_Call{Call: _e.mock.On("PendingNonceAt", _a0, _a1)}
}

func (_c *mockClient_PendingNonceAt_Call) Run(run func(_a0 context.Context, _a1 common.Address)) *mockClient_PendingNonceAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address))
	})
	return _c
}

func (_c *mockClient_PendingNonceAt_Call) Return(_a0 uint64, _a1 error) *mockClient_PendingNonceAt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockClient_PendingNonceAt_Call) RunAndReturn(run func(context.Context, common.Address) (uint64, error)) *mockClient_PendingNonceAt_Call {
	_c.Call.Return(run)
	return _c
}

// SendTransaction provides a mock function with given fields: ctx, tx, attempt
func (_m *mockClient) SendTransaction(ctx context.Context, tx *types.Transaction, attempt *types.Attempt) error {
	ret := _m.Called(ctx, tx, attempt)

	if len(ret) == 0 {
		panic("no return value specified for SendTransaction")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Transaction, *types.Attempt) error); ok {
		r0 = rf(ctx, tx, attempt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockClient_SendTransaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendTransaction'
type mockClient_SendTransaction_Call struct {
	*mock.Call
}

// SendTransaction is a helper method to define mock.On call
//   - ctx context.Context
//   - tx *types.Transaction
//   - attempt *types.Attempt
func (_e *mockClient_Expecter) SendTransaction(ctx interface{}, tx interface{}, attempt interface{}) *mockClient_SendTransaction_Call {
	return &mockClient_SendTransaction_Call{Call: _e.mock.On("SendTransaction", ctx, tx, attempt)}
}

func (_c *mockClient_SendTransaction_Call) Run(run func(ctx context.Context, tx *types.Transaction, attempt *types.Attempt)) *mockClient_SendTransaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*types.Transaction), args[2].(*types.Attempt))
	})
	return _c
}

func (_c *mockClient_SendTransaction_Call) Return(_a0 error) *mockClient_SendTransaction_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockClient_SendTransaction_Call) RunAndReturn(run func(context.Context, *types.Transaction, *types.Attempt) error) *mockClient_SendTransaction_Call {
	_c.Call.Return(run)
	return _c
}

// newMockClient creates a new instance of mockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockClient {
	mock := &mockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
