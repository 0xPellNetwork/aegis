// Code generated by mockery v2.48.0. DO NOT EDIT.

package mocks

import (
	context "context"

	common "github.com/ethereum/go-ethereum/common"

	cosmos_sdktypes "github.com/cosmos/cosmos-sdk/types"

	mock "github.com/stretchr/testify/mock"

	restakingtypes "github.com/0xPellNetwork/aegis/x/restaking/types"

	types "github.com/evmos/ethermint/x/evm/types"

	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// XSecurityPevmKeeper is an autogenerated mock type for the XSecurityPevmKeeper type
type XSecurityPevmKeeper struct {
	mock.Mock
}

// CallRegistryRouterFactory provides a mock function with given fields: ctx, dvsChainApprover, churnApprover, ejector, pauser, unpauser, initialPausedStatus
func (_m *XSecurityPevmKeeper) CallRegistryRouterFactory(ctx context.Context, dvsChainApprover common.Address, churnApprover common.Address, ejector common.Address, pauser common.Address, unpauser common.Address, initialPausedStatus uint) (*types.MsgEthereumTxResponse, bool, error) {
	ret := _m.Called(ctx, dvsChainApprover, churnApprover, ejector, pauser, unpauser, initialPausedStatus)

	if len(ret) == 0 {
		panic("no return value specified for CallRegistryRouterFactory")
	}

	var r0 *types.MsgEthereumTxResponse
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, common.Address, common.Address, common.Address, common.Address, uint) (*types.MsgEthereumTxResponse, bool, error)); ok {
		return rf(ctx, dvsChainApprover, churnApprover, ejector, pauser, unpauser, initialPausedStatus)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, common.Address, common.Address, common.Address, common.Address, uint) *types.MsgEthereumTxResponse); ok {
		r0 = rf(ctx, dvsChainApprover, churnApprover, ejector, pauser, unpauser, initialPausedStatus)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgEthereumTxResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address, common.Address, common.Address, common.Address, common.Address, uint) bool); ok {
		r1 = rf(ctx, dvsChainApprover, churnApprover, ejector, pauser, unpauser, initialPausedStatus)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, common.Address, common.Address, common.Address, common.Address, common.Address, uint) error); ok {
		r2 = rf(ctx, dvsChainApprover, churnApprover, ejector, pauser, unpauser, initialPausedStatus)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CallRegistryRouterToCreateGroup provides a mock function with given fields: ctx, registryRouterAddress, operatorSetParams, minimumStake, poolParams, groupEjectionParams
func (_m *XSecurityPevmKeeper) CallRegistryRouterToCreateGroup(ctx cosmos_sdktypes.Context, registryRouterAddress common.Address, operatorSetParams restakingtypes.OperatorSetParam, minimumStake int64, poolParams []restakingtypes.PoolParams, groupEjectionParams restakingtypes.GroupEjectionParam) (*types.MsgEthereumTxResponse, bool, error) {
	ret := _m.Called(ctx, registryRouterAddress, operatorSetParams, minimumStake, poolParams, groupEjectionParams)

	if len(ret) == 0 {
		panic("no return value specified for CallRegistryRouterToCreateGroup")
	}

	var r0 *types.MsgEthereumTxResponse
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, restakingtypes.OperatorSetParam, int64, []restakingtypes.PoolParams, restakingtypes.GroupEjectionParam) (*types.MsgEthereumTxResponse, bool, error)); ok {
		return rf(ctx, registryRouterAddress, operatorSetParams, minimumStake, poolParams, groupEjectionParams)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, restakingtypes.OperatorSetParam, int64, []restakingtypes.PoolParams, restakingtypes.GroupEjectionParam) *types.MsgEthereumTxResponse); ok {
		r0 = rf(ctx, registryRouterAddress, operatorSetParams, minimumStake, poolParams, groupEjectionParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgEthereumTxResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context, common.Address, restakingtypes.OperatorSetParam, int64, []restakingtypes.PoolParams, restakingtypes.GroupEjectionParam) bool); ok {
		r1 = rf(ctx, registryRouterAddress, operatorSetParams, minimumStake, poolParams, groupEjectionParams)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(cosmos_sdktypes.Context, common.Address, restakingtypes.OperatorSetParam, int64, []restakingtypes.PoolParams, restakingtypes.GroupEjectionParam) error); ok {
		r2 = rf(ctx, registryRouterAddress, operatorSetParams, minimumStake, poolParams, groupEjectionParams)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CallRegistryRouterToRegisterOperator provides a mock function with given fields: ctx, registryRouterAddress, param, operatorAddress, groupNumbers
func (_m *XSecurityPevmKeeper) CallRegistryRouterToRegisterOperator(ctx cosmos_sdktypes.Context, registryRouterAddress common.Address, param xsecuritytypes.RegisterOperatorParam, operatorAddress common.Address, groupNumbers uint64) (*types.MsgEthereumTxResponse, bool, error) {
	ret := _m.Called(ctx, registryRouterAddress, param, operatorAddress, groupNumbers)

	if len(ret) == 0 {
		panic("no return value specified for CallRegistryRouterToRegisterOperator")
	}

	var r0 *types.MsgEthereumTxResponse
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, xsecuritytypes.RegisterOperatorParam, common.Address, uint64) (*types.MsgEthereumTxResponse, bool, error)); ok {
		return rf(ctx, registryRouterAddress, param, operatorAddress, groupNumbers)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, xsecuritytypes.RegisterOperatorParam, common.Address, uint64) *types.MsgEthereumTxResponse); ok {
		r0 = rf(ctx, registryRouterAddress, param, operatorAddress, groupNumbers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgEthereumTxResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context, common.Address, xsecuritytypes.RegisterOperatorParam, common.Address, uint64) bool); ok {
		r1 = rf(ctx, registryRouterAddress, param, operatorAddress, groupNumbers)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(cosmos_sdktypes.Context, common.Address, xsecuritytypes.RegisterOperatorParam, common.Address, uint64) error); ok {
		r2 = rf(ctx, registryRouterAddress, param, operatorAddress, groupNumbers)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CallRegistryRouterToSetOperatorSetParams provides a mock function with given fields: ctx, stakeRegistryRouterAddress, groupNumbers, operatorSetParams
func (_m *XSecurityPevmKeeper) CallRegistryRouterToSetOperatorSetParams(ctx cosmos_sdktypes.Context, stakeRegistryRouterAddress common.Address, groupNumbers uint64, operatorSetParams *restakingtypes.OperatorSetParam) (*types.MsgEthereumTxResponse, bool, error) {
	ret := _m.Called(ctx, stakeRegistryRouterAddress, groupNumbers, operatorSetParams)

	if len(ret) == 0 {
		panic("no return value specified for CallRegistryRouterToSetOperatorSetParams")
	}

	var r0 *types.MsgEthereumTxResponse
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, uint64, *restakingtypes.OperatorSetParam) (*types.MsgEthereumTxResponse, bool, error)); ok {
		return rf(ctx, stakeRegistryRouterAddress, groupNumbers, operatorSetParams)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, uint64, *restakingtypes.OperatorSetParam) *types.MsgEthereumTxResponse); ok {
		r0 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, operatorSetParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgEthereumTxResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context, common.Address, uint64, *restakingtypes.OperatorSetParam) bool); ok {
		r1 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, operatorSetParams)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(cosmos_sdktypes.Context, common.Address, uint64, *restakingtypes.OperatorSetParam) error); ok {
		r2 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, operatorSetParams)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CallStakelRegistryRouterToAddPools provides a mock function with given fields: ctx, stakeRegistryRouterAddress, groupNumbers, poolParams
func (_m *XSecurityPevmKeeper) CallStakelRegistryRouterToAddPools(ctx cosmos_sdktypes.Context, stakeRegistryRouterAddress common.Address, groupNumbers uint64, poolParams []*restakingtypes.PoolParams) (*types.MsgEthereumTxResponse, bool, error) {
	ret := _m.Called(ctx, stakeRegistryRouterAddress, groupNumbers, poolParams)

	if len(ret) == 0 {
		panic("no return value specified for CallStakelRegistryRouterToAddPools")
	}

	var r0 *types.MsgEthereumTxResponse
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, uint64, []*restakingtypes.PoolParams) (*types.MsgEthereumTxResponse, bool, error)); ok {
		return rf(ctx, stakeRegistryRouterAddress, groupNumbers, poolParams)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, uint64, []*restakingtypes.PoolParams) *types.MsgEthereumTxResponse); ok {
		r0 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, poolParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgEthereumTxResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context, common.Address, uint64, []*restakingtypes.PoolParams) bool); ok {
		r1 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, poolParams)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(cosmos_sdktypes.Context, common.Address, uint64, []*restakingtypes.PoolParams) error); ok {
		r2 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, poolParams)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CallStakelRegistryRouterToRemovePools provides a mock function with given fields: ctx, stakeRegistryRouterAddress, groupNumbers, indicesToRemove
func (_m *XSecurityPevmKeeper) CallStakelRegistryRouterToRemovePools(ctx cosmos_sdktypes.Context, stakeRegistryRouterAddress common.Address, groupNumbers uint64, indicesToRemove []uint) (*types.MsgEthereumTxResponse, bool, error) {
	ret := _m.Called(ctx, stakeRegistryRouterAddress, groupNumbers, indicesToRemove)

	if len(ret) == 0 {
		panic("no return value specified for CallStakelRegistryRouterToRemovePools")
	}

	var r0 *types.MsgEthereumTxResponse
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, uint64, []uint) (*types.MsgEthereumTxResponse, bool, error)); ok {
		return rf(ctx, stakeRegistryRouterAddress, groupNumbers, indicesToRemove)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, common.Address, uint64, []uint) *types.MsgEthereumTxResponse); ok {
		r0 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, indicesToRemove)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.MsgEthereumTxResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context, common.Address, uint64, []uint) bool); ok {
		r1 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, indicesToRemove)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(cosmos_sdktypes.Context, common.Address, uint64, []uint) error); ok {
		r2 = rf(ctx, stakeRegistryRouterAddress, groupNumbers, indicesToRemove)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CallUpdateDestinationAddressOnGasSwapPEVM provides a mock function with given fields: ctx, chainId, destinationAddress
func (_m *XSecurityPevmKeeper) CallUpdateDestinationAddressOnGasSwapPEVM(ctx context.Context, chainId int64, destinationAddress string) error {
	ret := _m.Called(ctx, chainId, destinationAddress)

	if len(ret) == 0 {
		panic("no return value specified for CallUpdateDestinationAddressOnGasSwapPEVM")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) error); ok {
		r0 = rf(ctx, chainId, destinationAddress)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CallUpdateDestinationAddressOnPellGateway provides a mock function with given fields: ctx, chainId, destinationAddress
func (_m *XSecurityPevmKeeper) CallUpdateDestinationAddressOnPellGateway(ctx context.Context, chainId int64, destinationAddress string) error {
	ret := _m.Called(ctx, chainId, destinationAddress)

	if len(ret) == 0 {
		panic("no return value specified for CallUpdateDestinationAddressOnPellGateway")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) error); ok {
		r0 = rf(ctx, chainId, destinationAddress)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CallUpdateSourceAddressOnPellGateway provides a mock function with given fields: ctx, chainId, sourceAddress
func (_m *XSecurityPevmKeeper) CallUpdateSourceAddressOnPellGateway(ctx context.Context, chainId int64, sourceAddress string) error {
	ret := _m.Called(ctx, chainId, sourceAddress)

	if len(ret) == 0 {
		panic("no return value specified for CallUpdateSourceAddressOnPellGateway")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) error); ok {
		r0 = rf(ctx, chainId, sourceAddress)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewXSecurityPevmKeeper creates a new instance of XSecurityPevmKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewXSecurityPevmKeeper(t interface {
	mock.TestingT
	Cleanup(func())
}) *XSecurityPevmKeeper {
	mock := &XSecurityPevmKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
