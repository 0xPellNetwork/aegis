package keeper_test

import (
	"testing"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_ProcessSuccessfulOutbound(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	xmsg := sample.Xmsg_pell(t, "test")
	// transition to reverted if pending revert
	xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_REVERT
	k.ProcessSuccessfulOutbound(ctx, xmsg)
	require.Equal(t, xmsg.XmsgStatus.Status, types.XmsgStatus_REVERTED)
	// transition to outbound mined if pending outbound
	xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
	k.ProcessSuccessfulOutbound(ctx, xmsg)
	require.Equal(t, xmsg.XmsgStatus.Status, types.XmsgStatus_OUTBOUND_MINED)
	// do nothing if it's in any other state
	k.ProcessSuccessfulOutbound(ctx, xmsg)
	require.Equal(t, xmsg.XmsgStatus.Status, types.XmsgStatus_OUTBOUND_MINED)
}

func TestKeeper_ProcessFailedOutbound(t *testing.T) {
	t.Run("successfully process failed outbound set to aborted for non-pell-chain", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		xmsg.InboundTxParams.SenderChainId = getValidEthChainID()
		err := k.ProcessFailedOutbound(ctx, xmsg)
		require.NoError(t, err)
		require.Equal(t, xmsg.XmsgStatus.Status, types.XmsgStatus_ABORTED)
		require.Equal(t, xmsg.GetCurrentOutTxParam().TxFinalizationStatus, types.TxFinalizationStatus_EXECUTED)
	})

	t.Run("unable to  process failed outbound if GetXmsgIndicesBytes fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		receiver := sample.EthAddress()
		xmsg := buildXmsg(t, receiver, chains.GoerliChain())
		xmsg.Index = ""
		xmsg.InboundTxParams.SenderChainId = chains.PellChainMainnet().Id
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_PellSent{
				PellSent: &types.PellSent{
					TxOrigin:            "",
					Sender:              "",
					ReceiverChainId:     0,
					Receiver:            "",
					Message:             "",
					PellParams:          pevmtypes.RevertableCall.String(),
					PellValue:           sdkmath.Uint{},
					DestinationGasLimit: sdkmath.Uint{},
				},
			},
		}
		err := k.ProcessFailedOutbound(ctx, xmsg)
		require.ErrorContains(t, err, "failed reverting GetXmsgIndicesBytes")
	})

	t.Run("unable to  process failed outbound if Adding Revert fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := sample.Xmsg_pell(t, "test")
		xmsg.InboundTxParams.SenderChainId = chains.PellChainMainnet().Id
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_PellSent{
				PellSent: &types.PellSent{
					TxOrigin:            "",
					Sender:              "",
					ReceiverChainId:     0,
					Receiver:            "",
					Message:             "",
					PellParams:          pevmtypes.RevertableCall.String(),
					PellValue:           sdkmath.Uint{},
					DestinationGasLimit: sdkmath.Uint{},
				},
			},
		}
		err := k.ProcessFailedOutbound(ctx, xmsg)
		require.ErrorContains(t, err, "failed AddRevertOutbound")
	})

	t.Run("successfully process failed outbound if original sender is a address", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.XmsgKeeper(t)
		receiver := sample.EthAddress()
		xmsg := buildXmsg(t, receiver, chains.GoerliChain())
		err := sdkk.EvmKeeper.SetAccount(ctx, ethcommon.HexToAddress(xmsg.InboundTxParams.Sender), *statedb.NewEmptyAccount())
		require.NoError(t, err)
		xmsg.InboundTxParams.SenderChainId = chains.PellChainMainnet().Id
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_PellSent{
				PellSent: &types.PellSent{
					TxOrigin:            "",
					Sender:              "",
					ReceiverChainId:     0,
					Receiver:            "",
					Message:             "",
					PellParams:          pevmtypes.RevertableCall.String(),
					PellValue:           sdkmath.Uint{},
					DestinationGasLimit: sdkmath.Uint{},
				},
			},
		}
		err = k.ProcessFailedOutbound(ctx, xmsg)
		require.NoError(t, err)
		require.Equal(t, types.XmsgStatus_REVERTED, xmsg.XmsgStatus.Status)
		require.Equal(t, xmsg.GetCurrentOutTxParam().TxFinalizationStatus, types.TxFinalizationStatus_EXECUTED)
	})

	t.Run("unable to process failed outbound if PELLRevertAndCallContract fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})
		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		receiver := sample.EthAddress()
		errorFailedPELLRevertAndCallContract := errors.New("test", 999, "failed PELLRevertAndCallContract")
		xmsg := buildXmsg(t, receiver, chains.GoerliChain())
		xmsg.InboundTxParams.SenderChainId = chains.PellChainMainnet().Id
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_PellSent{
				PellSent: &types.PellSent{
					TxOrigin:            "",
					Sender:              "",
					ReceiverChainId:     0,
					Receiver:            "",
					Message:             "",
					PellParams:          pevmtypes.RevertableCall.String(),
					PellValue:           sdkmath.Uint{},
					DestinationGasLimit: sdkmath.Uint{},
				},
			},
		}
		pevmMock.On("PELLRevertAndCallContract", mock.Anything,
			ethcommon.HexToAddress(xmsg.InboundTxParams.Sender),
			ethcommon.HexToAddress(xmsg.GetCurrentOutTxParam().Receiver),
			xmsg.InboundTxParams.SenderChainId,
			xmsg.GetCurrentOutTxParam().ReceiverChainId,
			mock.Anything,
			mock.Anything).Return(nil, errorFailedPELLRevertAndCallContract).Once()
		err := k.ProcessFailedOutbound(ctx, xmsg)
		require.ErrorContains(t, err, "failed PELLRevertAndCallContract")
	})
}

func TestKeeper_ProcessOutbound(t *testing.T) {
	t.Run("successfully process outbound with ballot finalized to success", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := buildXmsg(t, sample.EthAddress(), chains.GoerliChain())
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		err := k.ProcessOutbound(ctx, xmsg, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)
		require.NoError(t, err)
		require.Equal(t, xmsg.XmsgStatus.Status, types.XmsgStatus_OUTBOUND_MINED)
	})

	t.Run("successfully process outbound with ballot finalized to failed and old status is Pending Revert", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := buildXmsg(t, sample.EthAddress(), chains.GoerliChain())
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_REVERT
		err := k.ProcessOutbound(ctx, xmsg, relayertypes.BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION)
		require.NoError(t, err)
		require.Equal(t, xmsg.XmsgStatus.Status, types.XmsgStatus_ABORTED)
		require.Equal(t, xmsg.GetCurrentOutTxParam().TxFinalizationStatus, types.TxFinalizationStatus_EXECUTED)
	})

	t.Run("do not process if xmsg invalid", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		xmsg := buildXmsg(t, sample.EthAddress(), chains.GoerliChain())
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		xmsg.InboundTxParams = nil
		err := k.ProcessOutbound(ctx, xmsg, relayertypes.BallotStatus_BALLOT_IN_PROGRESS)
		require.Error(t, err)
	})
}
