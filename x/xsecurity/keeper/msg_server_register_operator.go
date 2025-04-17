package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// RegisterOperator registers an operator
// This transaction will first register the operator to DVS, and then bind the operator address to the validator address. The specific process is as follows:
// 1. Call the registry router contract of the pevm contract to register the operator as DVS. During registration, the contract has already verified the BLS signature and saved the operator -> BLS correspondence in the contract. If this step is successful, continue with the subsequent steps.
// 2. From the staking module, query the validator address information based on the Cosmos signer address.
// 3. From the registry router, query the BLS public key, operator ID, and other information based on the operator's address.
// 4. The transaction binds the operator's BLS public key to the validator_address.
// This transaction can only be initiated by the validator's account itself:
func (k Keeper) RegisterOperator(goCtx context.Context, msg *types.MsgRegisterOperator) (*types.MsgRegisterOperatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.Logger(ctx).Info(fmt.Sprintf("RegisterOperator receive request msg: %v", msg))

	// check if registry router already exists
	registryRouterAddress, exist := k.GetLSTRegistryRouterAddress(ctx)
	if !exist {
		k.Logger(ctx).Error("RegisterOperator registry router not exists")
		return nil, fmt.Errorf("registry router not exists")
	}

	// check if group already exists
	groupInfo, exist := k.GetGroupInfo(ctx)
	if !exist {
		k.Logger(ctx).Error("RegisterOperator group not exists")
		return nil, fmt.Errorf("group not exists")
	}

	// check if operator registration already exists
	list, exist := k.GetOperatorRegistrationList(ctx)
	if exist {
		for _, registration := range list.OperatorRegistrations {
			if registration.OperatorAddress == msg.OperatorAddress {
				k.Logger(ctx).Error(fmt.Sprintf("RegisterOperator operator address already exists: %v", msg.OperatorAddress))
				return nil, fmt.Errorf("operator address already exists")
			}
		}
	}

	// register operator to DVS
	operatorID, err := k.RegisterOperatorToDVS(ctx, msg, registryRouterAddress, groupInfo)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("RegisterOperator GetOperatorIdFromReceipt err: %v", err))
		return nil, err
	}
	k.Logger(ctx).Info(fmt.Sprintf("RegisterOperator operatorID: %v", operatorID))

	// query validator address info by signer
	validator, err := k.QueryValidatorInfoFromStakingModule(ctx, msg.Signer)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("RegisterOperator QueryValidatorInfoFromStakingModule err: %v", err))
		return nil, err
	}

	// binding the operator address to the validator address
	data := &types.LSTOperatorRegistration{
		OperatorAddress:       msg.OperatorAddress,
		OperatorId:            operatorID,
		RegisterOperatorParam: msg.RegisterOperatorParam,
		ValidatorAddress:      validator.OperatorAddress,
	}

	// save operator registration to store
	k.AddOperatorRegistration(ctx, data)

	return &types.MsgRegisterOperatorResponse{}, nil
}

// RegisterOperatorToDVS registers an operator to DVS
func (k Keeper) RegisterOperatorToDVS(ctx sdk.Context, msg *types.MsgRegisterOperator, registryRouterAddress *types.LSTRegistryRouterAddress, groupInfo *types.LSTGroupInfo) ([]byte, error) {
	// build params to call registry router to register operator
	registryRouterAddressParam := common.HexToAddress(registryRouterAddress.RegistryRouterAddress)
	operatorAddress := common.HexToAddress(msg.OperatorAddress)

	// call registry router to register operator
	receipt, _, err := k.pevmKeeper.CallRegistryRouterToRegisterOperator(ctx, registryRouterAddressParam, *msg.RegisterOperatorParam, operatorAddress, groupInfo.GroupNumber)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("RegisterOperator call pevm module err: %v", err))
		return nil, err
	}
	k.Logger(ctx).Info(fmt.Sprintf("RegisterOperator call pevm module success: %v", receipt))

	return k.GetOperatorIdFromReceipt(ctx, receipt)
}

// QueryValidatorInfoFromStakingModule queries the validator info from the staking module
func (k Keeper) QueryValidatorInfoFromStakingModule(ctx sdk.Context, signer string) (stakingtypes.Validator, error) {
	accAddr, err := sdk.AccAddressFromBech32(signer)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("RegisterOperator AccAddressFromBech32 err: %v", err))
		return stakingtypes.Validator{}, err
	}
	valAddr := sdk.ValAddress(accAddr)
	k.Logger(ctx).Info(fmt.Sprintf("RegisterOperator signer address: %v, accAddr: %v, valAddress: %v", signer, accAddr, valAddr))

	// get validator info
	return k.stakingKeeper.GetValidator(ctx, valAddr)
}

// GetOperatorIdFromReceipt extracts the operatorId from the SyncRegisterOperator event log
// in the transaction receipt.
func (k Keeper) GetOperatorIdFromReceipt(
	ctx sdk.Context,
	receipt *evmtypes.MsgEthereumTxResponse,
) ([]byte, error) {
	// Search for the SyncRegisterOperator event in all transaction logs
	for _, log := range receipt.Logs {
		// Check if this log represents the SyncRegisterOperator event
		// by comparing the first topic (event signature hash) with the known event ID
		if log.Topics[0] == registryRouterMetaDataABI.Events["SyncRegisterOperator"].ID.String() {
			// Validate that we have enough topics in the log
			// We need at least 3: event signature, first param, second param (operatorId)
			if len(log.Topics) < 3 {
				return nil, fmt.Errorf("insufficient topics in log: got %d, need at least 3", len(log.Topics))
			}

			return hex.DecodeString(strings.TrimPrefix(log.Topics[2], "0x"))
		}
	}

	// If we didn't find the event in any of the logs, return an error
	return nil, fmt.Errorf("SyncRegisterOperator event not found in transaction receipt")
}
