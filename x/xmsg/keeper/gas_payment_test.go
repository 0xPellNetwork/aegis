package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestKeeper_PayGasInPellAndUpdateXmsg(t *testing.T) {
	t.Run("can pay gas in pell", func(t *testing.T) {
		k, ctx, _, zk := testkeeper.XmsgKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)
		chainID := getValidEthChainID()
		setSupportedChain(ctx, zk, chainID)
		k.SetGasPrice(ctx, types.GasPrice{
			ChainId:     chainID,
			MedianIndex: 0,
			Prices:      []uint64{gasPrice},
		})

		// create a xmsg
		xmsg := types.Xmsg{
			InboundTxParams: &types.InboundTxParams{},
			OutboundTxParams: []*types.OutboundTxParams{
				{
					ReceiverChainId:    chainID,
					OutboundTxGasLimit: 1000,
				},
			},
		}

		err := k.PayNativeGasAndUpdateXmsg(ctx, chainID, &xmsg)
		require.NoError(t, err)
		require.Equal(t, "2", xmsg.GetCurrentOutTxParam().OutboundTxGasPrice)
	})

	t.Run("should fail if chain is not supported", func(t *testing.T) {
		k, ctx, _, _ := testkeeper.XmsgKeeper(t)
		xmsg := types.Xmsg{
			InboundTxParams: &types.InboundTxParams{},
		}
		err := k.PayNativeGasAndUpdateXmsg(ctx, 999999, &xmsg)
		require.ErrorIs(t, err, observertypes.ErrSupportedChains)
	})

	t.Run("should fail if can't query the gas price", func(t *testing.T) {
		k, ctx, _, zk := testkeeper.XmsgKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

		// deploy gas coin and set fee params
		chainID := getValidEthChainID()
		setSupportedChain(ctx, zk, chainID)

		// create a xmsg reverted from pell
		xmsg := types.Xmsg{
			InboundTxParams: &types.InboundTxParams{
				SenderChainId: chainID,
				Sender:        sample.EthAddress().String(),
			},
			OutboundTxParams: []*types.OutboundTxParams{
				{
					ReceiverChainId:    chainID,
					OutboundTxGasLimit: 1000,
				},
			},
		}

		err := k.PayNativeGasAndUpdateXmsg(ctx, chainID, &xmsg)
		require.Error(t, err)
	})
}
