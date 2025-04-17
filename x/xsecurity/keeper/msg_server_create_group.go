package keeper

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// CreateGroup creates a DVS group
func (k Keeper) CreateGroup(goCtx context.Context, msg *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.Logger(ctx).Info(fmt.Sprintf("CreateGroup receive request msg: %v", msg))

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgCreateGroupResponse{}, authoritytypes.ErrUnauthorized
	}

	// check if registry router already exists
	registryRouterAddress, exist := k.GetLSTRegistryRouterAddress(ctx)
	if !exist {
		k.Logger(ctx).Error("CreateGroup registry router not exists")
		return nil, fmt.Errorf("registry router not exists")
	}

	// check if group already exists
	if _, exist := k.GetGroupInfo(ctx); exist {
		k.Logger(ctx).Error("CreateGroup group already exists")
		return nil, fmt.Errorf("group already exists")
	}

	// build params to call registry router to create group
	registryRouterAddressParam := common.HexToAddress(registryRouterAddress.RegistryRouterAddress)
	operatorSetParams := restakingtypes.OperatorSetParam{
		MaxOperatorCount:        msg.OperatorSetParams.MaxOperatorCount,
		KickBipsOfOperatorStake: msg.OperatorSetParams.KickBipsOfOperatorStake,
		KickBipsOfTotalStake:    msg.OperatorSetParams.KickBipsOfTotalStake,
	}

	// check if minStake is valid
	minStakeUint64 := msg.MinStake.Uint64()
	if minStakeUint64 > math.MaxInt64 {
		return nil, fmt.Errorf("MinStake value exceeds int64 range")
	}
	minStake := int64(minStakeUint64)

	var poolParams []restakingtypes.PoolParams
	for _, v := range msg.PoolParams {
		poolParams = append(poolParams, restakingtypes.PoolParams{
			ChainId:    v.ChainId,
			Pool:       v.Pool,
			Multiplier: v.Multiplier,
		})
	}

	k.Logger(ctx).Info(fmt.Sprintf("CreateGroup call pevm module params: %v, %v, %v, %v, %v", registryRouterAddressParam, operatorSetParams, minStake, poolParams, msg.GroupEjectionParams))

	receipt, _, err := k.pevmKeeper.CallRegistryRouterToCreateGroup(ctx, registryRouterAddressParam, operatorSetParams, minStake, poolParams, *msg.GroupEjectionParams)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("CreateGroup call pevm module err: %v", err))
		return nil, err
	}
	k.Logger(ctx).Info(fmt.Sprintf("CreateGroup call pevm module success: %v", receipt))

	// get group number from receipt
	groupNumber, err := k.GetGroupNumberFromReceipt(ctx, receipt)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("CreateGroup filter group number err: %v", err))
		return nil, err
	}

	// the group number should be 0
	if groupNumber != 0 {
		k.Logger(ctx).Error(fmt.Sprintf("CreateGroup get wrong group number: %v", groupNumber))
		return nil, fmt.Errorf("CreateGroup get wrong group number: %v", groupNumber)
	}

	// save group info to store
	k.SetGroupInfo(ctx, &types.LSTGroupInfo{
		GroupNumber:        groupNumber,
		OperatorSetParam:   msg.OperatorSetParams,
		MinimumStake:       msg.MinStake,
		PoolParams:         msg.PoolParams,
		GroupEjectionParam: msg.GroupEjectionParams,
	})

	k.Logger(ctx).Info(fmt.Sprintf("SetGroupInfo success: %v", receipt))
	return &types.MsgCreateGroupResponse{}, nil
}

// GetGroupNumberFromReceipt filters the group number from the receipt
func (k Keeper) GetGroupNumberFromReceipt(ctx sdk.Context, receipt *evmtypes.MsgEthereumTxResponse) (uint64, error) {
	for _, log := range receipt.Logs {
		if log.Topics[0] == registryRouterMetaDataABI.Events["SyncCreateGroup"].ID.String() {
			groupNumberHex := strings.TrimPrefix(log.Topics[1], "0x")
			groupNumber, err := strconv.ParseUint(groupNumberHex, 16, 64)
			if err != nil {
				return 0, err
			}

			return groupNumber, nil
		}
	}

	return 0, fmt.Errorf("createGroup event not found in receipt")
}
