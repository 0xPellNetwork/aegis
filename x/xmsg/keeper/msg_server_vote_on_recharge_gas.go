package keeper

import (
	"context"
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// VoteOnGasRecharge defines the rpc handler method for MsgVoteOnGasRecharge.
func (k msgServer) VoteOnGasRecharge(goCtx context.Context, msg *types.MsgVoteOnGasRecharge) (*types.MsgVoteOnGasRechargeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	chain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, msg.ChainId)
	if chain == nil {
		return nil, cosmoserrors.Wrap(types.ErrUnsupportedChain, fmt.Sprintf("ChainID : %d ", msg.ChainId))
	}

	if ok := k.relayerKeeper.IsNonTombstonedObserver(ctx, msg.Signer); !ok {
		return nil, observertypes.ErrNotObserver
	}

	ballotFinalized, err := k.processGasRechargeBallot(ctx, msg)
	if err != nil || !ballotFinalized {
		k.Logger(ctx).Info(fmt.Sprintf("VoteOnGasRecharge: failed to process gas recharge ballot: %s", err))
		return &types.MsgVoteOnGasRechargeResponse{}, err
	}

	k.Logger(ctx).Info(fmt.Sprintf("VoteOnGasRecharge: processed gas recharge ballot, voteIndex: %d", msg.VoteIndex))

	if err := k.processGasRecharge(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgVoteOnGasRechargeResponse{}, nil
}

// processGasRechargeBallot processes the vote on adding gas token.
func (k msgServer) processGasRechargeBallot(ctx sdk.Context, msg *types.MsgVoteOnGasRecharge) (bool, error) {
	tmpCtx, commit := ctx.CacheContext()
	finalized, isNew, err := k.relayerKeeper.VoteOnAddGasTokenBallot(
		tmpCtx,
		msg.ChainId,
		msg.Signer,
		msg.VoteIndex,
	)
	if err != nil {
		return false, err
	}

	if isNew && k.IsGasRechargeOperationIndexFinalized(tmpCtx, msg.ChainId, msg.VoteIndex) {
		return false, cosmoserrors.Wrap(
			types.ErrObservedTxAlreadyFinalized,
			fmt.Sprintf("senderChainId:%d, voteIndex:%d", msg.ChainId, msg.VoteIndex),
		)
	}
	commit()
	return finalized, nil
}

// Process incoming inbound event serially. If the previous event is not confirmed, an error will be returned.
func (k msgServer) processGasRecharge(ctx sdk.Context, msg *types.MsgVoteOnGasRecharge) error {
	// check if the vote is already finalized
	finalized := k.IsGasRechargeOperationIndexFinalized(ctx, msg.ChainId, msg.VoteIndex)
	if finalized {
		return cosmoserrors.Wrap(
			types.ErrObservedTxAlreadyFinalized,
			fmt.Sprintf("senderChainId:%d, voteIndex:%d", msg.ChainId, msg.VoteIndex),
		)
	}

	// get the tss address
	tss, err := k.relayerKeeper.GetTssAddress(ctx, &observertypes.QueryGetTssAddressRequest{
		BitcoinChainId: msg.ChainId,
	})
	if err != nil {
		return types.ErrCannotFindTSSKeys
	}

	// get the chain params
	params, found := k.relayerKeeper.GetChainParamsByChainID(ctx, msg.ChainId)
	if !found {
		return types.ErrUnsupportedChain
	}

	// call bridge pell method on contract
	evmTxResponse, contractCall, err := k.pevmKeeper.CallSwapOnPellGasSwap(
		ctx,
		msg.ChainId,
		params.GasTokenRechargeAmount.BigInt(),
		ethcommon.HexToAddress(tss.Eth),
	)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("processGasRecharge call pevm err: %s", err.Error()))
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("processGasRecharge call pevm success: %s, voteIndex: %d", evmTxResponse.String(), msg.VoteIndex))

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
	k.SetGasRechargeOperationIndex(ctx, msg.ChainId, msg.VoteIndex)
	return nil
}
