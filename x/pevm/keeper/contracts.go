package keeper

import (
	"context"
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	eth "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
	xsecuritytypes "github.com/pell-chain/pellcore/x/xsecurity/types"
)

// CallSyncDepositStateOnPellStrategyManager calls the contract
// returns [txResponse, isContractCall, error]
// isContractCall is true if the receiver is a contract and a contract call was made
func (k Keeper) CallSyncDepositStateOnPellStrategyManager(
	ctx context.Context,
	from []byte,
	senderChainID int64,
	staker, strategy eth.Address,
	shares *big.Int,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callSyncDepositStateOnPellStrategyManager(sdk.UnwrapSDKContext(ctx), big.NewInt(senderChainID), staker, strategy, shares)
	return res, true, err
}

// CallSyncDelegatedStateOnPellDelegationManager calls the contract
// returns [txResponse, isContractCall, error]
// isContractCall is true if the receiver is a contract and a contract call was made
func (k Keeper) CallSyncDelegatedStateOnPellDelegationManager(
	ctx context.Context,
	from []byte,
	senderChainID int64,
	staker, operator eth.Address,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callSyncDelegatedStateOnPellDelegationManager(sdk.UnwrapSDKContext(ctx), big.NewInt(senderChainID), staker, operator)
	return res, true, err
}

// CallSyncWithdrawalStateOnPellDelegationManager calls the contract
// returns [txResponse, isContractCall, error]
// isContractCall is true if the receiver is a contract and a contract call was made
func (k Keeper) CallSyncWithdrawalStateOnPellDelegationManager(
	ctx context.Context,
	senderChainID int64,
	staker eth.Address,
	withdrawalParam *types.WithdrawalQueued,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callSyncWithdrawalStateOnPellDelegationManager(sdk.UnwrapSDKContext(ctx), big.NewInt(senderChainID), staker, withdrawalParam)
	return res, true, err
}

// CallSyncUndelegateStateOnPellDelegationManager calls the contract
// returns [txResponse, isContractCall, error]
// isContractCall is true if the receiver is a contract and a contract call was made
func (k Keeper) CallSyncUndelegateStateOnPellDelegationManager(
	ctx context.Context,
	senderChainID int64,
	staker eth.Address,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callSyncUndelegatedStateOnPellDelegationManager(sdk.UnwrapSDKContext(ctx), big.NewInt(senderChainID), staker)
	return res, true, err
}

// CallBridgePellOnPellGatewayPEVM calls the contract
func (k Keeper) CallBridgePellOnPellGateway(
	ctx context.Context,
	destinationChainId int64,
	receiver eth.Address,
	amount *big.Int,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callBridgePellOnPellGateway(sdk.UnwrapSDKContext(ctx), big.NewInt(destinationChainId), receiver, amount)
	return res, true, err
}

// CallSwapOnPellGasSwapEVM calls the contract
func (k Keeper) CallSwapOnPellGasSwap(
	ctx context.Context,
	destinationChainId int64,
	amountIn *big.Int,
	receiver eth.Address,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callSwapOnPellGasSwap(sdk.UnwrapSDKContext(ctx), big.NewInt(destinationChainId), amountIn, receiver)
	return res, true, err
}

// CallUpdateDestinationAddressOnPellGatewayPEVM calls the contract
func (k Keeper) CallUpdateDestinationAddressOnPellGateway(
	ctx context.Context,
	chainId int64,
	destinationAddress string,
) error {
	destinationAddr := eth.HexToAddress(destinationAddress)
	res, err := k.callUpdateDestinationAddressOnPellGateway(sdk.UnwrapSDKContext(ctx), big.NewInt(chainId), destinationAddr)
	if err == nil && res != nil && res.Failed() {
		err = errors.New(res.VmError)
	}
	return err
}

// CallUpdateSourceAddressOnPellGateway calls the contract
func (k Keeper) CallUpdateSourceAddressOnPellGateway(
	ctx context.Context,
	chainId int64,
	destinationAddress string,
) error {
	destinationAddr := eth.HexToAddress(destinationAddress)
	res, err := k.callUpdateSourceAddressOnPellGateway(sdk.UnwrapSDKContext(ctx), big.NewInt(chainId), destinationAddr)
	if err == nil && res != nil && res.Failed() {
		err = errors.New(res.VmError)
	}
	return err
}

// CallUpdateDestinationAddressOnGasSwapPEVM calls the contract
func (k Keeper) CallUpdateDestinationAddressOnGasSwapPEVM(
	ctx context.Context,
	chainId int64,
	destinationAddress string,
) error {
	destinationAddr := eth.HexToAddress(destinationAddress)
	res, err := k.callUpdateDestinationAddressOnGasSwapPEVM(sdk.UnwrapSDKContext(ctx), big.NewInt(chainId), destinationAddr)
	if err == nil && res != nil && res.Failed() {
		err = errors.New(res.VmError)
	}
	return err
}

func (k Keeper) CallAddSupportedChainOnRegistryRouter(
	ctx sdk.Context,
	params *types.RegisterChainDVSToPell,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callAddSupportedChainOnRegistryRouter(sdk.UnwrapSDKContext(ctx), params)
	if err == nil && res != nil && res.Failed() {
		err = errors.New(res.VmError)
	}

	return res, true, err
}

func (k Keeper) CallProcessPellSent(ctx sdk.Context, action *types.PellSent, xmsgIndex string) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	res, err := k.callProcessPellSent(sdk.UnwrapSDKContext(ctx), action, xmsgIndex)
	if err == nil && res != nil && res.Failed() {
		err = errors.New(res.VmError)
	}

	return res, true, err
}

// ------------- LST Token staking -------------

// CallRegistryRouterFactory call the RegistryRouterFactory contract to create a new RegistryRouter
func (k Keeper) CallRegistryRouterFactory(
	ctx context.Context,
	dvsChainApprover, churnApprover, ejector, pauser, unpauser eth.Address,
	initialPausedStatus uint,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	res, err := k.callRegistryRouterFactory(sdkCtx, dvsChainApprover, churnApprover, ejector, pauser, unpauser, initialPausedStatus)
	return res, true, err
}

// CallRegistryRouterToCreateGroup call the RegistryRouter contract to create a new group
func (k Keeper) CallRegistryRouterToCreateGroup(
	ctx sdk.Context,
	registryRouterAddress eth.Address,
	operatorSetParams restakingtypes.OperatorSetParam,
	minimumStake int64,
	poolParams []restakingtypes.PoolParams,
	groupEjectionParams restakingtypes.GroupEjectionParam,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	res, err := k.callRegistryRouterToCreateGroup(sdkCtx, registryRouterAddress, operatorSetParams, minimumStake, poolParams, groupEjectionParams)
	return res, true, err
}

// CallRegistryRouterToEjectGroup call the RegistryRouter contract to eject a group
func (k Keeper) CallRegistryRouterToRegisterOperator(
	ctx sdk.Context,
	registryRouterAddress eth.Address,
	param xsecuritytypes.RegisterOperatorParam,
	operatorAddress eth.Address,
	groupNumbers uint64,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	res, err := k.callRegistryRouterToRegisterOperator(sdkCtx, registryRouterAddress, param, operatorAddress, groupNumbers)
	return res, true, err
}

// CallRegistryRouterToEjectGroup call the RegistryRouter contract to eject a group
func (k Keeper) CallStakelRegistryRouterToAddPools(
	ctx sdk.Context,
	stakeRegistryRouterAddress eth.Address,
	groupNumbers uint64,
	poolParams []*restakingtypes.PoolParams,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	res, err := k.callStakeRegistryRouterToAddPools(sdkCtx, stakeRegistryRouterAddress, groupNumbers, poolParams)
	return res, true, err
}

// CallRegistryRouterToEjectGroup call the RegistryRouter contract to eject a group
func (k Keeper) CallStakelRegistryRouterToRemovePools(
	ctx sdk.Context,
	stakeRegistryRouterAddress eth.Address,
	groupNumbers uint64,
	indicesToRemove []uint,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	res, err := k.callStakeRegistryRouterToRemovePools(sdkCtx, stakeRegistryRouterAddress, groupNumbers, indicesToRemove)
	return res, true, err
}

// CallRegistryRouterToEjectGroup call the RegistryRouter contract to eject a group
func (k Keeper) CallRegistryRouterToSetOperatorSetParams(
	ctx sdk.Context,
	registryRouterAddress eth.Address,
	groupNumbers uint64,
	operatorSetParams *restakingtypes.OperatorSetParam,
) (*evmtypes.MsgEthereumTxResponse, bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	res, err := k.callRegistryRouterToSetOperatorSetParams(sdkCtx, registryRouterAddress, groupNumbers, operatorSetParams)
	return res, true, err
}
