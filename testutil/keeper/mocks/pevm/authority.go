// Code generated by mockery v2.48.0. DO NOT EDIT.

package mocks

import (
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	mock "github.com/stretchr/testify/mock"

	types "github.com/cosmos/cosmos-sdk/types"
)

// PevmAuthorityKeeper is an autogenerated mock type for the PevmAuthorityKeeper type
type PevmAuthorityKeeper struct {
	mock.Mock
}

// IsAuthorized provides a mock function with given fields: ctx, address, policyType
func (_m *PevmAuthorityKeeper) IsAuthorized(ctx types.Context, address string, policyType authoritytypes.PolicyType) bool {
	ret := _m.Called(ctx, address, policyType)

	if len(ret) == 0 {
		panic("no return value specified for IsAuthorized")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(types.Context, string, authoritytypes.PolicyType) bool); ok {
		r0 = rf(ctx, address, policyType)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NewPevmAuthorityKeeper creates a new instance of PevmAuthorityKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPevmAuthorityKeeper(t interface {
	mock.TestingT
	Cleanup(func())
}) *PevmAuthorityKeeper {
	mock := &PevmAuthorityKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
