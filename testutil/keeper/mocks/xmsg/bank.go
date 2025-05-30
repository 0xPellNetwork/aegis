// Code generated by mockery v2.48.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/cosmos/cosmos-sdk/types"
)

// XmsgBankKeeper is an autogenerated mock type for the XmsgBankKeeper type
type XmsgBankKeeper struct {
	mock.Mock
}

// BurnCoins provides a mock function with given fields: ctx, name, amt
func (_m *XmsgBankKeeper) BurnCoins(ctx context.Context, name string, amt types.Coins) error {
	ret := _m.Called(ctx, name, amt)

	if len(ret) == 0 {
		panic("no return value specified for BurnCoins")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, types.Coins) error); ok {
		r0 = rf(ctx, name, amt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MintCoins provides a mock function with given fields: ctx, moduleName, amt
func (_m *XmsgBankKeeper) MintCoins(ctx context.Context, moduleName string, amt types.Coins) error {
	ret := _m.Called(ctx, moduleName, amt)

	if len(ret) == 0 {
		panic("no return value specified for MintCoins")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, types.Coins) error); ok {
		r0 = rf(ctx, moduleName, amt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendCoinsFromAccountToModule provides a mock function with given fields: ctx, senderAddr, recipientModule, amt
func (_m *XmsgBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr types.AccAddress, recipientModule string, amt types.Coins) error {
	ret := _m.Called(ctx, senderAddr, recipientModule, amt)

	if len(ret) == 0 {
		panic("no return value specified for SendCoinsFromAccountToModule")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress, string, types.Coins) error); ok {
		r0 = rf(ctx, senderAddr, recipientModule, amt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewXmsgBankKeeper creates a new instance of XmsgBankKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewXmsgBankKeeper(t interface {
	mock.TestingT
	Cleanup(func())
}) *XmsgBankKeeper {
	mock := &XmsgBankKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
