package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_VoteOnObservedOutboundTx(t *testing.T) {
	t.Run("successfully vote on outbound tx with status pending outbound ,vote-type success", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		// Setup mock data
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		receiver := sample.EthAddress()
		senderChain := getValidPellChain()
		observer := sample.AccAddress()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetObserverSet(ctx, relayertypes.RelayerSet{RelayerList: []string{observer}})
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().TssPubkey = tss.TssPubkey
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, *xmsg)
		observerMock.On("GetTSS", ctx).Return(relayertypes.TSS{}, true).Once()

		// Successfully mock VoteOnOutboundBallot
		keepertest.MockVoteOnOutboundSuccessBallot_pell(observerMock, ctx, xmsg, *senderChain, observer)

		// Successfully mock GetOutBound
		keepertest.MockGetOutBound_pell(observerMock, ctx)

		// Successfully mock SaveSuccessfulOutbound
		keepertest.MockSaveOutBound_pell(observerMock, ctx, xmsg, tss)

		msgServer := keeper.NewMsgServerImpl(*k)
		_, err := msgServer.VoteOnObservedOutboundTx(ctx, &types.MsgVoteOnObservedOutboundTx{
			XmsgHash:                       xmsg.Index,
			OutTxTssNonce:                  xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			OutTxChain:                     xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Status:                         chains.ReceiveStatus_SUCCESS,
			Signer:                         observer,
			ObservedOutTxHash:              sample.Hash().String(),
			ObservedOutTxBlockHeight:       10,
			ObservedOutTxEffectiveGasPrice: math.NewInt(21),
			ObservedOutTxGasUsed:           21,
		})
		require.NoError(t, err)
		c, found := k.GetXmsg(ctx, xmsg.Index)
		require.True(t, found)
		require.Equal(t, types.XmsgStatus_OUTBOUND_MINED, c.XmsgStatus.Status)
	})

	t.Run("successfully vote on outbound tx, vote-type failed", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
			UsePevmMock:     true,
		})

		// Setup mock data
		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		receiver := sample.EthAddress()
		senderChain := getValidPellChain()
		observer := sample.AccAddress()
		tss := sample.Tss_pell()

		zk.ObserverKeeper.SetObserverSet(ctx, relayertypes.RelayerSet{RelayerList: []string{observer}})
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().TssPubkey = tss.TssPubkey
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_PellSent{
				PellSent: &types.PellSent{
					TxOrigin:        "",
					Sender:          "",
					ReceiverChainId: 0,
					Receiver:        "",
					Message:         "",
					PellParams:      pevmtypes.RevertableCall.String(),
				},
			},
		}
		k.SetXmsg(ctx, *xmsg)

		observerMock.On("GetTSS", ctx).Return(relayertypes.TSS{}, true).Once()
		observerMock.On("RemoveFromPendingNonces", ctx, tss.TssPubkey, xmsg.GetCurrentOutTxParam().ReceiverChainId, mock.Anything).Return().Once()

		// Successfully mock VoteOnOutboundBallot
		keepertest.MockVoteOnOutboundFailedBallot_pell(observerMock, ctx, xmsg, *senderChain, observer)

		// Successfully mock GetOutBound
		keepertest.MockGetOutBound_pell(observerMock, ctx)

		// Successfully mock ProcessOutbound
		keepertest.MockProcessFailedOutboundForPEVMTx_pell(pevmMock, ctx, xmsg)

		//Successfully mock SaveOutBound
		keepertest.MockSaveOutBound_pell(observerMock, ctx, xmsg, tss)

		oldParamsLen := len(xmsg.OutboundTxParams)
		msgServer := keeper.NewMsgServerImpl(*k)
		_, err := msgServer.VoteOnObservedOutboundTx(ctx, &types.MsgVoteOnObservedOutboundTx{
			XmsgHash:                       xmsg.Index,
			OutTxTssNonce:                  xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			OutTxChain:                     xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Status:                         chains.ReceiveStatus_FAILED,
			Signer:                         observer,
			ObservedOutTxHash:              sample.Hash().String(),
			ObservedOutTxBlockHeight:       10,
			ObservedOutTxEffectiveGasPrice: math.NewInt(21),
			ObservedOutTxGasUsed:           21,
		})
		require.NoError(t, err)

		c, found := k.GetXmsg(ctx, xmsg.Index)
		require.True(t, found)
		require.Equal(t, types.XmsgStatus_REVERTED, c.XmsgStatus.Status)
		require.Equal(t, oldParamsLen+1, len(c.OutboundTxParams))
		require.Equal(t, types.TxFinalizationStatus_EXECUTED, c.OutboundTxParams[oldParamsLen-1].TxFinalizationStatus)
		require.Equal(t, types.TxFinalizationStatus_NOT_FINALIZED, xmsg.GetCurrentOutTxParam().TxFinalizationStatus)
	})

	t.Run("fail to finalize outbound if not a finalizing vote", func(t *testing.T) {
		k, ctx, sk, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{})

		// Setup mock data
		receiver := sample.EthAddress()
		senderChain := getValidPellChain()
		r := rand.New(rand.NewSource(42))
		validator := sample.Validator(t, r)
		tss := sample.Tss_pell()

		// set state to successfully vote on outbound tx
		accAddress, err := relayertypes.GetAccAddressFromOperatorAddress(validator.OperatorAddress)
		require.NoError(t, err)
		zk.ObserverKeeper.SetObserverSet(ctx, relayertypes.RelayerSet{RelayerList: []string{accAddress.String(), sample.AccAddress(), sample.AccAddress()}})
		sk.StakingKeeper.SetValidator(ctx, validator)
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().ReceiverChainId = getValidEthChain().Id
		xmsg.GetCurrentOutTxParam().TssPubkey = tss.TssPubkey
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, *xmsg)
		zk.ObserverKeeper.SetTSS(ctx, tss)

		msgServer := keeper.NewMsgServerImpl(*k)
		msg := &types.MsgVoteOnObservedOutboundTx{
			XmsgHash:                       xmsg.Index,
			OutTxTssNonce:                  xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			OutTxChain:                     xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Status:                         chains.ReceiveStatus_SUCCESS,
			Signer:                         accAddress.String(),
			ObservedOutTxHash:              sample.Hash().String(),
			ObservedOutTxBlockHeight:       10,
			ObservedOutTxEffectiveGasPrice: math.NewInt(21),
			ObservedOutTxGasUsed:           21,
		}
		_, err = msgServer.VoteOnObservedOutboundTx(ctx, msg)
		require.NoError(t, err)
		c, found := k.GetXmsg(ctx, xmsg.Index)
		require.True(t, found)
		require.Equal(t, types.XmsgStatus_PENDING_OUTBOUND, c.XmsgStatus.Status)
		ballot, found := zk.ObserverKeeper.GetBallot(ctx, msg.Digest())
		require.True(t, found)
		require.True(t, ballot.HasVoted(accAddress.String()))
	})

	t.Run("unable to add vote if tss is not present", func(t *testing.T) {
		k, ctx, sk, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{})

		// Setup mock data
		receiver := sample.EthAddress()
		senderChain := getValidPellChain()
		r := rand.New(rand.NewSource(42))
		validator := sample.Validator(t, r)
		tss := sample.Tss_pell()

		// set state to successfully vote on outbound tx
		accAddress, err := relayertypes.GetAccAddressFromOperatorAddress(validator.OperatorAddress)
		require.NoError(t, err)
		zk.ObserverKeeper.SetObserverSet(ctx, relayertypes.RelayerSet{RelayerList: []string{accAddress.String()}})
		sk.StakingKeeper.SetValidator(ctx, validator)
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().TssPubkey = tss.TssPubkey
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, *xmsg)

		msgServer := keeper.NewMsgServerImpl(*k)
		msg := &types.MsgVoteOnObservedOutboundTx{
			XmsgHash:                       xmsg.Index,
			OutTxTssNonce:                  xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			OutTxChain:                     xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Status:                         chains.ReceiveStatus_SUCCESS,
			Signer:                         accAddress.String(),
			ObservedOutTxHash:              sample.Hash().String(),
			ObservedOutTxBlockHeight:       10,
			ObservedOutTxEffectiveGasPrice: math.NewInt(21),
			ObservedOutTxGasUsed:           21,
		}
		_, err = msgServer.VoteOnObservedOutboundTx(ctx, msg)
		require.ErrorIs(t, err, types.ErrCannotFindTSSKeys)
		c, found := k.GetXmsg(ctx, xmsg.Index)
		require.True(t, found)
		require.Equal(t, types.XmsgStatus_PENDING_OUTBOUND, c.XmsgStatus.Status)
		_, found = zk.ObserverKeeper.GetBallot(ctx, msg.Digest())
		require.False(t, found)
	})
}

func TestKeeper_SaveFailedOutBound(t *testing.T) {
	t.Run("successfully save failed outbound", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		k.SetOutTxTracker(ctx, types.OutTxTracker{
			Index:     "",
			ChainId:   xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Nonce:     xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			HashLists: nil,
		})
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SaveFailedOutbound(ctx, xmsg, sample.String(), sample.PellIndex(t))
		require.Equal(t, xmsg.XmsgStatus.Status, types.XmsgStatus_ABORTED)
		_, found := k.GetOutTxTracker(ctx, xmsg.GetCurrentOutTxParam().ReceiverChainId, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce)
		require.False(t, found)
	})
}

func TestKeeper_SaveSuccessfulOutBound(t *testing.T) {
	t.Run("successfully save successful outbound", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		k.SetOutTxTracker(ctx, types.OutTxTracker{
			Index:     "",
			ChainId:   xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Nonce:     xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			HashLists: nil,
		})
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SaveSuccessfulOutbound(ctx, xmsg, sample.String())
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxBallotIndex, sample.String())
		_, found := k.GetOutTxTracker(ctx, xmsg.GetCurrentOutTxParam().ReceiverChainId, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce)
		require.False(t, found)
	})
}

func TestKeeper_SaveOutbound(t *testing.T) {
	t.Run("successfully save outbound", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)

		// setup state for xmsg and observer modules
		xmsg := sample.Xmsg_pell(t, "test")
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		ballotIndex := sample.String()
		k.SetOutTxTracker(ctx, types.OutTxTracker{
			Index:     "",
			ChainId:   xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Nonce:     xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			HashLists: nil,
		})

		zk.ObserverKeeper.SetPendingNonces(ctx, relayertypes.PendingNonces{
			NonceLow:  int64(xmsg.GetCurrentOutTxParam().OutboundTxTssNonce) - 1,
			NonceHigh: int64(xmsg.GetCurrentOutTxParam().OutboundTxTssNonce) + 1,
			ChainId:   xmsg.GetCurrentOutTxParam().ReceiverChainId,
			Tss:       xmsg.GetCurrentOutTxParam().TssPubkey,
		})
		zk.ObserverKeeper.SetTSS(ctx, relayertypes.TSS{
			TssPubkey: xmsg.GetCurrentOutTxParam().TssPubkey,
		})

		// Save outbound and assert all values are successfully saved
		k.SaveOutbound(ctx, xmsg, ballotIndex)
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxBallotIndex, ballotIndex)
		_, found := k.GetOutTxTracker(ctx, xmsg.GetCurrentOutTxParam().ReceiverChainId, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce)
		require.False(t, found)
		pn, found := zk.ObserverKeeper.GetPendingNonces(ctx, xmsg.GetCurrentOutTxParam().TssPubkey, xmsg.GetCurrentOutTxParam().ReceiverChainId)
		require.True(t, found)
		require.Equal(t, pn.NonceLow, int64(xmsg.GetCurrentOutTxParam().OutboundTxTssNonce)+1)
		require.Equal(t, pn.NonceHigh, int64(xmsg.GetCurrentOutTxParam().OutboundTxTssNonce)+1)
		_, found = k.GetInTxHashToXmsg(ctx, xmsg.InboundTxParams.InboundTxHash)
		require.True(t, found)
		_, found = zk.ObserverKeeper.GetNonceToXmsg(ctx, xmsg.GetCurrentOutTxParam().TssPubkey, xmsg.GetCurrentOutTxParam().ReceiverChainId, int64(xmsg.GetCurrentOutTxParam().OutboundTxTssNonce))
		require.True(t, found)
	})
}

func TestKeeper_ValidateOutboundMessage(t *testing.T) {
	t.Run("successfully validate outbound message", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		k.SetXmsg(ctx, *xmsg)
		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
		_, err := k.ValidateOutboundMessage(ctx, types.MsgVoteOnObservedOutboundTx{
			XmsgHash:      xmsg.Index,
			OutTxTssNonce: xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			OutTxChain:    xmsg.GetCurrentOutTxParam().ReceiverChainId,
		})
		require.NoError(t, err)
	})

	t.Run("failed to validate outbound message if xmsg not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		msg := types.MsgVoteOnObservedOutboundTx{
			XmsgHash:      sample.String(),
			OutTxTssNonce: 1,
		}
		_, err := k.ValidateOutboundMessage(ctx, msg)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		require.ErrorContains(t, err, fmt.Sprintf("Xmsg %s does not exist", msg.XmsgHash))
	})

	t.Run("failed to validate outbound message if nonce does not match", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		k.SetXmsg(ctx, *xmsg)
		msg := types.MsgVoteOnObservedOutboundTx{
			XmsgHash:      xmsg.Index,
			OutTxTssNonce: 2,
		}
		_, err := k.ValidateOutboundMessage(ctx, msg)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		require.ErrorContains(t, err, fmt.Sprintf("OutTxTssNonce %d does not match Xmsg OutTxTssNonce %d", msg.OutTxTssNonce, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce))
	})

	t.Run("failed to validate outbound message if tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		k.SetXmsg(ctx, *xmsg)
		_, err := k.ValidateOutboundMessage(ctx, types.MsgVoteOnObservedOutboundTx{
			XmsgHash:      xmsg.Index,
			OutTxTssNonce: xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
		})
		require.ErrorIs(t, err, types.ErrCannotFindTSSKeys)
	})

	t.Run("failed to validate outbound message if chain does not match", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		k.SetXmsg(ctx, *xmsg)
		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
		_, err := k.ValidateOutboundMessage(ctx, types.MsgVoteOnObservedOutboundTx{
			XmsgHash:      xmsg.Index,
			OutTxTssNonce: xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
			OutTxChain:    2,
		})
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		require.ErrorContains(t, err, fmt.Sprintf("OutTxChain %d does not match Xmsg OutTxChain %d", 2, xmsg.GetCurrentOutTxParam().ReceiverChainId))
	})
}
