package types_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestXmsg_GetXmsgIndicesBytes(t *testing.T) {
	xmsg := sample.Xmsg_pell(t, "sample")
	indexBytes, err := xmsg.GetXmsgIndicesBytes()
	require.NoError(t, err)
	require.Equal(t, xmsg.Index, types.GetXmsgIndicesFromBytes(indexBytes))
}

func Test_InitializeXmsg(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	t.Run("should return a xmsg with correct values", func(t *testing.T) {
		_, ctx, _, _ := keepertest.XmsgKeeper(t)
		senderChain := chains.GoerliChain()
		sender := sample.EthAddress()
		receiverChain := chains.GoerliChain()
		receiver := sample.EthAddress()
		creator := sample.AccAddress()
		intxBlockHeight := uint64(420)
		intxHash := sample.Hash()
		gasLimit := uint64(100)
		eventIndex := uint64(1)
		tss := sample.Tss_pell()
		pellTx := sample.InboundPellTx_StakerDeposited_pell(r)
		msg := types.MsgVoteOnObservedInboundTx{
			Signer:        creator,
			Sender:        sender.String(),
			SenderChainId: senderChain.Id,
			Receiver:      receiver.String(),
			ReceiverChain: receiverChain.Id,
			InTxHash:      intxHash.String(),
			InBlockHeight: intxBlockHeight,
			GasLimit:      gasLimit,
			TxOrigin:      sender.String(),
			EventIndex:    eventIndex,
			PellTx:        pellTx,
		}
		xmsg, err := types.NewXmsg(ctx, msg, tss.TssPubkey)
		require.NoError(t, err)
		require.Equal(t, receiver.String(), xmsg.GetCurrentOutTxParam().Receiver)
		require.Equal(t, receiverChain.Id, xmsg.GetCurrentOutTxParam().ReceiverChainId)
		require.Equal(t, sender.String(), xmsg.GetInboundTxParams().Sender)
		require.Equal(t, senderChain.Id, xmsg.GetInboundTxParams().SenderChainId)
		require.Equal(t, intxHash.String(), xmsg.GetInboundTxParams().InboundTxHash)
		require.Equal(t, intxBlockHeight, xmsg.GetInboundTxParams().InboundTxBlockHeight)
		require.Equal(t, gasLimit, xmsg.GetCurrentOutTxParam().OutboundTxGasLimit)
		require.Equal(t, uint64(0), xmsg.GetCurrentOutTxParam().OutboundTxTssNonce)
		require.Equal(t, types.XmsgStatus_PENDING_INBOUND, xmsg.XmsgStatus.Status)
		require.Equal(t, pellTx, xmsg.GetInboundTxParams().InboundPellTx)
	})
	t.Run("should return an error if the xmsg is invalid", func(t *testing.T) {
		_, ctx, _, _ := keepertest.XmsgKeeper(t)
		senderChain := chains.GoerliChain()
		sender := sample.EthAddress()
		receiverChain := chains.GoerliChain()
		receiver := sample.EthAddress()
		creator := sample.AccAddress()
		intxBlockHeight := uint64(420)
		intxHash := sample.Hash()
		gasLimit := uint64(100)
		eventIndex := uint64(1)
		tss := sample.Tss_pell()
		pellTx := sample.InboundPellTx_StakerDelegated_pell(r)
		msg := types.MsgVoteOnObservedInboundTx{
			Signer:        creator,
			Sender:        "",
			SenderChainId: senderChain.Id,
			Receiver:      receiver.String(),
			ReceiverChain: receiverChain.Id,
			InTxHash:      intxHash.String(),
			InBlockHeight: intxBlockHeight,
			GasLimit:      gasLimit,
			TxOrigin:      sender.String(),
			EventIndex:    eventIndex,
			PellTx:        pellTx,
		}
		_, err := types.NewXmsg(ctx, msg, tss.TssPubkey)
		require.ErrorContains(t, err, "sender cannot be empty")
	})
}

func TestXmsg_Validate(t *testing.T) {
	xmsg := sample.Xmsg_pell(t, "foo")
	xmsg.InboundTxParams = nil
	require.ErrorContains(t, xmsg.Validate(), "inbound tx params cannot be nil")
	xmsg = sample.Xmsg_pell(t, "foo")
	xmsg.OutboundTxParams = nil
	require.ErrorContains(t, xmsg.Validate(), "outbound tx params cannot be nil")
	xmsg = sample.Xmsg_pell(t, "foo")
	xmsg.XmsgStatus = nil
	require.ErrorContains(t, xmsg.Validate(), "xmsg status cannot be nil")
	xmsg = sample.Xmsg_pell(t, "foo")
	xmsg.OutboundTxParams = make([]*types.OutboundTxParams, 3)
	require.ErrorContains(t, xmsg.Validate(), "outbound tx params cannot be more than 2")
	xmsg = sample.Xmsg_pell(t, "foo")
	xmsg.Index = "0"
	require.ErrorContains(t, xmsg.Validate(), "invalid index length 1")
	xmsg = sample.Xmsg_pell(t, "foo")
	xmsg.InboundTxParams = sample.InboundTxParamsValidChainID_pell(rand.New(rand.NewSource(42)))
	xmsg.InboundTxParams.SenderChainId = 1000
	require.ErrorContains(t, xmsg.Validate(), "invalid sender chain id 1000")
	xmsg = sample.Xmsg_pell(t, "foo")
	xmsg.OutboundTxParams = []*types.OutboundTxParams{sample.OutboundTxParamsValidChainID_pell(rand.New(rand.NewSource(42)))}
	xmsg.InboundTxParams = sample.InboundTxParamsValidChainID_pell(rand.New(rand.NewSource(42)))
	xmsg.InboundTxParams.InboundTxHash = sample.Hash().String()
	xmsg.InboundTxParams.InboundTxBallotIndex = sample.PellIndex(t)
	xmsg.OutboundTxParams[0].ReceiverChainId = 1000
	require.ErrorContains(t, xmsg.Validate(), "invalid receiver chain id 1000")
}

func TestXmsg_GetCurrentOutTxParam(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	xmsg := sample.Xmsg_pell(t, "foo")

	xmsg.OutboundTxParams = []*types.OutboundTxParams{}
	require.Equal(t, &types.OutboundTxParams{}, xmsg.GetCurrentOutTxParam())

	xmsg.OutboundTxParams = []*types.OutboundTxParams{sample.OutboundTxParams_pell(r)}
	require.Equal(t, xmsg.OutboundTxParams[0], xmsg.GetCurrentOutTxParam())

	xmsg.OutboundTxParams = []*types.OutboundTxParams{sample.OutboundTxParams_pell(r), sample.OutboundTxParams_pell(r)}
	require.Equal(t, xmsg.OutboundTxParams[1], xmsg.GetCurrentOutTxParam())
}

func TestXmsg_IsCurrentOutTxRevert(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	xmsg := sample.Xmsg_pell(t, "foo")

	xmsg.OutboundTxParams = []*types.OutboundTxParams{}
	require.False(t, xmsg.IsCurrentOutTxRevert())

	xmsg.OutboundTxParams = []*types.OutboundTxParams{sample.OutboundTxParams_pell(r)}
	require.False(t, xmsg.IsCurrentOutTxRevert())

	xmsg.OutboundTxParams = []*types.OutboundTxParams{sample.OutboundTxParams_pell(r), sample.OutboundTxParams_pell(r)}
	require.True(t, xmsg.IsCurrentOutTxRevert())
}

func TestXmsg_OriginalDestinationChainID(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	xmsg := sample.Xmsg_pell(t, "foo")

	xmsg.OutboundTxParams = []*types.OutboundTxParams{}
	require.Equal(t, int64(-1), xmsg.OriginalDestinationChainID())

	xmsg.OutboundTxParams = []*types.OutboundTxParams{sample.OutboundTxParams_pell(r)}
	require.Equal(t, xmsg.OutboundTxParams[0].ReceiverChainId, xmsg.OriginalDestinationChainID())

	xmsg.OutboundTxParams = []*types.OutboundTxParams{sample.OutboundTxParams_pell(r), sample.OutboundTxParams_pell(r)}
	require.Equal(t, xmsg.OutboundTxParams[0].ReceiverChainId, xmsg.OriginalDestinationChainID())
}

func TestXmsg_AddOutbound(t *testing.T) {
	t.Run("successfully get outbound tx", func(t *testing.T) {
		_, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		hash := sample.Hash().String()

		err := xmsg.AddOutbound(ctx, types.MsgVoteOnObservedOutboundTx{
			ObservedOutTxHash:              hash,
			ObservedOutTxBlockHeight:       10,
			ObservedOutTxGasUsed:           100,
			ObservedOutTxEffectiveGasPrice: sdkmath.NewInt(100),
			ObservedOutTxEffectiveGasLimit: 20,
		}, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)
		require.NoError(t, err)
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxHash, hash)
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxGasUsed, uint64(100))
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxEffectiveGasPrice, sdkmath.NewInt(100))
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxEffectiveGasLimit, uint64(20))
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxExternalHeight, uint64(10))
		require.Equal(t, xmsg.XmsgStatus.LastUpdateTimestamp, ctx.BlockHeader().Time.Unix())
	})

	t.Run("successfully get outbound tx for failed ballot without amount check", func(t *testing.T) {
		_, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		hash := sample.Hash().String()

		err := xmsg.AddOutbound(ctx, types.MsgVoteOnObservedOutboundTx{
			ObservedOutTxHash:              hash,
			ObservedOutTxBlockHeight:       10,
			ObservedOutTxGasUsed:           100,
			ObservedOutTxEffectiveGasPrice: sdkmath.NewInt(100),
			ObservedOutTxEffectiveGasLimit: 20,
		}, relayertypes.BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION)
		require.NoError(t, err)
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxHash, hash)
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxGasUsed, uint64(100))
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxEffectiveGasPrice, sdkmath.NewInt(100))
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxEffectiveGasLimit, uint64(20))
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxExternalHeight, uint64(10))
		require.Equal(t, xmsg.XmsgStatus.LastUpdateTimestamp, ctx.BlockHeader().Time.Unix())
	})
}

func Test_SetRevertOutboundValues(t *testing.T) {
	t.Run("successfully set revert outbound values", func(t *testing.T) {
		xmsg := sample.Xmsg_pell(t, "test")
		xmsg.OutboundTxParams = xmsg.OutboundTxParams[:1]
		err := xmsg.AddRevertOutbound(100)
		require.NoError(t, err)
		require.Len(t, xmsg.OutboundTxParams, 2)
		require.Equal(t, xmsg.GetCurrentOutTxParam().Receiver, xmsg.InboundTxParams.Sender)
		require.Equal(t, xmsg.GetCurrentOutTxParam().ReceiverChainId, xmsg.InboundTxParams.SenderChainId)
		require.Equal(t, xmsg.GetCurrentOutTxParam().OutboundTxGasLimit, uint64(100))
		require.Equal(t, xmsg.GetCurrentOutTxParam().TssPubkey, xmsg.OutboundTxParams[0].TssPubkey)
		require.Equal(t, types.TxFinalizationStatus_EXECUTED, xmsg.OutboundTxParams[0].TxFinalizationStatus)
	})

	t.Run("failed to set revert outbound values if revert outbound already exists", func(t *testing.T) {
		xmsg := sample.Xmsg_pell(t, "test")
		err := xmsg.AddRevertOutbound(100)
		require.ErrorContains(t, err, "cannot revert a revert tx")
	})

	t.Run("failed to set revert outbound values if revert outbound already exists", func(t *testing.T) {
		xmsg := sample.Xmsg_pell(t, "test")
		xmsg.OutboundTxParams = make([]*types.OutboundTxParams, 0)
		err := xmsg.AddRevertOutbound(100)
		require.ErrorContains(t, err, "cannot revert before trying to process an outbound tx")
	})
}

func TestXmsg_SetAbort(t *testing.T) {
	xmsg := sample.Xmsg_pell(t, "test")
	xmsg.SetAbort("test")
	require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
	require.Equal(t, "test", "test")
}

func TestXmsg_SetPendingRevert(t *testing.T) {
	xmsg := sample.Xmsg_pell(t, "test")
	xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
	xmsg.SetPendingRevert("test")
	require.Equal(t, types.XmsgStatus_PENDING_REVERT, xmsg.XmsgStatus.Status)
	require.Contains(t, xmsg.XmsgStatus.StatusMessage, "test")
}

func TestXmsg_SetPendingOutbound(t *testing.T) {
	xmsg := sample.Xmsg_pell(t, "test")
	xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_INBOUND
	xmsg.SetPendingOutbound("test")
	require.Equal(t, types.XmsgStatus_PENDING_OUTBOUND, xmsg.XmsgStatus.Status)
	require.Contains(t, xmsg.XmsgStatus.StatusMessage, "test")
}

func TestXmsg_SetOutBoundMined(t *testing.T) {
	xmsg := sample.Xmsg_pell(t, "test")
	xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
	xmsg.SetOutBoundMined("test")
	require.Equal(t, types.XmsgStatus_OUTBOUND_MINED, xmsg.XmsgStatus.Status)
	require.Contains(t, xmsg.XmsgStatus.StatusMessage, "test")
}

func TestXmsg_SetReverted(t *testing.T) {
	xmsg := sample.Xmsg_pell(t, "test")
	xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_REVERT
	xmsg.SetReverted("test")
	require.Equal(t, types.XmsgStatus_REVERTED, xmsg.XmsgStatus.Status)
	require.Contains(t, xmsg.XmsgStatus.StatusMessage, "test")
}
