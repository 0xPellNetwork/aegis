package keeper

import (
	"context"
	"errors"
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/cosmos"
	pellcrypto "github.com/0xPellNetwork/aegis/pkg/crypto"
	"github.com/0xPellNetwork/aegis/pkg/gas"
	"github.com/0xPellNetwork/aegis/relayer/tss"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// MigrateTssFunds migrates the funds from the current TSS to the new TSS
func (k msgServer) MigrateTssFunds(goCtx context.Context, msg *types.MsgMigrateTssFunds) (*types.MsgMigrateTssFundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if authorized
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, "Update can only be executed by the correct policy account")
	}

	if k.relayerKeeper.IsInboundEnabled(ctx) {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, "cannot migrate funds while inbound is enabled")
	}

	tss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, "cannot find current TSS")
	}

	tssHistory := k.relayerKeeper.GetAllTSS(ctx)
	if len(tssHistory) == 0 {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, "empty TSS history")
	}

	sort.SliceStable(tssHistory, func(i, j int) bool {
		return tssHistory[i].FinalizedPellHeight < tssHistory[j].FinalizedPellHeight
	})

	if tss.TssPubkey == tssHistory[len(tssHistory)-1].TssPubkey {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, "no new tss address has been generated")
	}

	// This check is to deal with an edge case where the current TSS is not part of the TSS history list at all
	if tss.FinalizedPellHeight >= tssHistory[len(tssHistory)-1].FinalizedPellHeight {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, "current tss is the latest")
	}

	pendingNonces, found := k.GetRelayerKeeper().GetPendingNonces(ctx, tss.TssPubkey, msg.ChainId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, "cannot find pending nonces for chain")
	}

	if pendingNonces.NonceLow != pendingNonces.NonceHigh {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, "cannot migrate funds when there are pending nonces")
	}

	err := k.MigrateTSSFundsForChain(ctx, msg.ChainId, msg.Amount, tss, tssHistory)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrCannotMigrateTssFunds, err.Error())
	}

	return &types.MsgMigrateTssFundsResponse{}, nil
}

func (k Keeper) MigrateTSSFundsForChain(ctx sdk.Context, chainID int64, amount sdkmath.Uint, currentTss observertypes.TSS, tssList []observertypes.TSS) error {
	// Always migrate to the latest TSS if multiple TSS addresses have been generated
	newTss := tssList[len(tssList)-1]
	medianGasPrice, isFound := k.GetMedianGasPriceInUint(ctx, chainID)
	if !isFound {
		return types.ErrUnableToGetGasPrice
	}

	currentTssAddr, err := tss.GetTssAddrEVM(currentTss.TssPubkey)
	if err != nil {
		return err
	}

	newTssAddr, err := tss.GetTssAddrEVM(newTss.TssPubkey)
	if err != nil {
		return err
	}

	pellSent := types.InboundPellEvent{
		PellData: &types.InboundPellEvent_PellSent{
			PellSent: &types.PellSent{
				TxOrigin:            "",
				Sender:              currentTssAddr.String(),
				ReceiverChainId:     chainID,
				Receiver:            newTssAddr.String(),
				Message:             "",
				PellParams:          pevmtypes.Transfer.String(),
				PellValue:           amount,
				DestinationGasLimit: sdkmath.NewUint(pevmtypes.PellSentDefaultDestinationGasLimit),
			},
		},
	}

	chainParams, found := k.relayerKeeper.GetChainParamsByChainID(ctx, chainID)
	if !found {
		return errors.New("chain params not found")
	}

	senderChain, err := chains.PellChainFromChainID(ctx.ChainID())
	if err != nil {
		return errors.New("sender chain not found")
	}

	receiverChain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, chainID)
	if receiverChain == nil {
		return errors.New("receiver chain not found")
	}

	eventLog := &ethtypes.Log{}
	msg := types.NewMsgVoteOnObservedInboundTx(
		"",
		currentTssAddr.String(),
		senderChain.Id,
		"",
		newTssAddr.String(),
		chainID,
		eventLog.TxHash.String(),
		eventLog.BlockNumber,
		chainParams.GasLimit,
		eventLog.Index,
		pellSent,
	)

	xmsg, err := k.processMigrationXmsg(ctx, msg, receiverChain, currentTss.TssPubkey)
	if err != nil {
		return err
	}

	// Set the sender and receiver addresses for EVM chain
	if chains.IsEVMChain(chainID) {
		ethAddressOld, err := pellcrypto.GetTssAddrEVM(currentTss.TssPubkey)
		if err != nil {
			return err
		}
		ethAddressNew, err := pellcrypto.GetTssAddrEVM(newTss.TssPubkey)
		if err != nil {
			return err
		}
		xmsg.InboundTxParams.Sender = ethAddressOld.String()
		xmsg.GetCurrentOutTxParam().Receiver = ethAddressNew.String()
		// Tss migration is a send transaction, so the gas limit is set to 21000
		xmsg.GetCurrentOutTxParam().OutboundTxGasLimit = gas.EVMSend
		// Multiple current gas price with standard multiplier to add some buffer
		multipliedGasPrice, err := gas.MultiplyGasPrice(medianGasPrice, types.TssMigrationGasMultiplierEVM)
		if err != nil {
			return err
		}
		xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = multipliedGasPrice.String()

	}

	if xmsg.GetCurrentOutTxParam().Receiver == "" {
		return errorsmod.Wrap(types.ErrReceiverIsEmpty, fmt.Sprintf("chain %d is not supported", chainID))
	}

	// The migrate funds can be run again to update the migration xmsg index if the migration fails
	// This should be used after carefully calculating the amount again
	existingMigrationInfo, found := k.relayerKeeper.GetFundMigrator(ctx, chainID)
	if found {
		olderMigrationXmsg, found := k.GetXmsg(ctx, existingMigrationInfo.MigrationXmsgIndex)
		if !found {
			return errorsmod.Wrapf(types.ErrCannotFindXmsg, "cannot find existing migration xmsg but migration info is present for chainID %d , migrator info : %s", chainID, existingMigrationInfo.String())
		}
		if olderMigrationXmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_OUTBOUND {
			return errorsmod.Wrapf(types.ErrUnsupportedStatus, "cannot migrate funds while there are pending migrations , migrator info :  %s", existingMigrationInfo.String())
		}
	}

	currCompressedPubkey, err := cosmos.GetPubKeyFromBech32(cosmos.Bech32PubKeyTypeAccPub, currentTss.TssPubkey)
	if err != nil {
		return err
	}

	newCompressedPubkey, err := cosmos.GetPubKeyFromBech32(cosmos.Bech32PubKeyTypeAccPub, newTss.TssPubkey)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info("Migrating TSS funds",
		"chainID", chainID,
		"amount", amount.String(),
		"currentTssPubkey", currentTss.TssPubkey,
		"newTssPubkey", newTss.TssPubkey,
		"currentTssAddr", currentTssAddr.String(),
		"newTssAddr", newTssAddr.String(),
		"currentTssCompressedPubkey", currCompressedPubkey,
		"newTssCompressedPubkey", newCompressedPubkey,
		"xmsgIndex", xmsg.Index)

	k.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, *xmsg)
	k.relayerKeeper.SetFundMigrator(ctx, observertypes.TssFundMigratorInfo{
		ChainId:            chainID,
		MigrationXmsgIndex: xmsg.Index,
	})
	EmitEventInboundFinalized(ctx, xmsg)

	return nil
}

func GetIndexStringForTssMigration(currentTssPubkey, newTssPubkey string, chainID int64, amount sdkmath.Uint, height int64) string {
	return fmt.Sprintf("%s-%s-%d-%s-%d", currentTssPubkey, newTssPubkey, chainID, amount.String(), height)
}

func (k *Keeper) processMigrationXmsg(ctx sdk.Context, msg *types.MsgVoteOnObservedInboundTx, receiverChain *chains.Chain, oldTssPybkey string) (*types.Xmsg, error) {
	// Create a new xmsg with status as pending Inbound, this is created directly from the event without waiting for any observer votes
	xmsg, err := types.NewXmsg(ctx, *msg, oldTssPybkey)
	if err != nil {
		return nil, err
	}

	xmsg.SetPendingOutbound("Cross-chain message ready for outbound processing from tss funds migration")

	// Get gas price and amount
	gasPrice, found := k.GetGasPrice(ctx, receiverChain.Id)
	if !found {
		return nil, errors.New("gas price not found")
	}

	xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = fmt.Sprint(gasPrice.Prices[gasPrice.MedianIndex])
	EmitEventPellSent(ctx, xmsg)

	if err := k.ProcessXmsg(ctx, xmsg, receiverChain); err != nil {
		return nil, err
	}

	return &xmsg, nil
}
