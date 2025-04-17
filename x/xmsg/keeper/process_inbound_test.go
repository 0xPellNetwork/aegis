package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestKeeper_ProcessInbound(t *testing.T) {
	t.Run("abort xmsg due to receiverChain is not pellchain", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)

		// Setup mock data
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()

		// call ProcessInbound
		xmsg := buildXmsg(t, receiver, *senderChain)
		k.ProcessInbound(ctx, xmsg)
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Equal(t, fmt.Sprintf("invalid receiver chainID %d", xmsg.InboundTxParams.SenderChainId), xmsg.XmsgStatus.StatusMessage)
	})

	t.Run("abort xmsg due to non-pell-xmsg", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)

		// Setup mock data
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()

		// call ProcessInbound
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.InboundTxParams.InboundPellTx = nil
		xmsg.GetCurrentOutTxParam().ReceiverChainId = chains.PellPrivnetChain().Id
		require.False(t, xmsg.IsCrossChainPellTx())
		k.ProcessInbound(ctx, xmsg)
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Equal(t, fmt.Sprintf("invalid xmsg[%s]", xmsg.Index), xmsg.XmsgStatus.StatusMessage)
	})

	t.Run("unable to process HandleEVMEvents reverts fails at GetSupportedChainFromChainID", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock:     true,
			UseObserverMock: true,
		})

		// Setup mock data
		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()
		err := fmt.Errorf("invalid sender chain id 1337")

		// Setup expected calls
		// mock unsuccessful HandlePEVMEvents which reverts , i.e returns err and isContractReverted = true
		keepertest.MockRevertForHandleEVMEvents_pell(pevmMock, senderChain.Id, err)

		// call ProcessInbound
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().ReceiverChainId = chains.PellPrivnetChain().Id
		k.ProcessInbound(ctx, xmsg)
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Equal(t, fmt.Sprintf("invalid sender chain id %d", xmsg.InboundTxParams.SenderChainId), xmsg.XmsgStatus.StatusMessage)
	})

	t.Run("unable to process HandleEVMEvents revert fails at PayGasInERC20AndUpdateXmsg", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock:     true,
			UseObserverMock: true,
		})

		// Setup mock data
		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()
		err := fmt.Errorf("processInbound revert message: processInbound revert message")

		// Setup expected calls
		keepertest.MockRevertForHandleEVMEvents_pell(pevmMock, senderChain.Id, err)

		// call ProcessInbound
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().ReceiverChainId = chains.PellPrivnetChain().Id
		k.ProcessInbound(ctx, xmsg)
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Equal(t, fmt.Sprint("processInbound revert message: processInbound revert message"), xmsg.XmsgStatus.StatusMessage)
	})

	t.Run("unable to process HandleEVMEvents reverts fails at UpdateNonce", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock:     true,
			UseObserverMock: true,
		})

		// Setup mock data
		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()
		err := fmt.Errorf("stakerdelegated failed\" does not contain \"cannot find receiver chain nonce")

		keepertest.MockRevertForHandleEVMEvents_pell(pevmMock, senderChain.Id, err)

		// call ProcessInbound
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().ReceiverChainId = chains.PellPrivnetChain().Id
		k.ProcessInbound(ctx, xmsg)
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Contains(t, xmsg.XmsgStatus.StatusMessage, "cannot find receiver chain nonce")
	})

	t.Run("unable to process HandleEVMEvents revert successfully", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock:     true,
			UseObserverMock: true,
		})

		// Setup mock data
		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()
		err := fmt.Errorf("stakerdelegated failed")

		// Setup expected calls
		// mock unsuccessful HandleEVMDeposit which reverts , i.e returns err and isContractReverted = true
		keepertest.MockRevertForHandleEVMEvents_pell(pevmMock, senderChain.Id, err)

		// call ProcessInbound
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().ReceiverChainId = chains.PellPrivnetChain().Id
		k.ProcessInbound(ctx, xmsg)
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Equal(t, err.Error(), xmsg.XmsgStatus.StatusMessage)
	})

	t.Run("unable to process HandleEVMEvents revert fails as the xmsg has already been reverted", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock:     true,
			UseObserverMock: true,
		})

		// Setup mock data
		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()
		err := fmt.Errorf("stakerdelegated failed\" does not contain \"revert outbound error: cannot revert a revert tx")

		// Setup expected calls
		// mock unsuccessful HandleEVMDeposit which reverts , i.e returns err and isContractReverted = true
		keepertest.MockRevertForHandleEVMEvents_pell(pevmMock, senderChain.Id, err)

		// call ProcessInbound
		xmsg := buildXmsg(t, receiver, *senderChain)
		xmsg.GetCurrentOutTxParam().ReceiverChainId = chains.PellPrivnetChain().Id
		xmsg.OutboundTxParams = append(xmsg.OutboundTxParams, xmsg.GetCurrentOutTxParam())
		k.ProcessInbound(ctx, xmsg)
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Contains(t, xmsg.XmsgStatus.StatusMessage, fmt.Sprintf("revert outbound error: %s", "cannot revert a revert tx"))
	})
}
