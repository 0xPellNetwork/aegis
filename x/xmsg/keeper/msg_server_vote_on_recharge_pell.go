package keeper

import (
	"context"
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// VoteOnPellRecharge defines the rpc handler method for MsgVoteOnPellRecharge.
func (k msgServer) VoteOnPellRecharge(goCtx context.Context, msg *types.MsgVoteOnPellRecharge) (*types.MsgVoteOnPellRechargeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	chain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, msg.ChainId)
	if chain == nil {
		return nil, cosmoserrors.Wrap(types.ErrUnsupportedChain, fmt.Sprintf("ChainID : %d ", msg.ChainId))
	}

	if ok := k.relayerKeeper.IsNonTombstonedObserver(ctx, msg.Signer); !ok {
		return nil, relayertypes.ErrNotObserver
	}

	ballotFinalized, err := k.processPellRechargeBallot(ctx, msg)
	if err != nil || !ballotFinalized {
		k.Logger(ctx).Info(fmt.Sprintf("VoteOnPellRecharge: failed to process pell recharge ballot: %s", err))
		return &types.MsgVoteOnPellRechargeResponse{}, err
	}

	k.Logger(ctx).Info(fmt.Sprintf("VoteOnPellRecharge: processed pell recharge ballot, voteIndex: %d", msg.VoteIndex))

	if err := k.processPellRecharge(ctx, msg); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("VoteOnPellRecharge: failed to process pell recharge: %sm voteIndex: %d", err.Error(), msg.VoteIndex))
		return nil, err
	}

	return &types.MsgVoteOnPellRechargeResponse{}, nil
}

// processPellRechargeBallot processes the vote on adding pell token.
func (k msgServer) processPellRechargeBallot(ctx sdk.Context, msg *types.MsgVoteOnPellRecharge) (bool, error) {
	tmpCtx, commit := ctx.CacheContext()
	finalized, isNew, err := k.relayerKeeper.VoteOnAddPellTokenBallot(
		tmpCtx,
		msg.ChainId,
		msg.Signer,
		msg.VoteIndex,
	)
	if err != nil {
		return false, err
	}

	if isNew && k.IsPellRechargeOperationIndexFinalized(tmpCtx, msg.ChainId, msg.VoteIndex) {
		return false, cosmoserrors.Wrap(
			types.ErrObservedTxAlreadyFinalized,
			fmt.Sprintf("senderChainId:%d, voteIndex:%d", msg.ChainId, msg.VoteIndex),
		)
	}
	commit()
	return finalized, nil
}

// Process incoming inbound event serially. If the previous event is not confirmed, an error will be returned.
func (k msgServer) processPellRecharge(ctx sdk.Context, msg *types.MsgVoteOnPellRecharge) error {
	// check if the vote is already finalized
	finalized := k.IsPellRechargeOperationIndexFinalized(ctx, msg.ChainId, msg.VoteIndex)
	if finalized {
		return cosmoserrors.Wrap(
			types.ErrObservedTxAlreadyFinalized,
			fmt.Sprintf("senderChainId:%d, voteIndex:%d", msg.ChainId, msg.VoteIndex),
		)
	}

	// get the chain params
	params, found := k.relayerKeeper.GetChainParamsByChainID(ctx, msg.ChainId)
	if !found {
		return types.ErrUnsupportedChain
	}

	// call bridge pell method on contract
	evmTxResponse, contractCall, err := k.pevmKeeper.CallBridgePellOnPellGateway(
		ctx,
		msg.ChainId,
		ethcommon.HexToAddress(params.GasSwapContractAddress),
		params.PellTokenRechargeAmount.BigInt(),
	)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("VoteOnPellRecharge call pevm success: %s, voteIndex: %d", evmTxResponse.String(), msg.VoteIndex))

	// non-empty msg.Message means this is a contract call; therefore the logs should be processed.
	if !evmTxResponse.Failed() && contractCall {
		logs := evmtypes.LogsToEthereum(evmTxResponse.Logs)
		if len(logs) > 0 {
			gatewayAddr, err := k.pevmKeeper.GetPellGatewayEVMContractAddress(ctx)
			if err != nil {
				return err
			}

			internalMsg := buildEVMMsgByInternalLogs(evmTxResponse.Hash, gatewayAddr)
			if err := k.Hooks().PostTxProcessing(ctx, internalMsg, &ethtypes.Receipt{Logs: logs}); err != nil {
				return err
			}
		}
	}

	// save the vote data
	k.SetPellRechargeOperationIndex(ctx, msg.ChainId, msg.VoteIndex)
	return nil
}
